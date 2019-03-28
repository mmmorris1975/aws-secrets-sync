package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// SecretsManagerBackend is the type for storing a KMS encrypted item attribute in AWS Secrets Manager
type SecretsManagerBackend struct {
	kmsRequired bool
	c           *secretsmanager.SecretsManager
}

// NewSecretsManagerBackend creates a Secrets Manager SecretsBackender.
func NewSecretsManagerBackend() *SecretsManagerBackend {
	return &SecretsManagerBackend{
		kmsRequired: false,
		c:           secretsmanager.New(ses),
	}
}

// KmsRequired returns whether or not this backend requires a KMS key to encrypt the value when doing
// a Store().  For Secrets Manager this will always be false since the key is defined on the Secret
// definition, and not required when storing values using this backend.
func (b *SecretsManagerBackend) KmsRequired() bool {
	return b.kmsRequired
}

// Store writes the value to Secrets Manager using the name defined by the key parameter. String
// values will be stored as SecretString types, any other data type will be stored as a SecretBinary
// type.  AWS enforces a maximum size of 7168 bytes for the value, so attempting to store values larger
// than that is likely to result in an error.
func (b *SecretsManagerBackend) Store(key string, value interface{}) error {
	i := secretsmanager.PutSecretValueInput{SecretId: aws.String(key)}

	switch t := value.(type) {
	case string:
		i.SecretString = aws.String(t)
	default:
		data, err := readBinary(value)
		if err != nil {
			return err
		}
		i.SecretBinary = data
	}

	o, err := b.c.PutSecretValue(&i)
	if err != nil {
		return err
	}
	log.Debugf("set secret %s, version %s", *o.Name, *o.VersionId)

	return nil
}
