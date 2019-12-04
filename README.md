# Gantry (Container Crane)

[![GoDoc](https://godoc.org/github.com/ad-freiburg/gantry?status.svg)](https://godoc.org/github.com/ad-freiburg/gantry)
[![Licence](https://img.shields.io/github/license/ad-freiburg/gantry)](./LICENSE)
[![Build Status](https://travis-ci.org/ad-freiburg/gantry.svg?branch=master)](https://travis-ci.org/ad-freiburg/gantry)
[![Go Report Card](https://goreportcard.com/badge/github.com/ad-freiburg/gantry)](https://goreportcard.com/report/github.com/ad-freiburg/gantry)
[![Release](https://img.shields.io/github/v/release/ad-freiburg/gantry?include_prereleases)](https://github.com/ad-freiburg/gantry/releases)
[![AUR package](https://repology.org/badge/version-for-repo/aur/gantry.svg)](https://aur.archlinux.org/packages/gantry)

Gantry is a pipeline management tool using containers for all relevant steps.
It supports a basic `docker-compose` subset allowing `docker-compose` like
deployments with [wharfer](https://github.com/ad-freiburg/wharfer). If `wharfer`
is not installed `docker` will be used directly.

## Services and Steps

Services define docker containers which provide a continued service to other
tasks in and outside of the pipeline. They directly resemble the service
concept from `docker-compose` and once started run concurrently to the rest of
the pipeline including anything depending on the service.

Steps on the other hand run to completion and only then are their dependents
executed. Note however, that independent steps are executed concurrently with
each other. Steps are often used for tasks that produce a result that is
needed by their dependents such as a download, creation of a database index
or training of a machine learning model.

The End-to-End tests of QLever [example](./examples/qlever_e2e) demonstrates
the usage and interaction of both container types.

## Installation

### Download a prebuild Release

Binary releases are provided as
[github releases](https://github.com/ad-freiburg/gantry/releases).

    cd /tmp
    rm gantry_$(uname -m).tar.bz2
    wget https://github.com/ad-freiburg/gantry/releases/download/v0.5.0/gantry_$(uname -m).tar.bz2
    tar -xavf gantry_$(uname -m).tar.bz2
    sudo mv gantry_$(uname -m)/gantry /usr/local/bin/gantry

    gantry --version

### Arch Linux

Gantry is available in the [AUR](https://aur.archlinux.org/) as
[gantry](https://aur.archlinux.org/packages/gantry) and
[gantry-git](https://aur.archlinux.org/packages/gantry-git)

### From source

To install gantry into the users `~/go/bin` path it is enough to just run

    go install -ldflags="-X github.com/ad-freiburg/gantry.Version=$(git describe --always --long --dirty)" ./...

To install gantry globally copy it from `~/go/bin` path or use

    go build -ldflags="-X github.com/ad-freiburg/gantry.Version=$(git describe --always --long --dirty)" cmd/gantry/gantry.go
    sudo mv gantry /usr/local/bin/

    gantry --version

### Using go get

Gantry is go getable through

    go get github.com/ad-freiburg/gantry/cmd/gantry

This will result in a binary without a set version.

## Building a (new) Release

To build a release version first make sure everything works, then edit the
[Download a prebuild release](#download-a-prebuild-release) section of this
Readme so the download link points to the future version. *Only after
committing this final change tag the release*

    git tag -a vX.Y.Z -m <message>

Then build with `-ldflags` such that the version is added to the binary

    go build -ldflags="-X github.com/ad-freiburg/gantry.Version=$(git describe --always --long --dirty)" cmd/gantry/gantry.go

Finally use the GitHub Releases mechanism to release a new version.
