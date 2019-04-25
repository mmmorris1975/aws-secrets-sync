package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

// DynamoDbBackend is the type for storing a KMS encrypted item attribute in DynamoDB
type DynamoDbBackend struct {
	kmsRequired bool
	c           dynamodbiface.DynamoDBAPI
	k           kmsiface.KMSAPI
	table       string
	pk          string
}

// NewDynamoDbBackend creates a basic DynamoDB SecretsBackender.  Note that the table name
// is not defined with this call, see WithTable() to set that before making any calls to Store()
func NewDynamoDbBackend() *DynamoDbBackend {
	return &DynamoDbBackend{
		kmsRequired: true,
		c:           dynamodb.New(ses),
		k:           kms.New(ses),
	}
}

// WithTable sets the DynamoDB table name to store the encrypted value.  The table will be inspected
// to ensure it exists, and to determine what the Partition/HASH key attribute is
func (b *DynamoDbBackend) WithTable(t string) (*DynamoDbBackend, error) {
	b.table = t

	i := dynamodb.DescribeTableInput{TableName: aws.String(t)}
	o, err := b.c.DescribeTable(&i)
	if err != nil {
		return nil, fmt.Errorf("error describing dynamodb table: %v", err)
	}

	for _, v := range o.Table.KeySchema {
		if *v.KeyType == dynamodb.KeyTypeHash {
			b.pk = *v.AttributeName
		}
	}

	return b, nil
}

// KmsRequired returns whether or not this backend requires a KMS key to encrypt the value when
// doing a Store().  For DynamoDB this will always be true since we need to explicitly do a KMS
// Encrypt before we store the value in the table.
func (b *DynamoDbBackend) KmsRequired() bool {
	return b.kmsRequired
}

// Store writes the value to the table using the Partition key defined in the key parameter
// All attribute values will be stored as String types.  In addition to the Partition key
// attribute, the "encrypted" attribute will be set on the item with a value of "true", and
// the "value" attribute will hold the base64 encoded value of the encrypted value.
//
// KMS limits the size of the encrypted data to 4096 bytes, so attempting to store values larger
// than that is likely to result in an error.
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
			"encrypted": {BOOL: aws.Bool(true)},
		},
	}

	log.Debugf("writing key %s in DynamoDB table %s", key, b.table)
	if _, err := b.c.PutItem(&i); err != nil {
		return err
	}

	return nil
}

// max size of value is 4096 bytes due to max size of KMS encrypt operation input
func (b *DynamoDbBackend) encrypt(value interface{}) (string, error) {
	r, err := readBinary(value)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(r)
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
