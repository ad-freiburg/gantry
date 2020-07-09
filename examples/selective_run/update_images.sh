#!/bin/sh
echo "Update ${PWD}"
gantry dot --output pipeline.dot
dot -T svg -o pipeline.svg pipeline.dot
gantry dot --output test_new_service.dot -i prepare_new_service_version test_new_service
dot -T svg -o test_new_service.svg test_new_service.dot
gantry dot --output test_new_service_pruned.dot --hide-ignored -i prepare_new_service_version test_new_service
dot -T svg -o test_new_service_pruned.svg test_new_service_pruned.dot
