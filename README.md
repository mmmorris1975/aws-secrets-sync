# aws-secrets-sync

[![CircleCI](https://circleci.com/gh/mmmorris1975/aws-secrets-sync.svg?style=svg)](https://circleci.com/gh/mmmorris1975/aws-secrets-sync)
[![Go Report Card](https://goreportcard.com/badge/github.com/mmmorris1975/aws-secrets-sync)](https://goreportcard.com/report/github.com/mmmorris1975/aws-secrets-sync)

The code and Dockerfile necessary to create a program or container which can upload values to AWS services that can store
sensitive data.  This allows sensitive data to be synchronized to AWS for use with tools like Terraform without the need
to store the sensitive data in the Terraform state file.

The program works by taking a value passed in as the 1st argument to the command, or the system standard input, which is
expected to be a json map of keys and values to upload to the service.  It then calls the appropriate AWS service backend
API to store the value. The preferred input format is a base64 encoded, gzip compressed string of the json values to
upload.  Other supported formats are a base64 encoded string of json values (not compressed), or just the raw json value
directly.


Usage
-----
```text
Usage of aws-secrets-sync:
  -V	Print program version
  -b string
    	S3 bucket name, required only for s3 backend, ignored by all others
  -k string
    	KMS key ARN, ID, or alias (required for dynamodb and s3 backends, optional for ssm backend, not used for secretsmanager backend)
  -o	run in one-shot mode, providing the key and value to store on the command line
  -s string
    	Secrets storage backend: dynamodb, s3, secretsmanager, ssm
  -t string
    	DynamoDB table name, required only for dynamodb backend, ignored by all others
  -v	Print verbose output
```

### Environment Variables
The tool behavior can also be modified using environment variables detailed in the table below

| Name             | Description |
|------------------|-------------|
| SECRETS_BACKEND  | The secrets backend to use for managing the secret data. Equivalent to the `-s` option. |
| KMS_KEY          | The KMS key ARN, ID, or alias to use for encrypting the secret data. Equivalent to the `-k` option. |
| VERBOSE          | Print verbose output. Equivalent to the `-v` option. |
| ONE_SHOT         | Use ['one-shot'](#one-shot-mode) mode, storing the key and value from the command line. Equivalent to the `-o` option. |
| DYNAMODB_TABLE   | The DynamoDB table name to use for storing the secrets. Equivalent to the `-t` option.
| S3_BUCKET        | The S3 bucket to use for storing the secrets. Equivalent to the `-b` option. |
| S3_STORAGE_CLASS | Set the S3 storage class for the secrets, defaults to `STANDARD`.  Refer to the [S3 service documentation](https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-compare) for valid values. |


Backends
--------

### SSM
This backend will upload the data to the SSM Parameter Store service as Secure String types using the paths defined in
the JSON keys.  If the using "pathed" or namespaced key names, AWS expects that the path values are separated by the `/`
character, and that the parameter name value starts with a `/`

A KMS key is not required to be supplied when using this backend.  If a key is not provided, the service default key will
be used to encrypt the value.  The service default KMS key alias is `alias/aws/ssm`.

The maximum size of the secret value is 4096 bytes.

#### Example
```text
aws-secrets-sync -s ssm '{"/my/secret": "shhhh, this is a secret!"}'
```

#### IAM Permissions Required
ssm:PutParameter  
kms:DescribeKey  
kms:Encrypt


### Secrets Manager
This backend will upload the data to the Secrets Manager service, using the JSON key as the name of the secret.
If the value is a string, then the data will be stored as a SecretString type, otherwise it will be stored as a
SecretBinary type.  If the using "pathed" or namespaced key names, AWS expects that the path values are separated by the `/`
character.  It is a preferred practice that you do **not** prefix the key path with a `/`, otherwise it will require the
use of a '//' when referencing the parameter name if using the SSM Parameter Store -\> Secrets Manager magic parameter
name link.  (see https://docs.aws.amazon.com/systems-manager/latest/userguide/integration-ps-secretsmanager.html for
more info)

The Secrets Manager service implements 2 distinct API methods, one to create the Secret resource (which contains metadata
about the secret, including the Secret name and KMS key to encrypt with), and the other to define the Secret's value.
This tool assumes that the Secret resource is already defined, and will not create new ones if it finds a key in the
supplied JSON data that does not exist in the AWS service.  Therefore it is important that the name of the Secret in AWS
and the name of the key in the JSON match, in order to update the value.  Since the KMS key is also defined as part of the
Secret resource, it is not necessary to specify a KMS key when using this tool.  (It will be rightly ignored if you do
supply one, however)

The maximum size of the secret value is 7168 bytes.

#### Example
```text
aws-secrets-sync -s secretsmanager '{"my/secret": "shhhh, this is a secret!"}'
```

#### IAM Permissions Required
secretsmanager:PutSecretValue  
kms:DescribeKey  
kms:Encrypt


### DynamoDB
This backend will upload the data to DynamoDB, using the JSON key as the partition key value in the provided table.
Specifying a KMS key to use for encrypting the secret data is required when using this backend, as DynamoDB has no native
ability to encrypt item attributes as part of the API.  The secret data is encrypted using the provided KMS key and stored
as a base64 encoded value of the KMS ciphertext, and is stored using the attribute name `value`.

The tool will inspect the specified DynamoDB table and dynamically determine the partition key attribute name.  Implying
that the DynamoDB table already exists before running this tool.

The maximum size of the secret value is 4096 bytes, as this is the maximum size of plaintext data the KMS service allows
in a single Encrypt call.

#### Example
```text
aws-secrets-sync -s dynamodb -t my-table -k alias/my/key '{"/my/secret": "shhhh, this is a secret!"}'
```

#### IAM Permissions Required
dynamodb:DescribeTable  
dynamodb:PutItem  
kms:DescribeKey  
kms:Encrypt


### S3
This backend will upload the data to S3, using the JSON key as the S3 object key name in the provided bucket.
Specifying a KMS key to use for encrypting the secret data is required when using this backend to correctly upload the
object to S3 with encryption.  S3 transparently encrypts and decrypts the object data, provided that the API keys have
the necessary Get/Put Object permissions, and Encrypt and Decrypt permissions for the KMS key in use.

Since we are leveraging the encryption facilities of the S3 service to encrypt the secret values, in theory the maximum
secret value size is bound only by the limits of the S3 service

#### Examples
Store value as a command argument
```text
aws-secrets-sync -s s3 -b my-bucket -k alias/my/key '{"/my/secret": "shhhh, this is a secret!"}'
```

Store large value from a file
```text
aws-secrets-sync -s s3 -b my-bucket -k alias/my/key < /path/to/my/data
```

#### IAM Permissions Required
s3:GetObject  
s3:PutObject  
kms:DescribeKey  
kms:Encrypt  
kms:GenerateDataKey


One-Shot Mode
-------------
The tool supports execution using a 'one-shot' mode where the key is supplied as a command line argument, and the value
is supplied as either a command line argument, or via stdin.  This allows you to store simple data, or possibly very large
values, without having to roll it into a json document first.

**WARNING** providing the secret value on the command line is a security risk since it will be visible in a process list,
or stored in the command shell history. It is a more secure practice to use stdin file redirection for providing the value,
similar to the second example below.

#### Examples
Providing the key and value on the command line
```text
aws-secrets-sync -s ssm -o my-key my-value
```

Providing the key on the command line and providing the value on stdin (in this case uploading a large file to S3).
```text
aws-secrets-sync -s s3 -b my-bucket -k alias/my-kms-key -o my-key < /path/to/a/large_file
```
The stdin redirect method will not work with the `ssm` backend since it only supports explicit string values, and can not
determine if the redirected file only contains string data.  When using this method with the `secretsmanager` backend,
it will store the data as a SecretsBinary type, for the same reason as the ssm backend.


Docker example
--------------
An example to run the command using the docker container built from the supplied Dockerfile to store gzip'd input in the
SSM Parameter Store service
```text
docker run --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION aws-secrets-sync \
 -s ssm $(echo raw_json | gzip -c | base64 -i -)
```


Terraform usage
---------------
Examples for each of the backend can be found at the following locations:

[dynamodb](resources/terraform_dynamodb.txt)  
[s3](resources/terraform_s3.txt)  
[secretsmanager](resources/terraform_secrets_manager.txt)  
[ssm](resources/terraform_ssm.txt)


IAM sample policy
-----------------
A sample policy detailing the permissions needed to access all of the backends supported by the tool can be found
[here](resources/iam_policy.txt).  To make it relevant for your setup, you may want to replace the `*` values for the
ARN region with a specific region.  You will be required to update the AWS account number value in the resource ARNs with
the appropriate value for your setup.  KMS key ID(s), DynamoDB table name(s), S3 bucket name(s) and object path(s), and
Secrets Manager and SSM paths will all need to be updated to reflect your particular situation.

The policy above only grants permissions to store the data across the various backends, and does not provide the authority
to read or decrypt the stored values.  An example policy for retrieving the secrets can be found
[here](resources/iam_policy_reader.txt), and will require the same modifications as the other policy to be effective.
One thing to keep in mind is that the dynamodb service does not automatically decrypt the item value, like the rest of
the back ends do, so the GetItem call will retrieve the base64 encoded encrypted data, and it will be up to the caller
to "un-base64" the value, then decrypt using the kms:Decrypt operation.

For the pedants out there, yes, the policies could be condensed to a single statement, however the examples clearly
delineate the permissions needed across the various AWS services.


Building
--------
The code for the tool can be built using the default target in the supplied Makefile, which will create a file called
`aws-secrets-sync` in the current directory, appropriate for execution on the platform it was built on.

A local docker container can be built using the `docker` target in the Makefile.  This will compile the tool for Linux,
and use the Dockerfile in the repo to create an image with the name `aws-secrets-sync`, which will be tagged according to
the most recent tag and commit as determined by running `git describe --tags`