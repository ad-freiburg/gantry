# Selective run
This example demonstrates the ability to execute only parts of a pipeline
using the `-i` command and explicit step/service selection.

The structure of the example is visualized in [pipeline.svg](./pipeline.svg).

The pipeline depicts how an active service could be updated with minimal
downtime and test of a new version.

To rerun the test without preparing a new version
```
gantry -i prepare_new_service_version test_new_service
```
can be used. This forces gantry to ignore all non-dependencies of
`test_new_service`. `prepare_new_service_version` is explicitly ignored thus
`pre_prepare_0` and `pre_prepare_1` are not considered dependencies for
`test_new_service` and are ignored.

The resulting structure is visualized in
[test_new_service.svg](./test_new_service.svg) and again with everything but
the executed services and steps removed in
[test_new_service_pruned.svg](./test_new_service_pruned.svg).
