gen:
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:.

clean-gen:
	rm -rf pb/*.go
