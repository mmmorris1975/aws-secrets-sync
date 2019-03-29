package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
)

// S3Backend is the type for storing a KMS encrypted item attribute in S3
type S3Backend struct {
	kmsRequired  bool
	c            *s3manager.Uploader
	k            *kms.KMS
	bucket       string
	storageClass string
}

// NewS3Backend creates a basic S3 SecretsBackender.  Note that the bucket name is not
// defined with this call, see WithBucket() to set that before making any calls to Store()
func NewS3Backend() *S3Backend {
	cls := s3.StorageClassStandard
	if v, ok := os.LookupEnv("S3_STORAGE_CLASS"); ok {
		cls = v
	}

	return &S3Backend{
		kmsRequired:  true,
		c:            s3manager.NewUploader(ses),
		k:            kms.New(ses),
		storageClass: cls,
	}
}

// WithBucket sets the S3 bucket name to store the encrypted data. No validation is performed
// to verify the bucket existence in this method
func (b *S3Backend) WithBucket(bucket string) *S3Backend {
	b.bucket = bucket
	return b
}

// WithStorageClass sets the S3 storage class to store the encrypted data. No validation is performed
// to verify the provided storage class name is valid in this method
func (b *S3Backend) WithStorageClass(cls string) *S3Backend {
	b.storageClass = cls
	return b
}

// KmsRequired returns whether or not this backend requires a KMS key to encrypt the value when
// doing a Store().  For S3 this will always be true since we need to explicitly provide the KMS
// key information when storing an object in S3.
func (b *S3Backend) KmsRequired() bool {
	return b.kmsRequired
}

// Store writes the value to the bucket using the provided key as the object's key in the bucket.
// The size of the secret value to store in S3 is only limited by the S3 object size limit.  This is
// currently 5TB
func (b *S3Backend) Store(key string, value interface{}) error {
	var r io.Reader
	var err error

	switch t := value.(type) {
	case io.Reader:
		r = t
	default:
		r, err = readBinary(t)
		if err != nil {
			return err
		}
	}

	i := s3manager.UploadInput{
		Bucket:               aws.String(b.bucket),
		Key:                  aws.String(key),
		Body:                 r,
		ServerSideEncryption: aws.String(s3.ServerSideEncryptionAwsKms),
		SSEKMSKeyId:          aws.String(keyArn.String()),
		StorageClass:         aws.String(b.storageClass),
	}

	log.Debugf("uploading S3 object to %s", key)
	o, err := b.c.Upload(&i)
	if err != nil {
		return err
	}
	log.Debugf("object uploaded to %s", o.Location)

	return nil
}
