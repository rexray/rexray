# Applications

Persistence for applications in containers.

---

## Getting Started

This tutorial will serve as a generic guide for taking Docker images found on
[Docker Hub](http://hub.docker.com) and utilizing persistent external storage
via REX-Ray. This should provide guidance for certain applications, but also
generically so you can add persistence properly to other applications.

## Instructions

The following are a set of instructions for investigating an existing container
image to determine how to properly apply persistence.

The first step is to determine which application you are looking to deploy, then
proceed to its [Docker Hub](http://hub.docker.com) page. In this example, we
will be using [PostgreSQL on Docker Hub](https://hub.docker.com/_/postgres/).

Most application vendors will post their Dockerfile on the main page for that
given image. Many of them will also make them available by version minimally via
a download or via Github. Continuing with our PostgreSQL example, we will use
the [Dockerfile for version 9.3](https://github.com/docker-library/postgres/blob/ed23320582f4ec5b0e5e35c99d98966dacbc6ed8/9.3/Dockerfile)
since it happens to be the default version provided with Ubuntu 14.04.

Properly written Dockerfiles will include the proper information that separates
persistent information from the container image and deployed container. This is
visible when the author of the `Dockerfile` includes a ```VOLUME``` statement to
define where stateful information should be held.

Open the `Dockerfile` and do a search for ```VOLUME``` and take note of the
volumes that will be created for this image. Then we can use REX-Ray or
`Docker` to create the external persistent volume for this image. In the
PostgreSQL 9.3 example, there is a single volume in the Dockerfile:

```
VOLUME /var/lib/postgresql/data
```

The single path or paths listed refer to the volumes that should be attached
when running the container. Following this you can create a volume and attach
it to a container with the `-v` flag.

- [PostgreSQL](https://hub.docker.com/_/postgres/)  
```
$ docker volume create --driver=rexray --name=postgresql --opt=size=<sizeInGB>
$ docker run -d -e POSTGRES_PASSWORD=mysecretpassword --volume-driver=rexray \
    -v data:/var/lib/postgresql/data postgres
```

## Popular Applications
External persistent storage can be applied to any number of applications
including but not limited the following examples.

 * [Cassandra](https://hub.docker.com/_/cassandra/)
 * [PostgreSQL](https://hub.docker.com/_/postgres/)

        $ docker volume create --driver=rexray --name=postgresql --opt=size=<sizeInGB>
        $ docker run -d -e POSTGRES_PASSWORD=mysecretpassword --volume-driver=rexray \
            -v data:/var/lib/postgresql/data postgres

 * [MariaDB](https://hub.docker.com/_/mariadb/)
 * [MongoDB](https://hub.docker.com/_/mongo/)

        $ docker volume create --driver=rexray --name=mongodb --opt=size=<sizeInGB>
        $ docker run -d --volume-driver=rexray -v mongodb:/data/db mongo

 * [MySQL](https://hub.docker.com/_/mysql/)
 * [Redis](https://hub.docker.com/_/redis/)

        $ docker volume create --driver=rexray --name=redis --opt=size=<sizeInGB>
        $ docker run -d --volume-driver=rexray -v redis:/data redis
