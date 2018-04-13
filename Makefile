# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: glemo android ios glemo-cross swarm evm all test clean
.PHONY: glemo-linux glemo-linux-386 glemo-linux-amd64 glemo-linux-mips64 glemo-linux-mips64le
.PHONY: glemo-linux-arm glemo-linux-arm-5 glemo-linux-arm-6 glemo-linux-arm-7 glemo-linux-arm64
.PHONY: glemo-darwin glemo-darwin-386 glemo-darwin-amd64
.PHONY: glemo-windows glemo-windows-386 glemo-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

glemo:
	build/env.sh go run build/ci.go install ./cmd/glemo
	@echo "Done building."
	@echo "Run \"$(GOBIN)/glemo\" to launch glemo."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/glemo.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Glemo.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

glemo-cross: glemo-linux glemo-darwin glemo-windows glemo-android glemo-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/glemo-*

glemo-linux: glemo-linux-386 glemo-linux-amd64 glemo-linux-arm glemo-linux-mips64 glemo-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-*

glemo-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/glemo
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep 386

glemo-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/glemo
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep amd64

glemo-linux-arm: glemo-linux-arm-5 glemo-linux-arm-6 glemo-linux-arm-7 glemo-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep arm

glemo-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/glemo
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep arm-5

glemo-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/glemo
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep arm-6

glemo-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/glemo
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep arm-7

glemo-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/glemo
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep arm64

glemo-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/glemo
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep mips

glemo-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/glemo
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep mipsle

glemo-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/glemo
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep mips64

glemo-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/glemo
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/glemo-linux-* | grep mips64le

glemo-darwin: glemo-darwin-386 glemo-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/glemo-darwin-*

glemo-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/glemo
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-darwin-* | grep 386

glemo-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/glemo
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-darwin-* | grep amd64

glemo-windows: glemo-windows-386 glemo-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/glemo-windows-*

glemo-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/glemo
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-windows-* | grep 386

glemo-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/glemo
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/glemo-windows-* | grep amd64
