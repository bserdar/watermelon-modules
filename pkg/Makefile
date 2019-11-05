pkg: proto fmt
	@echo "+ $@"
	@go build 

yum/yum.pb.go: yum/yum.proto
	protoc --proto_path=../../watermelon/proto --proto_path=yum  --go_out=plugins=grpc:../../../../ $< 

proto: yum/yum.pb.go

.PHONY: fmt
fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v pb\.go | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

