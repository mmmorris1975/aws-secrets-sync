package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"testing"
)

type mockKmsClient struct {
	kmsiface.KMSAPI
}

func (m *mockKmsClient) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	if input.Plaintext == nil || len(input.Plaintext) < 1 {
		return nil, fmt.Errorf("plaintext min length is 1")
	}

	o := new(kms.EncryptOutput)
	o.CiphertextBlob = input.Plaintext
	return o, nil
}

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClient) DescribeTable(input *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
	if *input.TableName == "my-table" {
		o := new(dynamodb.DescribeTableOutput)
		o.Table = &dynamodb.TableDescription{
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					KeyType:       aws.String(dynamodb.KeyTypeHash),
					AttributeName: aws.String("key"),
				},
			},
		}

		return o, nil
	}

	return nil, fmt.Errorf(dynamodb.ErrCodeTableNotFoundException)
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	for k, v := range input.Item {
		if len(k) < 1 {
			return nil, fmt.Errorf("empty key")
		}

		if len(*v.S) < 1 {
			return nil, fmt.Errorf("empty value")
		}
	}
	return nil, nil
}

func TestNewDynamoDbBackend(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		d := NewDynamoDbBackend()
		if d == nil {
			t.Errorf("received nil DynamoDbBackend object")
			return
		}
	})

	t.Run("good", func(t *testing.T) {
		ses = session.Must(session.NewSession())
		d := NewDynamoDbBackend()
		if d == nil {
			t.Errorf("received nil DynamoDbBackend object")
			return
		}
	})
}

func TestDynamoDbBackend_KmsRequired(t *testing.T) {
	d := NewDynamoDbBackend()
	if !d.KmsRequired() {
		t.Errorf("KmsRequired() should never be false for DynamoDB backend")
	}
}

func TestDynamoDbBackend_WithTable(t *testing.T) {
	d := NewDynamoDbBackend()
	d.c = new(mockDynamoDBClient)
	d.k = new(mockKmsClient)

	t.Run("good name", func(t *testing.T) {
		d, err := d.WithTable("my-table")
		if err != nil {
			t.Error(err)
			return
		}

		if d.pk != "key" {
			t.Errorf("unexpected partition key name: %s", d.pk)
			return
		}
	})

	t.Run("bad name", func(t *testing.T) {
		_, err := d.WithTable("x")
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestDynamoDbBackend_Store(t *testing.T) {
	d := NewDynamoDbBackend()
	d.c = new(mockDynamoDBClient)
	d.k = new(mockKmsClient)
	d.table = "my-table"
	d.pk = "key"

	t.Run("good", func(t *testing.T) {
		if err := d.Store("my-key", "a value"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("empty key", func(t *testing.T) {
		if err := d.Store("", "my value"); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("nil value", func(t *testing.T) {
		// kms.Encrypt expects len(value) > 0
		if err := d.Store("a key", nil); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if err := d.Store("akey", ""); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("string value", func(t *testing.T) {
		if err := d.Store("my key", "my value"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("float value", func(t *testing.T) {
		if err := d.Store("float key", 3.14159); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("bytes value", func(t *testing.T) {
		if err := d.Store("bytes value", []byte("abcdefg")); err != nil {
			t.Error(err)
			return
		}
	})
}
