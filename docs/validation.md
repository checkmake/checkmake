# Validation

Validation needs to get done via Go code that runs through a Makefile struct
(see [Parsing][parsing]) and validates different rulesets. This should provide
a good foundation to get started and play with some validation rules with the
downside of not being super configurable for special rules.

A generic way to extend rulesets could be a goal later. For now they would
have to be added as code patches to the project itself.

[parsing]: https://github.com/checkmake/checkmake/blob/main/docs/parsing.md
