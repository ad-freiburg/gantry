# Gantry (Container Crane)
[![GoDoc](https://godoc.org/github.com/ad-freiburg/gantry?status.svg)](https://godoc.org/github.com/ad-freiburg/gantry)
[![Build Status](https://travis-ci.org/ad-freiburg/gantry.svg?branch=master)](https://travis-ci.org/ad-freiburg/gantry)
[![Go Report Card](https://goreportcard.com/badge/github.com/ad-freiburg/gantry)](https://goreportcard.com/report/github.com/ad-freiburg/gantry)

Gantry is a pipeline management tool using containers for all relevant steps.
It supports a basic `docker-compose` subset allowing `docker-compose` like
deployments with [wharfer](https://github.com/ad-freiburg/wharfer). If `wharfer`
is not installed `docker` will be used directly.

## Differences between `docker-compose` and `gantry`

- Additional `steps` which can be used for pipelines as they are sequentially
  executed as soon as all required prevous steps are executed instead of all at
  the same time as `services` in `docker-compose`. Steps and services can depend
  on one another. This allows steps to use deployed services. To enable this
  explicit sequentiality the `after` relation is introduced in `steps`.
- Generating `.dot` file showing dependencies using `gantry dot`.

## Build/Download
Binary releases are provided
[here](https://github.com/ad-freiburg/gantry/releases). Alternatively you can
build your own release as described in the *Building a Release* section.

Gantry is go getable through

    go get github.com/ad-freiburg/gantry/cmd/gantry

This will result in a binary without a set version.

## Building a Release
To build a release version first make sure everything works, then edit the
[Setup](#Setup) section of this Readme so the download link points to the
future version. *Only after committing this final change tag the release*

    git tag -a vX.Y.Z -m <message>

Then build with `-ldflags` such that the version is added to the binary

    go build -ldflags="-X github.com/ad-freiburg/gantry.Version=$(git describe --always --long --dirty)" cmd/gantry/gantry.go

Finally use the GitHub Releases mechanism to release a new version

## Installing from source
To install gantry into the users ~/go/bin path it is enough to just run

    go install -ldflags="-X github.com/ad-freiburg/gantry.Version=$(git describe --always --long --dirty)" ./...


## Setup

    # For a build from source
    sudo mv gantry /usr/local/bin/
    # or for the binary release
    cd /tmp
    rm gantry_$(uname -m).tar.bz2
    wget https://github.com/ad-freiburg/gantry/releases/download/v0.2.0/gantry_$(uname -m).tar.bz2
    tar -xavf gantry_$(uname -m).tar.bz2
    sudo mv gantry_$(uname -m)/gantry /usr/local/bin/gantry

    gantry --version
