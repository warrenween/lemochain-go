# Build Glemo in a stock Go builder container
FROM golang:1.10-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /lemochain-go
RUN cd /lemochain-go && make glemo

# Pull Glemo into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /lemochain-go/build/bin/glemo /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp 30304/udp
ENTRYPOINT ["glemo"]
