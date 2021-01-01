# webapp101

This repository is intended to be used in the `webapp101` workshop as an example
of a Web Scraper application.

Participants should do exercises while going through branches from
`0_repo_exercise` to `3_rest_exercise` together with following
 [slides](https://go-talks.appspot.com/github.com/mneverov/webapp101/slides/webapp101.slide#1)
.

This is only an example application, copy-paste on your own risk.

## Run

If you want to explore the repository alone, use the following instructions to
build and run the app.

### Prerequisites

- [go (1.15+)](https://golang.org/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [golangci-lint (1.31+)](https://golangci-lint.run/usage/install/#local-installation)
- [mockery](https://github.com/vektra/mockery)
- make

### Build & Run

```bash
make startdb && make build && ./bin/webapp101 --migrate-up
```

This command will start postgres DB in Docker, test and build the application
and start it, applying DB migrations on start. It will return an error if a 
container with name `webappDB` is already in use. To stop the DB use appropriate
 Docker commands or

```bash
make stopdb
```

### Parameters

There is no mandatory parameters. To see the full list of parameters run

```bash
./bin/webapp101 -h
```

## REST API

One can test the API using http client (Goland) or rest client (VS Code), and
 the provided [requests](./webapp101.http).