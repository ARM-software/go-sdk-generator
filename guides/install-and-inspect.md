<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Install and inspect go-sdk-generator

## When to use this guide

Start here before writing a manifest or running the generator. The CLI is named `go-sdk-generator`; its package is also named `go-sdk-generator`.

## Install on Linux

Once the appropriate package repository is configured, use the native package manager:

| System | Install commands |
| --- | --- |
| Debian or Ubuntu | `sudo apt update` then `sudo apt install go-sdk-generator` |
| Fedora or RHEL with dnf | `sudo dnf install go-sdk-generator` |
| RHEL or CentOS with yum | `sudo yum install go-sdk-generator` |

Released Linux packages include agent skills under `/etc/go-sdk-generator/.agents/skills`. For package upgrades, use your package manager's upgrade command. See the [installation skill](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/install-go-sdk-generator/SKILL.md) for platform-specific troubleshooting.

## Run an OCI image

For container-based use, choose a published version and image variant:

```sh
docker run --rm ghcr.io/arm-software/go-sdk-generator:v<version>-alpine --help
```

To generate into the current directory, mount it and set the working directory so input and output paths refer to the mounted files:

```sh
docker run --rm -v "$(pwd):/work" -w /work \
  ghcr.io/arm-software/go-sdk-generator:v<version>-alpine \
  generate -i generate.yml -o /work --go-version 1.27.3
```

Replace `<version>` with a published version and `1.27.3` with your target project's Go version.

### Accessing GHCR

If the image is not publicly accessible to your account, authenticate with GHCR first. With the GitHub CLI already authenticated:

```sh
gh auth token | docker login ghcr.io -u "$(gh api user --jq .login)" --password-stdin
```

If `docker pull` or `docker run` returns `403 Forbidden`, check that your account or CI token has package-read access. Local GitHub CLI credentials may need the `read:packages` scope; refresh them, then log in again:

```sh
docker logout ghcr.io
gh auth refresh --scopes read:packages
gh auth token | docker login ghcr.io -u "$(gh api user --jq .login)" --password-stdin
```

See [GitHub's container registry documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry) for permissions and authentication details.

## Check the installation and contract

```sh
go-sdk-generator --version
go-sdk-generator --help
go-sdk-generator generate --help
go-sdk-generator schema
```

`schema` prints the embedded manifest schema as JSON, suitable for inspecting allowed fields with other tools. The [schema source](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml) is also available in the repository. To install the embedded skills into the current repository, run `go-sdk-generator agent --init`.
