# ADR | Schema: Separate SDK Namespace and Repository URL

# Contributors

- adrcab01
- OpenCode

# Reviewers

- adrcab01

# Context

The generator previously assumed that the Go import namespace and the repository location were the same thing. In practice, that assumption is too rigid for projects that may keep one import namespace while moving the hosted repository to a different organisation for open-source publication, or vice versa.

The generator also emits two different kinds of outputs:

- Go-specific outputs such as modules, imports, mock generation paths, and lint import-prefix configuration.
- Repository-facing outputs such as issue links, pull request links, discussions links, schema source links, and GitHub Action repository defaults.

Treating those as one field makes the manifest less expressive and forces repository layout decisions onto import identity.

## Problem statement

How should the manifest model the SDK import namespace and the hosted repository location so that:

- Go module and import paths remain stable.
- The generator can create tools that do not live in the `Arm-Debug` organisation.
- Repository links and GitHub-specific references can move independently.
- The generator does not emit broken GitHub links for non-GitHub repositories.

# Options

## Option 1: Keep a single field and derive repository location from namespace

Pros:

- Minimal schema change.
- Preserves the previous mental model.

Cons:

- Prevents namespace and repository location from diverging intentionally.
- Makes open-sourcing across organisations awkward.
- Keeps repository links coupled to Go import identity.

## Option 2: Put namespace on individual interface methods

Pros:

- Keeps namespace physically close to method definitions.

Cons:

- The namespace is not method-specific.
- Generated module and import paths are project-wide, not per-method.
- Would duplicate a single concept across methods and risk inconsistency.

## Option 3: Separate `specification.repository_url` from `specification.interface.namespace`

Pros:

- Models repository location and SDK import identity as separate concepts.
- Allows the two values to differ intentionally.
- Keeps the namespace close to the SDK/interface concept rather than repository metadata.
- Supports GitHub-specific rendering only when the repository URL is actually a GitHub repository.

Cons:

- Adds a second field to the manifest.
- Requires template logic to distinguish GitHub-specific and generic repository outputs.

# Design Decision

Adopt Option 3.

The manifest will expose:

- `specification.interface.namespace` for the SDK import namespace.
- `specification.repository_url` for the hosted project location.

The generator will:

- use `interface.namespace` for Go module paths, import paths, mock generation paths, and Go import-prefix lint configuration
- use `repository_url` for repository-facing links and GitHub Action repository defaults
- allow the two values to differ deliberately
- generate GitHub-specific links only when `repository_url` is recognised as a GitHub repository URL
- avoid generating GitHub-specific links when the repository URL is not GitHub

# Impact

Positive consequences:

- The schema can represent projects whose import namespace and hosted repository are intentionally different.
- Generated Go code remains tied to the SDK import identity rather than the hosting platform.
- Repository-facing documentation becomes more accurate when a project is moved or published elsewhere.
- Non-GitHub repository URLs can be accepted without producing broken GitHub links.

Negative consequences:

- Template logic becomes more complex because it must distinguish Go-specific and repository-specific outputs.
- Some generated documentation and action examples are only available when the repository URL is GitHub.

Follow-up work:

- If support for non-GitHub forges becomes important, repository-specific rendering may need a broader forge abstraction rather than GitHub-only detection.

# Comments

- `interface.namespace` is normalised by the generator for Go usage, but the schema description stays language-agnostic.
- This decision is intentionally aimed at the generator manifest model rather than a single generated SDK.
