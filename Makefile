protob:
	@echo "--> Building Protocol Buffers"
	@for protocol in io; do \
		echo "Generating $$protocol.pb.go" ; \
		protoc --go_out=. ./proto/$$protocol.proto ; \
	done
	mv ./common/io.pb.go .
	rm -rf ./common

.PHONY: protob

