# Partial execution

The structure of the example is visualized in [pipeline.svg](./pipeline.svg).

When run without arguments the example starts a `service`, waits for its
availability and then executes a number of tests which are sequenced but
independent.

To skip `test_1` and `test_0` (see
[partial_execution_0.svg](./partial_execution_0.svg)) the following command
should be used:
```
gantry -i test_0 -i test_1
```

If only `test_2` shall be run after the service is available
([partial_execution_1.svg](./partial_execution_1.svg)) the following
command can be used:
```
gantry -i test_0 -i test_1 -i test_3
```

The same result can be achived combining `-i` with the selective run feature:
```
gantry -i test_1 wait_for_service test_2
```
This can reduce the number of explicitly ignored steps while making sure that
none are falsely executed, as only selected steps and their dependencies are
executed unless these dependencies are explicitly ignored.
In this example `test_3` is ignored because of the specific selection of
`test_2`. `test_1` is explicitly marked as ignored and `test_0` is ignored
because it is only a dependency for the ignored `test_1`. If
`wait_for_service` was not explicitly selected it would be ignored too, but as
it is selected `wait_for_service` and the dependency `service` are executed.
