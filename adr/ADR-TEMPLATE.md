# ADR Template

Create ADRs only for decisions about the generator that are hard to reverse, surprising without context, and the result of a real trade-off.

## Suggested structure

```md
# ADR | Area: Title

> This is the Architectural Decision Record template. To use:
>
> 1. Copy this template to a new file.
> 2. Name the file with a short, specific summary, for example `adr-builtin-types.md`.
> 3. Replace the title placeholder with a clear decision title.
> 4. Remove this quoted section.
> 5. Update the local ADR index, if the repository keeps one.
> 6. Use follow-up commits to record corrections or later changes.

# Contributors

- List of people involved in creating or researching the ADR.

# Reviewers

- List of people involved in reviewing or approving the ADR.

# Context

What is the current situation?

## Problem statement

What issue is driving the need for change?

# Options

The different options available to solve the problem. For each option, try to add a pros and cons table. Focus on architectural models or paradigms rather than only on framework selection.

# Design Decision

What decision is being proposed or has been made?

# Impact

What are the consequences of this change? If it is the result of a trade-off, explain what remains suboptimal. Also note any future improvements or follow-up work.

# Comments

Any comments or notes.
```

## Guidance

- Keep the ADR clear, factual, and concise.
- Keep the ADR focused on the decision and its trade-offs, not on implementation minutiae.
