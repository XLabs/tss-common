
protob:
	@echo "--> Building Protocol Buffers"
	@PROTO_FILES=$$(find proto -name '*.proto'); \
	if [ -z "$$PROTO_FILES" ]; then \
		echo "No .proto files found in proto/ directory."; \
	else \
		for p in $$PROTO_FILES; do \
			echo "Generating $${p%.proto}.pb.go"; \
			protoc --go_out=. --go-grpc_out=. $$p; \
		done; \
	fi
	mv ./github.com/xlabs/tss-common/io.pb.go ./io.pb.go
	rm -rf ./github.com
.PHONY: protob

