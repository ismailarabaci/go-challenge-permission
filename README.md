# Coding Challenge

[![CI](https://github.com/ismailarabaci/go-challenge-permission/workflows/CI/badge.svg)](https://github.com/ismailarabaci/go-challenge-permission/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ismailarabaci/go-challenge-permission)](https://goreportcard.com/report/github.com/ismailarabaci/go-challenge-permission)
[![codecov](https://codecov.io/gh/ismailarabaci/go-challenge-permission/branch/master/graph/badge.svg)](https://codecov.io/gh/ismailarabaci/go-challenge-permission)

## Setup

- Spin up a docker container with a mysql database

```sh
docker run \
   --detach \
   --rm \
   --name blp-mysql \
   -e MYSQL_ROOT_PASSWORD=my-secret-pw \
   -e MYSQL_DATABASE=blp-coding-challenge \
   -e MYSQL_USER=blp \
   -e MYSQL_PASSWORD=password \
   --volume=$(pwd)/db/initdb:/docker-entrypoint-initdb.d \
   --publish 3306:3306 \
   mysql:8.0
```

- Look around

```sh
$> docker exec -it blp-mysql bash
bash-4.4# mysql -ublp -ppassword
mysql> use blp-coding-challenge
```

## Challenge description

The goal of the challenge is to write server logic (not an actual http server) that can manage users, user groups and control access to them.
The challenge should be completed without using libraries other than the standard library.\
The challenge is divided into stages where each stage builds on the previous stage and can be evaluated separately.

In the file pkg/server/interface.go you will find interfaces that the server needs to implement in order to pass each stage (don't change the interfaces).\
In order to define an SQL database schema add statements to db/initdb/db.sql.

## Testing

### Run Tests Locally

```bash
# Run all tests
go test ./pkg/server/... -v

# Run with coverage
go test ./pkg/server/... -coverprofile=coverage.txt -covermode=atomic

# Run specific stage tests
go test ./pkg/server/... -v -run Test_Stage1

# Run integration tests only
go test ./pkg/server/... -v -run Test_Integration
```

### CI Pipeline

This project uses GitHub Actions for continuous integration:

- **Automated Testing**: Tests run on every push and pull request
- **Multi-Version Support**: Tests against Go 1.21, 1.22, and 1.23
- **MySQL Integration**: Runs with MySQL 8.0 service container
- **Code Quality**: Automated linting with golangci-lint
- **Format Checking**: Ensures consistent code formatting
- **Coverage Reporting**: Uploads coverage to Codecov

The CI pipeline includes:
- **Test Job**: Runs full test suite with race detector and coverage
- **Lint Job**: Runs golangci-lint with custom configuration
- **Format Job**: Checks code formatting and runs go vet

### Test Structure

- **Unit Tests** (`server_test.go`): 24 focused tests organized by stages
- **Integration Tests** (`integration_test.go`): End-to-end HTTP tests with 2 comprehensive scenarios

See `assumptions_and_design_choices.md` for detailed testing approach and patterns used.