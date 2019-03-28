FROM alpine:3.9

RUN apk add -u ca-certificates && apk upgrade \
  && rm -rf /var/*/apk/* /var/log/*

ADD ./aws-secrets-sync-* /aws-secrets-sync

USER nobody

ENTRYPOINT ["/aws-secrets-sync"]