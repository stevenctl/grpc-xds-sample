docker:
	docker build -t slandow/grpc-xds-sample:latest .

docker.push: docker
	docker push slandow/grpc-xds-sample:latest

build:
	go build -o sample .

genproto:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./greeter/foo.proto
