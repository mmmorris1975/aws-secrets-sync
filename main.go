package main

import (
	"bytes"
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
	backendArg string
	kmsKeyArg  string
	verboseArg bool
	versionArg bool

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
// (-k) KMS_KEY = KMS key arn, id, or alias
// (-v) VERBOSE = verboseArg logging

// options
// -V print versionArg
func init() {
	backends.Sort()

	flag.StringVar(&backendArg, "b", os.Getenv("SECRETS_BACKEND"),
		fmt.Sprintf("Secrets storage backend: %s", strings.Join(backends, ", ")))
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

	validateArgs()

	//arg := flag.Arg(0)
}

func checkBoolEnv(v string) (b bool) {
	b, err := strconv.ParseBool(v)
	if err != nil {
		b = false
	}
	return b
}

// verify that we're called with a supported secrets backend, and, if used by the backend, the KMS key is valid
func validateArgs() {
	backendLc := strings.ToLower(backendArg)
	i := backends.Search(backendLc)

	if i >= len(backends) || backends[i] != backendLc {
		log.Fatalf("backend %s is not valid, must be one of: %s", backendArg, strings.Join(backends, ", "))
	}

	if err := backendFactory(backends[i]); err != nil {
		log.Fatal(err)
	}

	// KMS key is required, or a KMS key was explicitly passed with the parameterstore backend
	if sb.KmsRequired() || (backendArg == ssm.ServiceName && len(kmsKeyArg) > 0) {
		c := kms.New(ses)
		i := kms.DescribeKeyInput{KeyId: aws.String(kmsKeyArg)}
		o, err := c.DescribeKey(&i)
		if err != nil {
			log.Fatalf("failed to lookup KMS key %s, error: %v", kmsKeyArg, err)
		}

		keyArn, err = arn.Parse(*o.KeyMetadata.Arn)
		if err != nil {
			log.Fatalf("bad key ARN: %v", err)
		}
	}
}

func backendFactory(be string) error {
	switch be {
	case dynamoSvc:
		sb = NewDynamoDbBackend()
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
	b := new(bytes.Buffer)
	if _, err := fmt.Fprint(b, value); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(b)
}

/////////////////////////////////////////
//func main() {
//	flag.BoolVar(&verboseArg, "v", false, "Print verboseArg output")
//	flag.Parse()
//
//	if verboseArg {
//		log.SetLevel(logger.DEBUG)
//	}
//
//	arg := flag.Arg(0)
//
//	gz := true
//	data, err := base64.StdEncoding.DecodeString(arg)
//	if err != nil {
//		// bad base64, make a baseless assumption of plain text input
//		log.Debugf("error decoding base64 data: %v", err)
//		data = []byte(arg)
//		gz = false
//	}
//
//	j, err := ioutil.ReadAll(getReader(data, gz))
//	if err != nil {
//		log.Fatalf("error reading data: %v", err)
//	}
//
//	var m map[string]string
//	if err := json.Unmarshal(j, &m); err != nil {
//		log.Fatalf("error unmarshaling json: %v", err)
//	}
//
//	c := aws.NewConfig().WithLogger(log)
//	s := session.Must(session.NewSession(c))
//	sm := secretsmanager.New(s)
//	errCnt := 0
//
//	for k, v := range aws.StringMap(m) {
//		log.Debugf("%s => %s", k, *v)
//
//		r := secretsmanager.PutSecretValueInput{SecretId: aws.String(k), SecretString: v}
//		o, err := sm.PutSecretValue(&r)
//		if err != nil {
//			log.Warnf("error putting secret %s: %v", k, err)
//			errCnt++
//		} else {
//			log.Infof("Updated secret %s", *o.Name)
//		}
//	}
//
//	os.Exit(errCnt)
//}
//
//func getReader(data []byte, gz bool) io.Reader {
//	var r io.Reader
//	var err error
//
//	b := bytes.NewReader(data)
//
//	if gz {
//		r, err = gzip.NewReader(b)
//		if err != nil {
//			// bad gzip, boldly assume incoming data isn't compressed
//			log.Debugf("error creating gzip reader: %v", err)
//			b.Reset(data)
//			r = b
//		}
//	} else {
//		r = b
//	}
//
//	return r
//}
