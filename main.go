package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	logger "github.com/mmmorris1975/simple-logger"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	// Version is the program version, defined at build time
	Version string
	sb      SecretBackender

	log = logger.StdLogger

	// program args
	backendArg     string
	dynamoTableArg string
	bucketArg      string
	kmsKeyArg      string
	oneShotArg     bool
	verboseArg     bool
	versionArg     bool

	// AWS stuff
	cfg        = aws.NewConfig().WithLogger(log)
	ses        = session.Must(session.NewSession(cfg))
	dynamoSvc  = dynamodb.ServiceName
	ssmSvc     = ssm.ServiceName
	secretsSvc = secretsmanager.ServiceName
	s3Svc      = s3.ServiceName
	keyArn     arn.ARN

	backends = sort.StringSlice{dynamoSvc, ssmSvc, secretsSvc, s3Svc}
)

func init() {
	backends.Sort()

	flag.StringVar(&backendArg, "s", os.Getenv("SECRETS_BACKEND"),
		fmt.Sprintf("Secrets storage backend: %s", strings.Join(backends, ", ")))
	flag.StringVar(&dynamoTableArg, "t", os.Getenv("DYNAMODB_TABLE"),
		fmt.Sprintf("DynamoDB table name, required only for %s backend, ignored by all others", dynamoSvc))
	flag.StringVar(&bucketArg, "b", os.Getenv("S3_BUCKET"),
		fmt.Sprintf("S3 bucket name, required only for %s backend, ignored by all others", s3Svc))
	flag.StringVar(&kmsKeyArg, "k", os.Getenv("KMS_KEY"),
		fmt.Sprintf("KMS key ARN, ID, or alias (required for %s and %s backends, optional for %s backend, not used for %s backend)",
			dynamoSvc, s3Svc, ssmSvc, secretsSvc))
	flag.BoolVar(&oneShotArg, "o", checkBoolEnv("ONE_SHOT"), "run in one-shot mode, providing the key and value to store on the command line")
	flag.BoolVar(&verboseArg, "v", checkBoolEnv("VERBOSE"), "Print verbose output")
	flag.BoolVar(&versionArg, "V", false, "Print program version")
}

// SecretBackender is the interface type for conforming secrets backends
type SecretBackender interface {
	// KmsRequired returns true if the backend requires a KMS key argument for operation. Currently, only
	// the dynamodb backend sets this to true, since it is unable to perform transparent encryption of item values
	KmsRequired() bool

	// Store will set the supplied value in the backend as the provided key
	Store(string, interface{}) error
}

func main() {
	flag.Parse()

	if verboseArg {
		log.SetLevel(logger.DEBUG)
	}

	if versionArg {
		log.Printf("VERSION: %s", Version)
	}

	if err := validateBackend(); err != nil {
		log.Fatal(err)
	}

	if err := validateKey(); err != nil {
		log.Fatal(err)
	}

	errCnt := 0
	if oneShotArg {
		log.Debug("using one-shot mode")
		var v interface{}

		if len(flag.Args()) > 1 {
			v = flag.Arg(1)
		} else {
			v = os.Stdin
		}

		if err := oneShotHandler(flag.Arg(0), v); err != nil {
			log.Fatalf("error storing secret: %v", err)
		}
	} else {
		log.Debug("using json mode")
		var in interface{}

		if len(flag.Arg(0)) > 0 {
			in = flag.Arg(0)
		} else {
			in = os.Stdin
		}

		errCnt = jsonHandler(in)
	}

	os.Exit(errCnt)
}

func oneShotHandler(k string, v interface{}) error {
	if err := sb.Store(k, v); err != nil {
		return err
	}

	log.Infof("updated secret %s", k)
	return nil
}

func jsonHandler(in interface{}) int {
	var errs int

	r := getReader(in)
	if r == nil {
		log.Error("received a nil reader to handle json input, something has gone very wrong")
		errs++
		return errs
	}

	j := json.NewDecoder(r)
	for {
		m := make(map[string]interface{})
		if err := j.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}

			// bad json, should probably not continue
			log.Errorf("error decoding json: %v", err)
			errs++
			break
		}

		for k, v := range m {
			if err := sb.Store(k, v); err != nil {
				log.Errorf("error storing secret: %v", err)
				errs++
			} else {
				log.Infof("updated secret %s", k)
			}
		}
	}

	return errs
}

// truth-y values are 1, t, T, TRUE, true, True; everything else is false
func checkBoolEnv(v string) bool {
	log.Debugf("checkBoolEnv input: %s", v)
	b, err := strconv.ParseBool(os.Getenv(v))
	if err != nil {
		log.Debugf("ParseBool error: %v", err)
		b = false
	}
	log.Debugf("checkBoolEnv returning: %v", b)
	return b
}

// verify that we're called with a supported secrets backend
func validateBackend() error {
	backendLc := strings.ToLower(backendArg)
	i := backends.Search(backendLc)

	if i >= len(backends) || backends[i] != backendLc {
		return fmt.Errorf("backend %s is not valid, must be one of: %s", backendArg, strings.Join(backends, ", "))
	}

	return backendFactory(backends[i])
}

// KMS key is required, or a KMS key was explicitly passed with the ssm backend
func validateKey() error {
	if (sb != nil && sb.KmsRequired()) || (backendArg == ssm.ServiceName && len(kmsKeyArg) > 0) {
		c := kms.New(ses)
		i := kms.DescribeKeyInput{KeyId: aws.String(kmsKeyArg)}
		o, err := c.DescribeKey(&i)
		if err != nil {
			return fmt.Errorf("failed to lookup KMS key %s, error: %v", kmsKeyArg, err)
		}

		keyArn, err = arn.Parse(*o.KeyMetadata.Arn)
		if err != nil {
			return fmt.Errorf("bad key ARN: %v", err)
		}
	}
	return nil
}

func backendFactory(be string) error {
	switch be {
	case dynamoSvc:
		if len(dynamoTableArg) < 1 {
			return fmt.Errorf("missing required table name for %s backend", dynamoSvc)
		}

		var err error
		sb, err = NewDynamoDbBackend().WithTable(dynamoTableArg)

		if err != nil {
			return err
		}
	case secretsSvc:
		sb = NewSecretsManagerBackend()
	case ssmSvc:
		sb = NewParameterStoreBackend()
	case s3Svc:
		if len(bucketArg) < 1 {
			return fmt.Errorf("missing required bucket name for %s backend", s3Svc)
		}

		sb = NewS3Backend().WithBucket(bucketArg)
	default:
		return fmt.Errorf("unsupported backend %s", be)
	}

	log.Debugf("setting backend to %s", be)
	return nil
}

func readBinary(value interface{}) (io.Reader, error) {
	if value == nil {
		return nil, fmt.Errorf("nil value")
	}

	switch t := value.(type) {
	case io.Reader:
		return t, nil
	}

	b := bytes.NewBuffer(make([]byte, 0, 4096))
	if _, err := fmt.Fprint(b, value); err != nil {
		return nil, err
	}
	return b, nil
}

func getReader(data interface{}) io.Reader {
	var in io.ReadSeeker

	switch t := data.(type) {
	case string:
		in = strings.NewReader(t)

		// 4 bytes is the minimum length of a base64 encoded single character, so if the input is less than that
		// there's no way it can be base64 encoded and there's no need to continue further
		if len(t) < 4 {
			return in
		}
	case io.ReadSeeker:
		in = t
	case []byte:
		in = bytes.NewReader(t)
	default:
		// since this is an internal method only, we'll pretend this is ok
		// and add a case statement if a new use case pops up
		return nil
	}

	b64, err := checkBase64(in)
	if err != nil {
		log.Debugf("returning source reader, base64 error: %v", err)
		return in
	}

	gz, err := gzip.NewReader(b64)
	if err != nil {
		// I think this will raise an error if it's not a gzip compressed stream
		// in which case, just return the base64 reader
		if err != io.EOF {
			log.Debugf("returning base64 reader, gzip error: %v", err)
			in.Seek(0, io.SeekStart)
			return b64
		}
	}

	log.Debugf("returning gzip base64 reader")
	return gz
}

// if the input is text which is also a valid base64 string, this method will happily
// decode the value to bytes, just so ya know
func checkBase64(in io.ReadSeeker) (io.Reader, error) {
	defer in.Seek(0, io.SeekStart)

	buf := make([]byte, 4096)
	n, err := in.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	if _, err := base64.StdEncoding.Decode(make([]byte, n), buf[0:n]); err != nil {
		return nil, err
	}

	return base64.NewDecoder(base64.StdEncoding, in), nil
}
