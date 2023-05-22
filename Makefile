# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: marker android ios atlas-cross evm all test clean
.PHONY: atlas android ios atlas-cross evm all test clean
.PHONY: atlas-linux atlas-linux-386 atlas-linux-amd64 atlas-linux-mips64 atlas-linux-mips64le
.PHONY: atlas-linux-arm atlas-linux-arm-5 atlas-linux-arm-6 atlas-linux-arm-7 atlas-linux-arm64
.PHONY: atlas-darwin atlas-darwin-386 atlas-darwin-amd64
.PHONY: atlas-windows atlas-windows-386 atlas-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on CGO_ENABLED=1 CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__" go run

atlas:
	$(GORUN) build/ci.go install .
	@echo "Done building."
	@echo "Run \"$(GOBIN)/atlas\" to launch atlas."

marker:
	cd ./cmd/new_marker && go build -o ./new_marker  *.go && mv ./new_marker ../../build/bin/marker
	@echo "Run \"$(GOBIN)/new_marker\" to launch marker."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/atlas.aar\" to use the library."
	@echo "Import \"$(GOBIN)/atlas-sources.jar\" to add javadocs"
	@echo "For more info see https://stackoverflow.com/questions/20994336/android-studio-how-to-attach-javadoc"

ios:
	$(GORUN) build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/atlas.framework\" to use the library."

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

atlas-cross: atlas-linux atlas-darwin atlas-windows atlas-android atlas-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/atlas-*

atlas-linux: atlas-linux-386 atlas-linux-amd64 atlas-linux-arm atlas-linux-mips64 atlas-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-*

atlas-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/atlas
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep 386

atlas-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/atlas
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep amd64

atlas-linux-arm: atlas-linux-arm-5 atlas-linux-arm-6 atlas-linux-arm-7 atlas-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep arm

atlas-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/atlas
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep arm-5

atlas-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/atlas
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep arm-6

atlas-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/atlas
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep arm-7

atlas-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/atlas
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep arm64

atlas-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/atlas
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep mips

atlas-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/atlas
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep mipsle

atlas-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/atlas
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep mips64

atlas-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/atlas
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/atlas-linux-* | grep mips64le

atlas-darwin: atlas-darwin-386 atlas-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/atlas-darwin-*

atlas-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/atlas
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-darwin-* | grep 386

atlas-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/atlas
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-darwin-* | grep amd64

atlas-windows: atlas-windows-386 atlas-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/atlas-windows-*

atlas-windows-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/atlas
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-windows-* | grep 386

atlas-windows-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/atlas
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/atlas-windows-* | grep amd64
