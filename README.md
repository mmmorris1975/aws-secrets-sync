The code and Dockerfile necessary to create a utility or container which can upload values to AWS services which can store
sensitive data.  This allows sensitive data to be synchronized to AWS for use with tools like Terraform without the need
to store the sensitive data in the Terraform state file.

The program works by taking a value passed in as the 1st argument to the command, which is expected to be a json map of
keys and values to upload to the service, and calling the appropriate AWS service backend API to store the value. The
preferred input format is a base64 encoded, gzip compressed string of the json values to upload.  Other supported formats
are a base64 encoded string of json values (not compressed), or just the raw json value directly; however using these
types of input data may cause warning from the program, and are less preferred than the base64, gzip format for the values.

Usage
-----
```text
Usage of ./secrets-sync:
  -V	Print program version
  -b string
    	S3 bucket name, required only for s3 backend, ignored by all others
  -k string
    	KMS key ARN, ID, or alias (required for dynamodb and s3 backends, optional for ssm backend, not used for secretsmanager backend)
  -s string
    	Secrets storage backend: dynamodb, s3, secretsmanager, ssm
  -t string
    	dynamodb table name, required only for dynamodb backend, ignored by all others
  -v	Print verbose output
```

Backends
--------

### SSM


### Secrets Manager


### DynamoDB


### S3



If using this tool as part of the `secrets-manager` terraform module, then the input data encoding and compression is
automatically handled for you.  If using the container or command outside of the terraform module, you will need to
ensure the incoming value is in the appropriate format.

Docker example
--------------
```text
docker run --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION secrets-sync \
 $(echo raw_json | gzip -c | base64 -i -)
```

Shell example
-------------
```text
/path/to/secrets-sync $(echo raw_json | gzip -c | base64 -i -)
```

Building
--------

The code for the tool can be built using the default target in the supplied Makefile, which will create a file called
`secrets-sync` in the current directory, appropriate for execution on the platform it was built on.

A local docker container can be built using the `docker` target in the Makefile.  This will compile the tool for Linux,
and use the Dockerfile in the repo to create an images with the name `secrets-sync`, which will be tagged according to
the most recent tag and commit as determined by running `git describe --tags`