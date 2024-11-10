compile:
	protoc --go_out=./ --go_opt=paths=source_relative api/v1/*.proto

test:
	go test -race ./...