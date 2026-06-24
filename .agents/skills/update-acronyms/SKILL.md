---
name: update-acronyms
description: Define or update acronym rules and exceptions for identifier casing in the generator.
license: Proprietary
compatibility: opencode
---

# Skill: Update Acronyms

Use this skill when the user wants to add a new acronym, change an existing acronym rule, or expand the exception list for a token such as `AI`, `API`, or `HAL`.

## Goal

Keep identifier casing predictable without introducing false positives.

In this repository, acronym handling is used to turn identifier-like strings into the expected Go-style forms. For example:

- `ApiClient` -> `APIClient`
- `OpenAiApiKey` -> `OpenAIAPIKey`

But the rules must avoid rewriting ordinary words by accident. For example, `apiary` should stay `apiary`, and `challenge` should stay `challenge` even if `HAL` is a known acronym.

## Files To Update

The usual files are:

- `generator/internal/acronyms/special-words.yaml`
- `generator/internal/acronyms/acronyms_test.go`

Relevant supporting files:

- `generator/internal/acronyms/special-words.schema.json`
- `generator/internal/acronyms/types.gen.go`
- `generator/internal/acronyms/acronyms.go`

Normally, changing the acronym list only requires updating the YAML data and tests. Do not change the schema or generated types unless the shape of the acronym data itself needs to change.

## Rule Format

Each rule in `special-words.yaml` looks like this:

```yaml
- token: Hal
  replacement: HAL
  exceptions:
    - half
    - challenge
    - inhale
```

Guidance:

- `token` should use the identifier form that the replacer matches, typically Pascal-like casing such as `Api`, `Ai`, or `Hal`
- `replacement` is the final acronym form, typically all caps such as `API`, `AI`, or `HAL`
- `exceptions` should be lowercase natural-language words that must not be rewritten when they contain the token as a substring

## How To Add Or Update A Rule

1. Decide whether the token is genuinely an acronym that should be normalised in identifiers.
2. Add or update the rule in `generator/internal/acronyms/special-words.yaml`.
3. Add practical tests in `generator/internal/acronyms/acronyms_test.go`.
4. Run the focused package tests.

## How To Choose Exceptions

Prefer practical, identifier-relevant exception words over exhaustive dictionaries.

Good exception candidates are:

- common English words that would otherwise be rewritten incorrectly
- domain words likely to appear in configuration, flags, or type names
- words that prove the rule does not over-match substrings inside ordinary identifiers
- words you notice breaking after generation, especially when unit or integration tests start failing because an ordinary identifier was rewritten unexpectedly

Ignore proper nouns unless there is a concrete, recurring identifier case in this repository that justifies adding one. The exception list should mainly protect common words and realistic identifier vocabulary, not try to encode every possible name.

If adding a new acronym rule causes existing unit or integration tests to fail, do not assume the tests should be rewritten straight away. First check whether the failures reveal ordinary words that should be added to the exception list. In practice, regression failures after generation are often how new pragmatic exceptions are discovered.

Examples:

- for `AI`, add words like `acquaint`, `email`, or `maintain`
- for `API`, add words like `apiary` or `capillary`
- for `HAL`, add words like `half`, `challenge`, or `inhale`

## Testing Expectations

Each acronym change should usually include both kinds of tests:

1. Positive rewrite tests
   Example: `HalClient` -> `HALClient`
2. False-positive protection tests
   Example: `challengeMode` stays `challengeMode`

Prefer table-driven examples when a rule has several exceptions.

## Verification

Run:

```bash
go test ./generator/internal/acronyms
```

If the YAML shape changes, additional regeneration or schema work may be needed, but that should be rare.

## Operating Rules

- Keep changes minimal: prefer updating the YAML data and tests over changing code.
- Do not add exceptions speculatively unless there is a realistic false-positive risk.
- Ignore proper nouns by default when building exception lists.
- Use lowercase exception words unless there is a concrete reason not to.
- Preserve existing behaviour for already-covered acronyms unless the user explicitly wants a behavioural change.
- Add tests that demonstrate both the intended rewrite and the protected non-rewrites.
