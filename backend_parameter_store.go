package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// ParameterStoreBackend is the type for storing a KMS encrypted item attribute in SSM Parameter Store
type ParameterStoreBackend struct {
	kmsRequired bool
	c           *ssm.SSM
}

// NewParameterStoreBackend creates a SSM Parameter Store SecretsBackender.
func NewParameterStoreBackend() *ParameterStoreBackend {
	return &ParameterStoreBackend{
		kmsRequired: false,
		c:           ssm.New(ses),
	}
}

// KmsRequired returns whether or not this backend requires a KMS key to encrypt the value when doing
// a Store().  For Parameter Store this will always be false since the service will use the service
// default KMS key if one is not explicitly supplied as part of the command.
func (b *ParameterStoreBackend) KmsRequired() bool {
	return b.kmsRequired
}

// Store writes the value to Parameter Store using the name defined by the key parameter. All
// values will be stored as SecureString types.  AWS enforces a maximum size of 4096 bytes for
// the value, so attempting to store values larger than that is likely to result in an error.
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
