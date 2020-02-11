# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gtau android ios gtau-cross evm all test clean
.PHONY: gtau-linux gtau-linux-386 gtau-linux-amd64 gtau-linux-mips64 gtau-linux-mips64le
.PHONY: gtau-linux-arm gtau-linux-arm-5 gtau-linux-arm-6 gtau-linux-arm-7 gtau-linux-arm64
.PHONY: gtau-darwin gtau-darwin-386 gtau-darwin-amd64
.PHONY: gtau-windows gtau-windows-386 gtau-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gtau:
	build/env.sh go run build/ci.go install ./cmd/gtau
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gtau\" to launch gtau."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gtau.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gtau.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	./build/clean_go_build_cache.sh
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

gtau-cross: gtau-linux gtau-darwin gtau-windows gtau-android gtau-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gtau-*

gtau-linux: gtau-linux-386 gtau-linux-amd64 gtau-linux-arm gtau-linux-mips64 gtau-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-*

gtau-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gtau
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep 386

gtau-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gtau
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep amd64

gtau-linux-arm: gtau-linux-arm-5 gtau-linux-arm-6 gtau-linux-arm-7 gtau-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep arm

gtau-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gtau
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep arm-5

gtau-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gtau
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep arm-6

gtau-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gtau
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep arm-7

gtau-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gtau
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep arm64

gtau-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gtau
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep mips

gtau-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gtau
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep mipsle

gtau-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gtau
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep mips64

gtau-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gtau
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gtau-linux-* | grep mips64le

gtau-darwin: gtau-darwin-386 gtau-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gtau-darwin-*

gtau-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gtau
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-darwin-* | grep 386

gtau-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gtau
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-darwin-* | grep amd64

gtau-windows: gtau-windows-386 gtau-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gtau-windows-*

gtau-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gtau
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-windows-* | grep 386

gtau-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gtau
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gtau-windows-* | grep amd64
