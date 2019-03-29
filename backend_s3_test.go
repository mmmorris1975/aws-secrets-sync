package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
	"testing"
)

// We're kind of limited to the tests we can run on the S3 backend code
// AWS provides an s3iface for mocking out S3 service calls, however it
// doesn't support the s3manager module we're using to handle a lot of
// the details around uploading data to S3.  Until we have a similar way
// to mock out the s3manager stuff, we won't be able to do any testing on
// the Store() method

func TestNewS3Backend(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		d := NewS3Backend()
		if d == nil {
			t.Errorf("received nil DynamoDbBackend object")
			return
		}
	})

	t.Run("good", func(t *testing.T) {
		ses = session.Must(session.NewSession())
		d := NewS3Backend()
		if d == nil {
			t.Errorf("received nil DynamoDbBackend object")
			return
		}
	})

	t.Run("storage class env var", func(t *testing.T) {
		os.Setenv("S3_STORAGE_CLASS", "cls")
		defer os.Unsetenv("S3_STORAGE_CLASS")
		d := NewS3Backend()

		if d.storageClass != "cls" {
			t.Error("storage class mismatch")
		}
	})
}

func TestS3Backend_KmsRequired(t *testing.T) {
	d := NewS3Backend()
	if !d.KmsRequired() {
		t.Errorf("KmsRequired() should never be false for DynamoDB backend")
	}
}

func TestS3Backend_WithBucket(t *testing.T) {
	d := NewS3Backend()

	if d = d.WithBucket("my-bucket"); d == nil {
		t.Error("received nil S3Backend object")
	}

	if d.bucket != "my-bucket" {
		t.Error("bucket name mismatch")
	}
}

func TestS3Backend_WithStorageClass(t *testing.T) {
	d := NewS3Backend()

	if d = d.WithStorageClass("q"); d == nil {
		t.Error("received nil S3Backend object")
	}

	if d.storageClass != "q" {
		t.Error("storage class mismatch")
	}
}
