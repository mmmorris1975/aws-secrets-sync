package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"testing"
)

type mockSsmClient struct {
	ssmiface.SSMAPI
}

func (m *mockSsmClient) PutParameter(input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	if input.Name == nil || len(*input.Name) < 1 {
		return nil, fmt.Errorf("parameter name too short")
	}

	if input.Value == nil || len(*input.Value) < 1 {
		return nil, fmt.Errorf("parameter value too short")
	}

	return &ssm.PutParameterOutput{Version: aws.Int64(1)}, nil
}

func TestNewParameterStoreBackend(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		b := NewParameterStoreBackend()
		if b == nil {
			t.Errorf("received nil ParameterStoreBackend object")
			return
		}
	})

	t.Run("good", func(t *testing.T) {
		ses = session.Must(session.NewSession())
		b := NewParameterStoreBackend()
		if b == nil {
			t.Errorf("received nil ParameterStoreBackend object")
			return
		}
	})
}

func TestParameterStoreBackend_KmsRequired(t *testing.T) {
	b := NewParameterStoreBackend()
	if b.KmsRequired() {
		t.Error("KmsRequired() should be false for ParameterStoreBackend")
	}
}

func TestParameterStoreBackend_Store(t *testing.T) {
	b := NewParameterStoreBackend()
	b.c = new(mockSsmClient)

	t.Run("good", func(t *testing.T) {
		if err := b.Store("key", "secret"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("empty key", func(t *testing.T) {
		if err := b.Store("", "value"); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("nil value", func(t *testing.T) {
		if err := b.Store("my key", nil); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if err := b.Store("a key", ""); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("string value", func(t *testing.T) {
		if err := b.Store("k", "secret value"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("float value", func(t *testing.T) {
		if err := b.Store("a key", 3.14159); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("bytes value", func(t *testing.T) {
		if err := b.Store("a key", []byte("abcdefg")); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestParameterStoreBackend_StoreWithKey(t *testing.T) {
	b := NewParameterStoreBackend()
	b.c = new(mockSsmClient)
	kmsKeyArg = "key"
	keyArn, _ = arn.Parse("arn:aws:kms:us-east-1:01234567891:key/4d4f2a2c-6bc6-4d9b-a50b-7d6f60c761c4")

	t.Run("string value", func(t *testing.T) {
		if err := b.Store("k", "secret value"); err != nil {
			t.Error(err)
			return
		}
	})
}
