#!/bin/sh
echo "Update ${PWD}"
gantry dot --output pipeline.dot
dot -T svg -o pipeline.svg pipeline.dot
gantry dot --output partial_execution_0.dot -i test_0 -i test_1
dot -T svg -o partial_execution_0.svg partial_execution_0.dot
gantry dot --hide-ignored --output partial_execution_0_reduced.dot -i test_0 -i test_1 -i test_3
dot -T svg -o partial_execution_0_reduced.svg partial_execution_0_reduced.dot
gantry dot --output partial_execution_1.dot -i test_0 -i test_1
dot -T svg -o partial_execution_1.svg partial_execution_1.dot
gantry dot --hide-ignored --output partial_execution_1_reduced.dot -i test_0 -i test_1 -i test_3
dot -T svg -o partial_execution_1_reduced.svg partial_execution_1_reduced.dot
