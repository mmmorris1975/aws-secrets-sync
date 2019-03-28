package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type S3Backend struct {
	kmsRequired  bool
	c            *s3manager.Uploader
	k            *kms.KMS
	bucket       string
	storageClass string
}

func NewS3Backend() *S3Backend {
	return &S3Backend{
		kmsRequired:  true,
		c:            s3manager.NewUploader(ses),
		k:            kms.New(ses),
		storageClass: s3.StorageClassStandard,
	}
}

func (b *S3Backend) WithBucket(bucket string) *S3Backend {
	b.bucket = bucket
	return b
}

func (b *S3Backend) WithStorageClass(cls string) *S3Backend {
	b.storageClass = cls
	return b
}

func (b *S3Backend) KmsRequired() bool {
	return b.kmsRequired
}

// max value size is only limited by S3 object limits
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

// max size of value is 4096 bytes due to max size of KMS encrypt operation input
//func (b *S3Backend) encrypt(value interface{}) (io.Reader, error) {
//	data, err := readBinary(value)
//	if err != nil {
//		return nil, err
//	}
//
//	i := kms.EncryptInput{KeyId: aws.String(keyArn.String()), Plaintext: data}
//	o, err := b.k.Encrypt(&i)
//	if err != nil {
//		return nil, err
//	}
//	log.Debugf("successfully encrypted data")
//
//	// Encrypt API call returns bytes, unlike the dynamodb backend, we'll store the bytes directly in S3
//	return bytes.NewReader(o.CiphertextBlob), nil
//}
