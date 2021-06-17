# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: atals android ios atals-cross evm all test clean
.PHONY: atals-linux atals-linux-386 atals-linux-amd64 atals-linux-mips64 atals-linux-mips64le
.PHONY: atals-linux-arm atals-linux-arm-5 atals-linux-arm-6 atals-linux-arm-7 atals-linux-arm64
.PHONY: atals-darwin atals-darwin-386 atals-darwin-amd64
.PHONY: atals-windows atals-windows-386 atals-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

atals:
	$(GORUN) build/ci.go install .
	@echo "Done building."
	@echo "Run \"$(GOBIN)/atals\" to launch atals."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/atals.aar\" to use the library."
	@echo "Import \"$(GOBIN)/atals-sources.jar\" to add javadocs"
	@echo "For more info see https://stackoverflow.com/questions/20994336/android-studio-how-to-attach-javadoc"

ios:
	$(GORUN) build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/atals.framework\" to use the library."

test: all
	$(GORUN) build/ci.go test

lint: ## Run linters.
	$(GORUN) build/ci.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go install golang.org/x/tools/cmd/stringer@latest
	env GOBIN= go install github.com/kevinburke/go-bindata/go-bindata@latest
	env GOBIN= go install github.com/fjl/gencodec@latest
	env GOBIN= go install github.com/golang/protobuf/protoc-gen-go@latest
	env GOBIN= go install ./cmd/abigen
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

atals-cross: atals-linux atals-darwin atals-windows atals-android atals-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/atals-*

atals-linux: atals-linux-386 atals-linux-amd64 atals-linux-arm atals-linux-mips64 atals-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-*

atals-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/atals
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep 386

atals-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/atals
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep amd64

atals-linux-arm: atals-linux-arm-5 atals-linux-arm-6 atals-linux-arm-7 atals-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep arm

atals-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/atals
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep arm-5

atals-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/atals
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep arm-6

atals-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/atals
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep arm-7

atals-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/atals
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep arm64

atals-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/atals
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep mips

atals-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/atals
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep mipsle

atals-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/atals
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep mips64

atals-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/atals
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/atals-linux-* | grep mips64le

atals-darwin: atals-darwin-386 atals-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/atals-darwin-*

atals-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/atals
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/atals-darwin-* | grep 386

atals-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/atals
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/atals-darwin-* | grep amd64

atals-windows: atals-windows-386 atals-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/atals-windows-*

atals-windows-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/atals
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/atals-windows-* | grep 386

atals-windows-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/atals
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/atals-windows-* | grep amd64
