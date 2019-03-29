package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"testing"
)

type mockSecretsManagerClient struct {
	secretsmanageriface.SecretsManagerAPI
}

func (m *mockSecretsManagerClient) PutSecretValue(input *secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error) {
	if input.SecretId == nil || len(*input.SecretId) < 1 {
		return nil, fmt.Errorf("secret name too short")
	}

	if (input.SecretBinary == nil || len(input.SecretBinary) < 1) &&
		(input.SecretString == nil || len(*input.SecretString) < 1) {
		return nil, fmt.Errorf("secret value too short")
	}

	return &secretsmanager.PutSecretValueOutput{Name: input.SecretId, VersionId: aws.String("VersionX")}, nil
}

func TestNewSecretsManagerBackend(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		b := NewSecretsManagerBackend()
		if b == nil {
			t.Errorf("recieved nil ParameterStoreBackend object")
			return
		}
	})

	t.Run("good", func(t *testing.T) {
		ses = session.Must(session.NewSession())
		b := NewSecretsManagerBackend()
		if b == nil {
			t.Errorf("recieved nil ParameterStoreBackend object")
			return
		}
	})
}

func TestSecretsManagerBackend_KmsRequired(t *testing.T) {
	b := NewSecretsManagerBackend()
	if b.KmsRequired() {
		t.Error("KmsRequired() should be false for SecretsManagerBackend")
	}
}

func TestSecretsManagerBackend_Store(t *testing.T) {
	b := NewSecretsManagerBackend()
	b.c = new(mockSecretsManagerClient)

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
		if err := b.Store("a key", 3.14159); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("bytes value", func(t *testing.T) {
		if err := b.Store("a key", []byte("abcdefg")); err != nil {
			t.Error(err)
			return
		}
	})
}
