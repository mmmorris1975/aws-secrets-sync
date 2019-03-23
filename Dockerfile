FROM scratch

ADD ./secrets-sync-* /secrets-sync

ENTRYPOINT ["/secrets-sync"]