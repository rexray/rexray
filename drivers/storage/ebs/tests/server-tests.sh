#!/bin/sh

# create new docker volumes named db1, db2, and db3
docker volume create --driver rexray --name db1
docker volume create --driver rexray --name db2
docker volume create --driver rexray --name db3
