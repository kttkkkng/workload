gen:
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:.

clean-gen:
	rm -rf pb/*.go

build:
	go build -o Invoker Invoker.go Checker.go EventGenerator.go client.go

clean:
	rm Invoker || true
