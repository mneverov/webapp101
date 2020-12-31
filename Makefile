clean:
	rm -rvf ./bin

runtests: clean
	golangci-lint run ./... &&\
	go test -v -mod=readonly -cover -count=1 -p 1 ./pkg/...

test: stopdb startdb runtests

build: test
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -mod=readonly -o ./bin/webapp101 ./cmd/webapp101

startdb:
	@docker run -d --rm \
			--name=webappDB \
			-p 5544:5432 \
			-e POSTGRES_DB=webapp101_test \
			-e POSTGRES_USER=webapp101 \
			-e POSTGRES_PASSWORD=webapp101 postgres:13-alpine
	@sleep 2

stopdb:
	@docker stop webappDB || exit 0
