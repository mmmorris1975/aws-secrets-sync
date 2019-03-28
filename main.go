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
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	logger "github.com/mmmorris1975/simple-logger"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	Version string

	log = logger.StdLogger

	// program args
	backendArg     string
	dynamoTableArg string
	kmsKeyArg      string
	verboseArg     bool
	versionArg     bool

	// AWS stuff
	cfg        = aws.NewConfig().WithLogger(log)
	ses        = session.Must(session.NewSession(cfg))
	keyArn     arn.ARN
	dynamoSvc  = dynamodb.ServiceName
	ssmSvc     = ssm.ServiceName
	secretsSvc = secretsmanager.ServiceName

	backends = sort.StringSlice{dynamoSvc, ssmSvc, secretsSvc}

	sb SecretBackender
)

// (option) env vars
// (-b) SECRETS_BACKEND = parameterstore | secretsmanager | dynamodb
// (-t) DYNAMODB_TABLE = required for dynamodb backend
// (-k) KMS_KEY = KMS key arn, id, or alias
// (-v) VERBOSE = verboseArg logging

// options
// -V print versionArg
func init() {
	backends.Sort()

	flag.StringVar(&backendArg, "b", os.Getenv("SECRETS_BACKEND"),
		fmt.Sprintf("Secrets storage backend: %s", strings.Join(backends, ", ")))
	flag.StringVar(&dynamoTableArg, "t", os.Getenv("DYNAMODB_TABLE"),
		fmt.Sprintf("dynamodb table name, required only for %s backend", dynamoSvc))
	flag.StringVar(&kmsKeyArg, "k", os.Getenv("KMS_KEY"),
		fmt.Sprintf("KMS key ARN, ID, or alias (required for %s backend, optional for %s backend, not used for %s backend",
			dynamoSvc, ssmSvc, secretsSvc))
	flag.BoolVar(&verboseArg, "v", checkBoolEnv("VERBOSE"), "Print verboseArg output")
	flag.BoolVar(&versionArg, "V", false, "Print program versionArg")
}

// interface type for conforming secrets backends
type SecretBackender interface {
	KmsRequired() bool
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
	j := json.NewDecoder(getReader())
	for {
		m := make(map[string]interface{})
		if err := j.Decode(m); err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("error decoding json: %v", err)
			errCnt++
		}

		for k, v := range m {
			if err := sb.Store(k, v); err != nil {
				log.Errorf("error storing secret: %v", err)
				errCnt++
			} else {
				log.Infof("updated secret %s", k)
			}
		}
	}

	os.Exit(errCnt)
}

func checkBoolEnv(v string) (b bool) {
	b, err := strconv.ParseBool(v)
	if err != nil {
		b = false
	}
	return b
}

// verify that we're called with a supported secrets backend
func validateBackend() error {
	backendLc := strings.ToLower(backendArg)
	i := backends.Search(backendLc)

	if i >= len(backends) || backends[i] != backendLc {
		return fmt.Errorf("backend %s is not valid, must be one of: %s", backendArg, strings.Join(backends, ", "))
	}

	if err := backendFactory(backends[i]); err != nil {
		return err
	}

	return nil
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
		sb = NewDynamoDbBackend().WithTable(dynamoTableArg)
	case secretsSvc:
		sb = NewSecretsManagerBackend()
	case ssmSvc:
		sb = NewParameterStoreBackend()
	default:
		return fmt.Errorf("unsupported backend %s", be)
	}
	return nil
}

func readBinary(value interface{}) ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 4096))
	if _, err := fmt.Fprint(b, value); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(b)
}

func getReader() io.Reader {
	in := strings.NewReader(flag.Arg(0))

	b64 := base64.NewDecoder(base64.StdEncoding, in)
	if _, err := b64.Read(make([]byte, 512)); err != nil {
		// not base 64, so can't be something that's compressed, probably just plain text
		// in which case, just return our string reader
		in.Seek(0, io.SeekStart)
		return in
	}

	gz, err := gzip.NewReader(b64)
	if err != nil {
		// I think this will raise an error if it's not a gzip compressed stream
		// in which case, just return the base64 reader
		in.Seek(0, io.SeekStart)
		return b64
	}

	return gz
}
