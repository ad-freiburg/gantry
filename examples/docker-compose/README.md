# Using docker-compose.yml

  * [docker-compose.yml](./gdocker-compose.yml)

![pipeline.svg](./pipeline.svg)

This is a slightly modified
[docker-compose example](https://docs.docker.com/compose/wordpress/) as
`docker-compose` supports different markup for environment variables and named
volumes are not implemented into `gantry`.

## Running using docker-compose
    docker-compose up -d
This creates a default network, starts both containers, and detaches from
them. The `-d` flag is provided to match the behavior of gantry which always
detaches from service containers.

To clean up stop the deployment using `docker-compose down`.

## Running using gantry
    gantry up
This creates a default network and starts both containers detached. The `up`
command can be omitted from gantry to simplify running deployments.

To clean up stop the deployment using `gantry down`.
