#!/bin/sh
echo "Update ${PWD}"
gantry dot -f types.yml --output types.dot
dot -T svg -o types.svg types.dot
gantry dot -f simple.yml --output simple.dot
dot -T svg -o simple.svg simple.dot
gantry dot -f ignored.yml --output ignored.dot -i ignored
dot -T svg -o ignored.svg ignored.dot
gantry dot -f ignored.yml --output ignored_hidden.dot -i ignored --hide-ignored
dot -T svg -o ignored_hidden.svg ignored_hidden.dot
