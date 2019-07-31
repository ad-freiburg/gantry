# Selective run
This example demonstrates the ability to execute only parts of a pipeline
using the `-i` command and explicit step/service selection.

The structure of the example is visualized in [pipeline.svg](./pipeline.svg).

The pipeline depicts how an active service could be updated with minimal
downtime and validation of a new version.

To rerun the validation without preparing a new version
```
gantry -i prepare_new_service_version test_new_service
```
can be used. This forces all gantry to ignore all non-dependencies of
`test_new_service`. `prepare_new_service_version` is explicitly ignored thus
`otherthing` and `something` are not considered prerequests for
`test_new_service` and also ignored.

The resulting structure is visualized in
[test_new_service.svg](./test_new_service.svg) and pruned in
[test_new_service_pruned.svg](./test_new_service_pruned.svg).
