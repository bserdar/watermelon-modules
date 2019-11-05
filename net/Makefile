net: proto fmt
	@echo "+ $@"
	@go build 

firewalld/firewalld.pb.go: firewalld/firewalld.proto
	protoc --proto_path=../../watermelon/proto --proto_path=firewalld  --go_out=plugins=grpc:../../../../ $< 

proto: firewalld/firewalld.pb.go

.PHONY: fmt
fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v pb\.go | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

