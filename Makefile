MODULE = github.com/bnb-chain/tss-lib/v2
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

protob:
	@echo "--> Building Protocol Buffers"
	@for protocol in io; do \
		echo "Generating $$protocol.pb.go" ; \
		protoc --go_out=. ./proto/$$protocol.proto ; \
	done
	mv ./common/io.pb.go .
	rm -rf ./common

build: protob
	go fmt ./...

# To avoid unintended conflicts with file names, always add to .PHONY
# # unless there is a reason not to.
# # https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: protob

