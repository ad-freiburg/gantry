#!/bin/sh

while ! wget -q -T 1 --spider http://qlever:7001/; do
	sleep 1
done
