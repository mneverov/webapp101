clean:
	rm -rvf ./bin
	find . -name "mock_*.go" -delete

generate: clean
	go generate ./...

test: generate
	golangci-lint run ./... &&\
	go test -v -mod=readonly -cover -count=1 ./...

# Check list of available GOOS and GOARCH here https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
build: test
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -mod=readonly -o ./bin/webapp101 ./cmd/webapp101
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -v -mod=readonly -o ./bin/webapp101_win ./cmd/webapp101
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -v -mod=readonly -o ./bin/webapp101_osx ./cmd/webapp101

startdb:
	@docker run -d --rm \
			--name=webappDB \
			-p 5544:5432 \
			-e POSTGRES_DB=webapp101 \
			-e POSTGRES_USER=webapp101 \
			-e POSTGRES_PASSWORD=webapp101 postgres:13-alpine

stopdb:
	@docker stop webappDB || exit 0
