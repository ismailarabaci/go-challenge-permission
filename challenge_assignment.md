# Coding Challenge

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

