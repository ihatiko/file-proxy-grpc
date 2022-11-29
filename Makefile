gen_proto:
	protoc --go_out=protoc/. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protoc/*.proto