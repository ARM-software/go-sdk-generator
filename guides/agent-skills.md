<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Agent Skills guide

Start with [Agent Skills in DOCUMENTATION.md](../DOCUMENTATION.md#agent-skills) for this project's catalogue, installation commands and package location. Skills teach agents how to use a tool; they complement the binary, SDK and CLI documentation.

## Agent-specific setup

The Skills CLI is the recommended cross-agent installer. Use `--agent` to select a tool. Native alternatives below use actual skill directories; no plugin or marketplace manifest is required. Copy the **complete** skill directory, not just `SKILL.md`. After installing or changing configuration, reload skills or restart the agent.

<details>
<summary>Claude Code</summary>

```sh
npx skills add https://github.com/ARM-software/go-sdk-generator --skill '*' --agent claude-code
```

Native alternative: copy a skill directory into `.claude/skills/` (project) or `~/.claude/skills/` (personal). Symlinked skill directories are supported. Invoke `/run-go-sdk-generator` or describe the task. See [Claude Code skills](https://code.claude.com/docs/en/skills).
</details>

<details>
<summary>OpenCode</summary>

```sh
npx skills add https://github.com/ARM-software/go-sdk-generator --skill '*' --agent opencode
```

Native alternative: copy the directories into `.opencode/skills/` or `~/.config/opencode/skills/`. OpenCode also discovers `.agents/skills/`. Explicit local sources can be configured using the `skills` setting supported by your OpenCode version; see [OpenCode skills](https://opencode.ai/docs/skills). Quit and restart OpenCode after configuration changes.
</details>

<details>
<summary>Codex</summary>

```sh
npx skills add https://github.com/ARM-software/go-sdk-generator --skill '*' --agent codex
```

Native alternative: copy or symlink each complete skill directory into `.agents/skills/` or `~/.agents/skills/`. Use `/skills` or mention `$run-go-sdk-generator` in a prompt. See [Codex skills](https://developers.openai.com/codex/skills).
</details>

<details>
<summary>Cursor</summary>

```sh
npx skills add https://github.com/ARM-software/go-sdk-generator --skill '*' --agent cursor
```

Native alternative: copy skill directories into `.cursor/skills/` or `.agents/skills/` for the project, or `~/.cursor/skills/` for personal use. Select a skill with `/` in Agent chat. See [Cursor skills](https://cursor.com/docs/context/skills).
</details>

<details>
<summary>GitHub Copilot</summary>

```sh
npx skills add https://github.com/ARM-software/go-sdk-generator --skill '*' --agent github-copilot
```

Native alternative: copy directories into `.github/skills/` or `.agents/skills/` for project use. Copilot CLI also supports personal skills in `~/.copilot/skills/`. IDE and cloud availability depend on the Copilot surface; see [GitHub's skill documentation](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills).
</details>

<details>
<summary>Gemini CLI</summary>

```sh
npx skills add https://github.com/ARM-software/go-sdk-generator --skill '*' --agent gemini-cli
```

Native alternative for a local checkout:

```sh
gemini skills install ./.agents/skills/run-go-sdk-generator --scope workspace
```

Gemini discovers `.gemini/skills/` and `.agents/skills/`; use `/skills list` and `/skills reload` in a session. See [Gemini CLI skills](https://geminicli.com/docs/cli/skills/).
</details>

<details>
<summary>OpenClaw</summary>

```sh
npx skills add https://github.com/ARM-software/go-sdk-generator --skill '*' --agent openclaw
```

Native alternative: copy complete directories into `<workspace>/skills/` or `~/.openclaw/skills/`. For package-managed skills, add `/etc/go-sdk-generator/.agents/skills` to `skills.load.extraDirs` in your OpenClaw configuration. This directly references the installed files. See [OpenClaw skills](https://docs.openclaw.ai/tools/skills).
</details>

For other tools, consult the [Skills CLI supported agents](https://github.com/vercel-labs/skills/blob/main/docs/agents.md) and the tool's own discovery documentation. Native plugin installation is not advertised here because this project distributes standard skill directories.

## Consuming packaged skills

The package manager installs the complete `.agents/skills/` collection into `/etc/go-sdk-generator/.agents/skills` without touching user settings. This includes handwritten skills and both scopes described below. You can use these files offline, with an already-installed agent or Skills CLI. See the [local installation commands](../DOCUMENTATION.md#skills-included-in-linux-packages).

For Codex, explicitly link one installed skill into the personal discovery directory:

```sh
mkdir -p "$HOME/.agents/skills"
ln -s "/etc/go-sdk-generator/.agents/skills/run-go-sdk-generator" "$HOME/.agents/skills/run-go-sdk-generator"
```

For Claude Code, use `$HOME/.claude/skills` instead. These commands deliberately do not overwrite an existing skill. Repeat for other desired skills. Links follow software package upgrades; remove dangling links after uninstalling the package. Agents running in a container need access to the package directory inside that container.

For OpenClaw, a direct source avoids copying:

```json
{"skills": {"load": {"extraDirs": ["/etc/go-sdk-generator/.agents/skills"]}}}
```

Use the matching native setup above for other agents. A copied skill needs refreshing after a package upgrade; a linked or configured directory uses the package-owned files directly.

## Skill scope

Record the intended audience in the standard frontmatter `metadata` map:

```yaml
metadata:
  scope: "tool"
```

- `tool`: helps consumers use the software, including installation, CLI usage, workflow integration and authoring or validating its inputs. Handwritten manifest-authoring skills belong here too.
- `repo`: helps contributors maintain this particular source repository, such as changing generator internals, updating regression fixtures or writing repository code. Use this value for local contributor workflows.

This is a project convention inside the Agent Skills specification's extensible, string-valued metadata map. It is not a new top-level frontmatter requirement. Existing or externally authored skills without it remain valid and appear as `unspecified` in the catalogue.

Scope is informational for now: it does not filter discovery, installation, validation, embedding or packaging. Linux packages currently include **every skill** under `.agents/skills/`. A future distribution policy can use this classification without conflating handwritten tool guidance with contributor-only skills.

## Validation and maintenance

The same deterministic checks run locally and through the composite action at `actions/validate-skills/action.yml`. The generated `.github/workflows/skills.yml` invokes it with `uses: ./actions/validate-skills`. After checkout, other workflows can use the same local action with optional `path` (default `.`) and `install` (default `true`) inputs. Install Python 3.11+, Go and Node.js 22.20+, then from the repository root:

```sh
go install github.com/coolapso/agent-skills-validator@v0.2.0
python -m pip install -r actions/validate-skills/requirements.txt
python actions/validate-skills/validate_skills.py
python actions/validate-skills/validate_skills.py --install
```

Ensure Go's binary directory is on `PATH`. The installation checks use Skills CLI **1.9.1**, temporary homes and projects, and disable telemetry. They check the complete collection and each individual skill, compare all installed assets, and never execute skill scripts. A local skill collection can also be checked with `python actions/validate-skills/validate_skills.py /path/to/skills --install`.

The validator checks the [Agent Skills specification](https://agentskills.io/specification), including metadata types and names. The small orchestration script adds discovery, duplicate-name checks, local Markdown resource containment and installation fidelity because the validator does not cover these. It discovers additional maintainer-authored skills automatically. Specification errors, missing resources and installation failures block CI. Length recommendations remain warnings; no LLM scoring or external link availability check gates distribution. The project's existing secret detection remains responsible for credentials.

Keep `SKILL.md` frontmatter first. Use `continuous-delivery-scripts>=4.8.2` when updating copyright headers: this release excludes `SKILL.md`. Other bundled files retain their appropriate licence notices.

Author descriptions that say **what** the skill does and **when** to use it. Keep the main instructions focused (under 500 lines and approximately 5,000 tokens is recommended); place details in `references/`, scripts in `scripts/`, and examples or templates in `assets/`. Bundle required resources within each skill and use relative links. Do not require sibling skills or files from the original checkout after installation. Commit scripts with executable permissions when they are meant to be run directly.

Add handwritten skills and their supporting files under `.agents/skills/`; the whole collection is picked up by Linux packaging. Do not add package hooks that modify users' home directories.

## Embedded skills in the generator

If you already have Go SDK Generator installed, run this from a target project:

```sh
go-sdk-generator agent --init
```

This installs its three embedded install, usage and GitHub Action skills into `.agents/skills/` without network access or Node.js. Re-running after upgrading the binary refreshes those files and overwrites matching local copies. For additional handwritten skills, including manifest authoring, use the full repository or Linux package collection. `metadata.scope` does not currently select which skills are embedded.

See [Embedded Agent Skills](embedded-agent-skills.md) for the source-to-binary lifecycle, the purpose of `generator/skills/`, overwrite rules, and how to rebuild and verify `agent --init` after changing a skill.

## Discoverability and registries

Keep the Git repository canonical, with valid, uniquely named skills under `.agents/skills/`. Verify discovery using `npx skills add https://github.com/ARM-software/go-sdk-generator --list`. Public repositories can be installed through Skills CLI and discovered through [Skills.sh](https://skills.sh); indexing and ranking are platform-controlled and are not guaranteed by validation.

[ClawHub](https://github.com/openclaw/clawhub), [SkillsMP](https://skillsmp.com/), [OpenClaw Skills Hub](https://openclaw-skillshub.com/), [Skills Hub](https://skills-hub.ai/), and [MyClaw Skills](https://myclaw.ai/skills) have their own submission, indexing and publishing policies. These portable directories provide a foundation for future publishing without marketplace-specific copies. No automatic marketplace publishing or synthetic installation telemetry is part of this project.
