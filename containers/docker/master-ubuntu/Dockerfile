FROM ubuntu:xenial

ENV PATH=/usr/lib/go-1.9/bin:$PATH

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang-1.9 git make gcc libc-dev ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/LemoFoundationLtd/lemochain-go && \
  (cd lemochain-go && make glemo) && \
  cp lemochain-go/build/bin/glemo /glemo && \
  apt-get remove -y golang-1.9 git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /lemochain-go

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/glemo"]
