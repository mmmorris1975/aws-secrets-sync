The code and Dockerfile necessary to create a container which can upload values to the AWS Secrets Manager service so
that the sensitive values are not stored in plaintext in the Terraform state file.

The program works by taking a value passed in as the 1st argument to the command, which is expected to be a json map of
keys and values to upload to the service.  The preferred input format is a base64 encoded, gzip compressed string of
the json values to upload.  Other supported formats are a base64 encoded string of json values (not compressed), or just
the raw json value; however using values of these types may cause warning from the program, and are less preferred than
the base64, gzip format for the values.

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
and using the Dockerfile in the repo, create an images with the name `secrets-sync` which will be tagged according to the
most recent tag and commit as determined by running `git describe --tags`