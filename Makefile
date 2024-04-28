protogen:
	@echo "Generating proto definitions"
	protoc --go_out=. --go-grpc_out=. proto/course_category.proto
	@echo "Proto files generated sucessfully"

run:
	@echo "Starting server"
	go run ./cmd/grpcServer/*.go