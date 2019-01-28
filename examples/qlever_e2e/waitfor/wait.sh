#!/bin/sh

while ! wget -q -T 1 --spider $1; do
	sleep 1
done
