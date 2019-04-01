package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCheckBoolEnv(t *testing.T) {
	t.Run("string 1", func(t *testing.T) {
		os.Setenv("MY_VAR", "1")
		defer os.Unsetenv("MY_VAR")

		if !checkBoolEnv("MY_VAR") {
			t.Error("unexpected false")
			return
		}
	})

	t.Run("string 0", func(t *testing.T) {
		os.Setenv("MY_VAR", "0")
		defer os.Unsetenv("MY_VAR")

		if checkBoolEnv("MY_VAR") {
			t.Error("unexpected true")
			return
		}
	})

	t.Run("t", func(t *testing.T) {
		os.Setenv("MY_VAR", "t")
		defer os.Unsetenv("MY_VAR")

		if !checkBoolEnv("MY_VAR") {
			t.Error("unexpected false")
			return
		}
	})

	t.Run("T", func(t *testing.T) {
		os.Setenv("MY_VAR", "T")
		defer os.Unsetenv("MY_VAR")

		if !checkBoolEnv("MY_VAR") {
			t.Error("unexpected false")
			return
		}
	})

	t.Run("TRUE", func(t *testing.T) {
		os.Setenv("MY_VAR", "TRUE")
		defer os.Unsetenv("MY_VAR")

		if !checkBoolEnv("MY_VAR") {
			t.Error("unexpected false")
			return
		}
	})

	t.Run("true", func(t *testing.T) {
		os.Setenv("MY_VAR", "true")
		defer os.Unsetenv("MY_VAR")

		if !checkBoolEnv("MY_VAR") {
			t.Error("unexpected false")
			return
		}
	})

	t.Run("True", func(t *testing.T) {
		os.Setenv("MY_VAR", "True")
		defer os.Unsetenv("MY_VAR")

		if !checkBoolEnv("MY_VAR") {
			t.Error("unexpected false")
			return
		}
	})

	t.Run("random string", func(t *testing.T) {
		os.Setenv("MY_VAR", "alwet")
		defer os.Unsetenv("MY_VAR")

		if checkBoolEnv("MY_VAR") {
			t.Error("unexpected true")
			return
		}
	})
}

func TestValidateBackend(t *testing.T) {
	defer func() { backendArg = "" }()

	t.Run("valid", func(t *testing.T) {
		backendArg = "SSM"
		if err := validateBackend(); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("invalid", func(t *testing.T) {
		backendArg = "ParameterStore"
		if err := validateBackend(); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestValidateKey(t *testing.T) {
	t.Run("kms required", func(t *testing.T) {
		t.Skip("requires kms")
	})

	t.Run("kms not required", func(t *testing.T) {
		sb = NewSecretsManagerBackend()
		if err := validateKey(); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("ssm, with key", func(t *testing.T) {
		t.Skip("requires kms")
	})

	t.Run("ssm, no key", func(t *testing.T) {
		sb = NewParameterStoreBackend()
		kmsKeyArg = ""
		if err := validateKey(); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("nil backend", func(t *testing.T) {
		sb = nil
		if err := validateKey(); err != nil {
			t.Error(err)
			return
		}
	})
}

func TestBackendFactory(t *testing.T) {
	t.Run("dynamodb", func(t *testing.T) {
		t.Skip("requires setting mock dynamodb client")
	})

	t.Run("dynamodb no table", func(t *testing.T) {
		dynamoTableArg = ""
		if err := backendFactory("dynamodb"); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("secretsmanager", func(t *testing.T) {
		if err := backendFactory("secretsmanager"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("ssm", func(t *testing.T) {
		if err := backendFactory("ssm"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("s3", func(t *testing.T) {
		bucketArg = "my-bucket"
		if err := backendFactory("s3"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("s3 no bucket", func(t *testing.T) {
		bucketArg = ""
		if err := backendFactory("s3"); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if err := backendFactory("invalid"); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestReadBinary(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if _, err := readBinary(nil); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("string", func(t *testing.T) {
		r, err := readBinary("x")
		if err != nil {
			t.Error(err)
			return
		}

		if r == nil {
			t.Error("unexpected nil")
			return
		}
	})

	t.Run("int", func(t *testing.T) {
		r, err := readBinary(1)
		if err != nil {
			t.Error(err)
			return
		}

		if r == nil {
			t.Error("unexpected nil")
			return
		}
	})

	t.Run("float", func(t *testing.T) {
		r, err := readBinary(math.Pi)
		if err != nil {
			t.Error(err)
			return
		}

		if r == nil {
			t.Error("unexpected nil")
			return
		}
	})

	t.Run("bool", func(t *testing.T) {
		r, err := readBinary(false)
		if err != nil {
			t.Error(err)
			return
		}

		if r == nil {
			t.Error("unexpected nil")
			return
		}
	})

	t.Run("bytes", func(t *testing.T) {
		r, err := readBinary([]byte("abcdefg"))
		if err != nil {
			t.Error(err)
			return
		}

		if r == nil {
			t.Error("unexpected nil")
			return
		}
	})
}

func TestGetReader(t *testing.T) {
	t.Run("plain string arg", func(t *testing.T) {
		r := getReader("test string 5000")
		if x := reflect.TypeOf(r).String(); x != "*strings.Reader" {
			t.Errorf("unexpected reader value %s", x)
		}
	})

	t.Run("base64 string arg", func(t *testing.T) {
		b64 := base64.StdEncoding.EncodeToString([]byte("my encoded value"))
		r := getReader(b64)
		if x := reflect.TypeOf(r).String(); x != "*base64.decoder" {
			t.Errorf("unexpected reader value %s", x)
		}
	})

	t.Run("base64 gzip arg", func(t *testing.T) {
		b := new(bytes.Buffer)
		gz := gzip.NewWriter(b)
		gz.Write([]byte("test"))

		b64 := base64.StdEncoding.EncodeToString(b.Bytes())
		r := getReader(b64)
		if x := reflect.TypeOf(r).String(); x != "*gzip.Reader" {
			t.Errorf("unexpected reader value %s", x)
		}
	})

	t.Run("base64 stdin", func(t *testing.T) {
		b64 := base64.StdEncoding.EncodeToString([]byte("my encoded value"))
		r := getReader(strings.NewReader(b64))
		if x := reflect.TypeOf(r).String(); x != "*base64.decoder" {
			t.Errorf("unexpected reader value %s", x)
		}
	})

	t.Run("base64 gzip stdin", func(t *testing.T) {
		b := new(bytes.Buffer)
		gz := gzip.NewWriter(b)
		gz.Write([]byte("test"))

		b64 := base64.StdEncoding.EncodeToString(b.Bytes())
		r := getReader(strings.NewReader(b64))
		if x := reflect.TypeOf(r).String(); x != "*gzip.Reader" {
			t.Errorf("unexpected reader value %s", x)
		}
	})

	t.Run("simple bytes", func(t *testing.T) {
		g, _ := time.Now().GobEncode()
		r := getReader(bytes.NewBuffer(g).Bytes())
		if x := reflect.TypeOf(r).String(); x != "*bytes.Reader" {
			t.Errorf("unexpected reader value %s", x)
		}
	})
}

func TestOneShotHandler(t *testing.T) {
	sb = newMockBackend()

	t.Run("good", func(t *testing.T) {
		if err := oneShotHandler("my-key", "my-value"); err != nil {
			t.Error(err)
		}
	})

	t.Run("empty key", func(t *testing.T) {
		if err := oneShotHandler("", "my-value"); err == nil {
			t.Error("did not receive expected error")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		if err := oneShotHandler("my-key", nil); err == nil {
			t.Error("did not receive expected error")
		}
	})
}

func TestJsonHandler(t *testing.T) {
	sb = newMockBackend()

	t.Run("good", func(t *testing.T) {
		if errs := jsonHandler(`{"my-key": "my-value"}`); errs > 0 {
			t.Error("got an error when storing a known good value")
			return
		}
	})

	t.Run("bad json", func(t *testing.T) {
		if errs := jsonHandler("this is not json"); errs < 1 {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("empty json", func(t *testing.T) {
		// since this is a map type, if there's nothing inside to iterate over, the code will fall through without error
		if errs := jsonHandler("{}"); errs > 0 {
			t.Error("got an error when storing empty json")
			return
		}
	})

	t.Run("empty json value", func(t *testing.T) {
		if errs := jsonHandler(`{"my-key": ""}`); errs < 1 {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("empty input", func(t *testing.T) {
		// reader returns EOF, no other error returned
		if errs := jsonHandler(""); errs > 0 {
			t.Error("got an error when storing empty string")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		if errs := jsonHandler(nil); errs < 1 {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("nested json", func(t *testing.T) {
		j := `{"M1": {"k1": "v1", "k2": ["v2", "v3"]}, "L1": ["e1", {"e2": "v4"}]}`
		if errs := jsonHandler(j); errs > 0 {
			t.Error("got an error when storing a known good value")
			return
		}
	})
}
