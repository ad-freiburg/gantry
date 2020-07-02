#!/bin/sh
echo "Update ${PWD}"
gantry dot --output pipeline.dot
dot -T svg -o pipeline.svg pipeline.dot
