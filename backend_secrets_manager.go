package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type SecretsManagerBackend struct {
	kmsRequired bool
	c           *secretsmanager.SecretsManager
}

func NewSecretsManagerBackend() *SecretsManagerBackend {
	return &SecretsManagerBackend{
		kmsRequired: false,
		c:           secretsmanager.New(ses),
	}
}

func (b *SecretsManagerBackend) KmsRequired() bool {
	return b.kmsRequired
}

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
