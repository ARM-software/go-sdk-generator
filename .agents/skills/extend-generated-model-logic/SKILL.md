---
name: extend-generated-model-logic
description: Put small traversal, capability, and preprocessing logic on the generated schema model types instead of centralising it in large helper functions.
license: Proprietary
compatibility: opencode
---

# Skill: Extend Generated Model Logic

Use this skill when changing generator behaviour that depends on traversing the generated manifest model types such as:

- capability checks
- preprocessing decisions
- small derived maps or lookups
- rewrite or normalisation passes over nested schema values

## Preferred Pattern

Prefer adding narrowly scoped methods to the generated model types rather than building one large helper with many nested loops.

Examples of good homes for logic:

- `ConfigurationOption`
- `FunctionConfigFields`
- `EntityField`
- `Type`
- `Types`
- `CliFunction`
- `CommandLineCommands`
- `SDKSpecification`

## Style Guidance

- keep read-only capability logic close to the types it describes
- use collection helpers where they naturally fit
- if mutation must happen in place, keep the traversal on the model type even if the internals cannot use the same helper as the read-only path
- prefer composing type methods from higher-level aggregate methods instead of repeating traversal logic in preprocessing functions

## In This Repository

The main reference file for this style is:

- `generator/types_extended_logic.go`

When a preprocessing or validation helper starts growing several nested loops, consider whether the loops should become methods on the model types instead.

## Review Checklist

- does the logic belong on one of the generated schema model types?
- can aggregate types compose lower-level methods instead of reimplementing traversal?
- are collection helpers used where they improve clarity?
- is the remaining top-level preprocessing helper now small and delegating rather than traversing everything itself?
