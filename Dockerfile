FROM scratch

ADD secrets-sync /secrets-sync

USER nobody

ENTRYPOINT ["/secrets-sync"]