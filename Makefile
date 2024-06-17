clean:
	go clean
	rm -f ./codemaker-sdk-go

compile:
	go build ./...

test:
	go test ./...

build: clean compile test
	@: