---
name: grill-me
description: Use when a generator design or refactor needs to be challenged against existing terminology, ADRs, and repo decisions before implementation.
license: MIT
compatibility: opencode
---

# Skill: Grill Me

Use this skill when the user wants to stress-test a plan against the generator's existing decisions and code.

This skill is influenced by [Matt Pocock's equivalent skill](https://github.com/mattpocock/skills/blob/main/skills/engineering/grill-with-docs/SKILL.md), but adapted for the generator.

## What To Do

Interview the user relentlessly about every meaningful aspect of the plan until there is a shared understanding.

- Walk each branch of the design tree and resolve dependencies between decisions one by one.
- For each question, provide your recommended answer.
- Ask questions one at a time and wait for feedback before continuing.
- If a question can be answered by exploring the codebase, explore the codebase instead of asking.

## Domain Awareness

During codebase exploration, also look for existing documentation.

### Common repo layout

In this repository, architectural decisions are typically recorded here:

```text
/
|- adr/
|  |- ADR-TEMPLATE.md
|  |- adr-index.md
|  \- adr-builtin-types.md
```

Create files lazily. Only create an ADR file when the first ADR is needed.

## During The Session

### Sharpen fuzzy language

When the user uses vague or overloaded terms, propose a precise canonical term.

Example: `You are saying account. Do you mean Customer or User? Those are different things.`

### Contribution

Make sure plan and generated artefacts comply with the contribution guide `CONTRIBUTING.md`

### Discuss concrete scenarios

When domain relationships are being discussed, stress-test them with specific scenarios. Invent scenarios that probe edge cases and force precision about boundaries between concepts.

### Cross-reference with code

When the user states how something works, check whether the code agrees. If there is a contradiction, surface it.

Example: `Your code cancels entire Orders, but you just said partial cancellation is possible. Which is right?`

### Create a news file

Do not create a news file for the plan itself.

If the session results in a new ADR, create a `.doc` news file following `CONTRIBUTING.md` so the ADR appears in the changelog.

Use the format `:book: <adr context>`, where `<adr context>` is a concise, user-facing summary of the ADR.

If multiple ADRs are created, create a separate `.doc` news file for each ADR.

### Offer ADRs sparingly

Only offer to create an ADR when one or more of the following are true:

1. Hard to reverse, or a major refactoring that may be difficult to follow: changing course later would be meaningfully expensive.
2. Surprising without the rationale, and too large for a news file to summarise properly: a future reader would reasonably ask why this approach was chosen.
3. The result of a real trade-off: there were real alternatives and the choice was deliberate.

Use the format in [ADR-TEMPLATE.md](../../../adr/ADR-TEMPLATE.md). Update the [ADR index](../../../adr/adr-index.md) if an ADR is added.

### Connect ADRs to news files

If a change is fundamental and an ADR is created, make sure the corresponding `.doc` news file points readers to that ADR.

## Operating Rules

- Keep the conversation concrete and adversarial in a useful way.
- Prefer resolving terminology before discussing implementation.
- Ground claims in the codebase whenever possible.
- Capture decisions as they happen so documentation stays in sync with the conversation.
