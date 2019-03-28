package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ParameterStoreBackend struct {
	kmsRequired bool
	c           *ssm.SSM
}

func NewParameterStoreBackend() *ParameterStoreBackend {
	return &ParameterStoreBackend{
		kmsRequired: false,
		c:           ssm.New(ses),
	}
}

func (b *ParameterStoreBackend) KmsRequired() bool {
	return b.kmsRequired
}

func (b *ParameterStoreBackend) Store(key string, value interface{}) error {
	switch t := value.(type) {
	case string:
		i := ssm.PutParameterInput{
			Name:  aws.String(key),
			Value: aws.String(t),
			Type:  aws.String(ssm.ParameterTypeSecureString),
		}

		if len(kmsKeyArg) > 0 {
			i.KeyId = aws.String(keyArn.String())
		}

		o, err := b.c.PutParameter(&i)
		if err != nil {
			return err
		}
		log.Debugf("set parameter %s, version %d", key, *o.Version)
	default:
		return fmt.Errorf("strings are the only supported parameter value for this backend")
	}
	return nil
}
