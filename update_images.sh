#!/bin/sh

ROOT=${PWD}
for script in `find ./*/ -name update_images.sh`; do
	echo ${script}
	cd `dirname ${ROOT}/${script}`
	./update_images.sh
done
