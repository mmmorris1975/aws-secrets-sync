package main

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
)

type DynamoDbBackend struct {
	kmsRequired bool
	c           *dynamodb.DynamoDB
	k           *kms.KMS
	table       string
	pk          string
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

	// describe the dynamodb table to find determine the HASH key for the table
	// instead of having static expectations for the key attribute name
	i := dynamodb.DescribeTableInput{TableName: aws.String(t)}
	o, err := b.c.DescribeTable(&i)
	if err != nil {
		// todo handle error
	}

	for _, v := range o.Table.KeySchema {
		if *v.KeyType == dynamodb.KeyTypeHash {
			b.pk = *v.AttributeName
		}
	}

	return b
}

func (b *DynamoDbBackend) KmsRequired() bool {
	return b.kmsRequired
}

func (b *DynamoDbBackend) Store(key string, value interface{}) error {
	data, err := b.encrypt(value)
	if err != nil {
		return err
	}
	log.Debugf("DynamoDB Encrypted: %s", data)

	i := dynamodb.PutItemInput{
		TableName: aws.String(b.table),
		Item: map[string]*dynamodb.AttributeValue{
			b.pk:        {S: aws.String(key)},
			"value":     {S: aws.String(data)},
			"encrypted": {S: aws.String("true")},
		},
	}

	if _, err := b.c.PutItem(&i); err != nil {
		return err
	}

	return nil
}

// max size of value is 4096 bytes due to max size of KMS encrypt operation input
func (b *DynamoDbBackend) encrypt(value interface{}) (string, error) {
	data, err := readBinary(value)
	if err != nil {
		return "", err
	}

	i := kms.EncryptInput{KeyId: aws.String(keyArn.String()), Plaintext: data}
	o, err := b.k.Encrypt(&i)
	if err != nil {
		return "", err
	}
	log.Debugf("successfully encrypted data")

	// Encrypt API call returns bytes, encode to base64 and return
	return base64.StdEncoding.EncodeToString(o.CiphertextBlob), nil
}
