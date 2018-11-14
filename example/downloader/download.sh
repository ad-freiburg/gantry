#!/bin/sh
cd $1
for FILE in $2; do
	wget -i $FILE
done
