.PHONY: protos

protos: 
	protoc --go_out=. --go-grpc_out=. chat.proto