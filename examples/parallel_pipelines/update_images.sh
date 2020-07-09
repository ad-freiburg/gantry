#!/bin/sh
echo "Update ${PWD}"
gantry dot --output types.dot
dot -T svg -o pipeline.svg pipeline.dot
