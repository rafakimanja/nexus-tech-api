# Project api

One Paragraph of project description goes here

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```

Populate the database independently from the API (requires the same database
environment variables as the API, loaded from `.env` or the shell):
```bash
make seed
```
This creates/updates the schema and inserts development areas plus one sample
demandante and professor. It is safe to run more than once. To run directly:
```bash
go run ./cmd/seed
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
