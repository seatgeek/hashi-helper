FROM ubuntu:trusty

RUN apt-get update \
    && apt-get install -y ssl-cert ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY build/hashi-helper-linux-amd64 /hashi-helper

ENTRYPOINT ["/hashi-helper"]
