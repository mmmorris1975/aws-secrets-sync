package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	logger "github.com/mmmorris1975/simple-logger"
	"io"
	"io/ioutil"
	"os"
)

var (
	Version string
	verbose bool

	log = logger.StdLogger
)

func main() {
	flag.BoolVar(&verbose, "v", false, "Print verbose output")
	flag.Parse()

	if verbose {
		log.SetLevel(logger.DEBUG)
	}

	arg := flag.Arg(0)

	gz := true
	data, err := base64.StdEncoding.DecodeString(arg)
	if err != nil {
		// bad base64, make a baseless assumption of plain text input
		log.Debugf("error decoding base64 data: %v", err)
		data = []byte(arg)
		gz = false
	}

	j, err := ioutil.ReadAll(getReader(data, gz))
	if err != nil {
		log.Fatalf("error reading data: %v", err)
	}

	var m map[string]string
	if err := json.Unmarshal(j, &m); err != nil {
		log.Fatalf("error unmarshaling json: %v", err)
	}

	c := aws.NewConfig().WithLogger(log)
	s := session.Must(session.NewSession(c))
	sm := secretsmanager.New(s)
	errCnt := 0

	for k, v := range aws.StringMap(m) {
		log.Debugf("%s => %s", k, *v)

		r := secretsmanager.PutSecretValueInput{SecretId: aws.String(k), SecretString: v}
		o, err := sm.PutSecretValue(&r)
		if err != nil {
			log.Warnf("error putting secret %s: %v", k, err)
			errCnt++
		} else {
			log.Debugf("NAME: %s, ARN: %s", *o.Name, *o.ARN)
		}
	}

	os.Exit(errCnt)
}

func getReader(data []byte, gz bool) io.Reader {
	var r io.Reader
	var err error

	b := bytes.NewReader(data)

	if gz {
		r, err = gzip.NewReader(b)
		if err != nil {
			// bad gzip, boldly assume incoming data isn't compressed
			log.Debugf("error creating gzip reader: %v", err)
			b.Reset(data)
			r = b
		}
	} else {
		r = b
	}

	return r
}
