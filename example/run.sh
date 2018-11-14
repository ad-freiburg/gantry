#!/bin/sh

# Copy demo files to destination
cp -r downloader /tmp
chmod 777 /tmp/downloader/out
cp -r catter /tmp

# Execute gantry
go run ../cmd/gantry/main.go showgoogle.def.yml showgoogle.env.yml

#clean up
rm -rf /tmp/downloader /tmp/catter
