GitTag=$(version)
os=$(shell go env GOOS)
arch=$(shell go env GOARCH)

build: clean ## 构建二进制
	@echo "build sealer and sealutil bin"
	hack/build.sh $(GitTag)

clean: ## clean
	@rm -rf _output

test_version:
	_output/bin/sealer/$(os)_$(arch)/sealer version
	_output/bin/seautil/$(os)_$(arch)/seautil version
