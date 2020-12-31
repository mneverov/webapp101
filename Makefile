clean:
	rm -rvf ./bin

test:
	golangci-lint run ./... &&\
	go test -v -mod=readonly -cover -count=1 ./...

build: test
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -mod=readonly -o ./bin/webapp101 ./cmd/webapp101

startdb:
	@docker run -d --rm \
			--name=webappDB \
			-p 5544:5432 \
			-e POSTGRES_DB=webapp101 \
			-e POSTGRES_USER=webapp101 \
			-e POSTGRES_PASSWORD=webapp101 postgres:13-alpine

stopdb:
	@docker stop webappDB || exit 0
