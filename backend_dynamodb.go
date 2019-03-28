package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
)

type DynamoDbBackend struct {
	kmsRequired bool
	c           *dynamodb.DynamoDB
	k           *kms.KMS
	table       string
}

func NewDynamoDbBackend() *DynamoDbBackend {
	return &DynamoDbBackend{
		kmsRequired: true,
		c:           dynamodb.New(ses),
		k:           kms.New(ses),
	}
}

func (b *DynamoDbBackend) WithTable(t string) *DynamoDbBackend {
	b.table = t
	return b
}

func (b *DynamoDbBackend) KmsRequired() bool {
	return b.kmsRequired
}

func (b *DynamoDbBackend) Store(key string, value interface{}) error {
	_, err := b.encrypt(value)
	if err != nil {
		return err
	}
	return nil
}

// max size of value is 4096 bytes due to max size of KMS encrypt operation input
func (b *DynamoDbBackend) encrypt(value interface{}) ([]byte, error) {
	data, err := readBinary(value)
	if err != nil {
		return nil, err
	}

	// AWS SDK says that the encrypted value is automatically base64 encoded, although I'm skeptical
	// maybe they mean they automatically encode what's coming and going over the wire?
	// max input size for Plaintext is 4096 bytes
	i := kms.EncryptInput{KeyId: aws.String(keyArn.String()), Plaintext: data}
	o, err := b.k.Encrypt(&i)
	if err != nil {
		return nil, err
	}
	log.Debugf("successfully encrypted data")
	return o.CiphertextBlob, nil
}
