# Parallel pipelines

  * [gantry.def.yml](./gantry.def.yml)

![pipeline.svg](./pipeline.svg)

Multiple pipelines in the *same* `gantry.def.yml` will be executed in
parallel. Usage of this feature is **not** advised unless these pipelines
**must always** be executed together.

Deterministic execution of steps is only guaranteed inside each pipeline.
