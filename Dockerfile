FROM alpine:3.9

RUN apk add -u ca-certificates && apk upgrade \
  && rm -rf /var/*/apk/* /var/log/*

ADD ./secrets-sync-* /secrets-sync

USER nobody

ENTRYPOINT ["/secrets-sync"]