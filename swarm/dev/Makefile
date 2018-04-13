.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/LemoFoundationLtd/lemochain-go/cmd/swarm
	go build -o bin/glemo     github.com/LemoFoundationLtd/lemochain-go/cmd/glemo
	go build -o bin/bootnode github.com/LemoFoundationLtd/lemochain-go/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/LemoFoundationLtd/lemochain-go/swarm/...
