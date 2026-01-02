<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.22.1" (2026-01-02)
=====================

Bugfixes
--------

- :bug: fix API key regex which is not supported by go (#20260102144124)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.22.0" (2025-12-31)
=====================

Features
--------

- :sparkles: add ability to specify the licence (#20251231113050)


Bugfixes
--------

- :bug: fix argument/parameter name casing (#20251231111234)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.21.0" (2025-12-30)
=====================

Features
--------

- :sparkles: Better load configuration from system/environment (#20251230141827)


Bugfixes
--------

- :camel: Upgrade dependencies (#20251230141756)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.20.0" (2025-12-30)
=====================

Features
--------

- :sparkles: Add ability to mark a CLI function as Hidden or SilenceUsage (#20251229165247)


Bugfixes
--------

- :bug: Fix typo in template (#20251229165334)
- :gear: Handle automatically the different extensions for yaml (#20251229171340)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.19.0" (2025-12-29)
=====================

Features
--------

- :sparkles: Add a way to specify a command calls Arm online services and so, required arguments will be added if missing (#20251222112700)


Bugfixes
--------

- :camel: Upgrade dependencies (#20251222112614)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.18.0" (2025-12-19)
=====================

Features
--------

- :sparkles: Generate a valid module name with invalid recipe input (#20251218170225)
- :bug: Fix dependabot template file (#20251218170251)


Bugfixes
--------

- Dependency upgrade: alpine-3.23.2 (#20251218051818)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.17.0" (2025-12-16)
=====================

Features
--------

- :sparkles: Add support for controlling the exit code when the command receives an interrupt signal (#20251215161106)


Bugfixes
--------

- Dependency upgrade: golang-1.25.5 (#20251203051707)
- Dependency upgrade: alpine-3.23.0 (#20251204051747)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.16.1" (2025-11-19)
=====================

Bugfixes
--------

- Dependency upgrade: alpine-3.22.2 (#20251009041838)
- :bug: Ensure that pointers are shown correctly in the generated documentation (#20251020144032)
- Dependency upgrade: golang-1.25.4 (#20251110051731)
- :bug: Ensure that `decimal` type uses floats as the default value (#20251118153521)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.16.0" (2025-10-07)
=====================

Features
--------

- :sparkles: Support go-releaser version 2 (#20251007104310)
- :sparkles: Support detect-secrets version 1.5.0 (#20251007104324)


Bugfixes
--------

- Dependency upgrade: golang-1.25.1 (#20250925041555)
- :bug: Ensure that root flags specified as root across multiple commands are not duplicated (#20251007102340)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.15.0" (2025-09-24)
=====================

Features
--------

- :sparkles: Add support for hidden flags. See the `Hidden` field in cobra https://pkg.go.dev/github.com/spf13/cobra. `MarkHidden()` can be used to hide a flag. (#20250828120011)


Bugfixes
--------

- Dependency upgrade: alpine-3.22.1 (#20250812130420)
- Dependency upgrade: golang-1.25.0 (#20250814043649)
- :bug: fix close stack overflow (#20250924164325)
- :camel: Upgrade dependencies (#20250924165037)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.14.0" (2025-08-11)
=====================

Features
--------

- :sparkles: define the manager as Closeable (#20250811133109)
- :sparkles: define short arguments for the generator (#20250811151440)


Bugfixes
--------

- Dependency upgrade: golang-1.24.5 (#20250709050051)
- Dependency upgrade: alpine-3.22.1 (#20250716045300)
- Dependency upgrade: golang-1.24.6 (#20250807050811)
- :camel: Upgrade dependencies (#20250811133148)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.13.1" (2025-07-08)
=====================

Bugfixes
--------

- :bug: Make sure that port validation rule expects a uint32 instead of a string (#20250707161852)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.13.0" (2025-07-07)
=====================

Features
--------

- :sparkles: Add support for binding multiple flags to a single argument (#20250704173259)
- :sparkles: Only regenerate .goreleaser.yml if overwrite is enabled (#20250707150640)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.12.3" (2025-07-04)
=====================

Bugfixes
--------

- :bug: Make sure that gosec tags don't remove newlines in CLI template (#20250703155657)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.12.2" (2025-07-01)
=====================

Bugfixes
--------

- :bug: Make sure that gosec tags don't remove newlines in root template (#20250701093318)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.12.1" (2025-06-25)
=====================

Bugfixes
--------

- :bug: Make sure that you can't specify -h or -v short args (#20250625140230)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.12.0" (2025-06-24)
=====================

Features
--------

- :sparkles: Add support for specifying short args on the CLI functions (#20250623152933)


Bugfixes
--------

- :bug: Make sure that any extra LangTypes are included in the ExtraValidTypes (#20250623152922)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.11.2" (2025-06-20)
=====================

Bugfixes
--------

- :bug: Explicitly call `chmod scripts/pre.sh` so that SDKs generated on windows will work (#20250620143449)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.11.1" (2025-06-20)
=====================

Bugfixes
--------

- Dependency upgrade: golang-1.24.4 (#20250620043232)
- Dependency upgrade: alpine-3.22.0 (#20250620043235)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.11.0" (2025-06-19)
=====================

Features
--------

- :camel: change the schema to use yaml so that yaml aliases can be used (#20250618111945)
- :sparkles: Automatic management of bindings (#20250619104617)
- :sparkles: Generation of [QLTY](https://docs.qlty.sh/qlty-toml) configuration file (#20250619113601)
- :sparkles: Automatically generate and install manpages during installation (#20250619140044)
- :sparkles: Make the output of help and usage commands pretty using fang (#20250619140046)
- :sparkles: Run go mod tidy and generate before releasing (#20250619144539)
- :recycling: dogfooding! (#20250619144625)


Bugfixes
--------

- :bug: Make sure that the arguments used in bindings take precedent over everything else when naming the arguments (#20250618084049)
- :bug: If no argument is specified in a binding, make sure to use the full name of the field (#20250618084123)
- :bug: Make sure that generated types use concrete types instead of interface{} (#20250618112020)
- :camel: Upgrade dependencies (#20250619104756)
- :bug: fix project definition template (#20250619111827)
- :gear: make more files not overwritten if manually modified (#20250619112239)
- :bug: Make sure that gitignore file is generated (#20250619144523)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.10.0" (2025-06-16)
=====================

Features
--------

- :sparkles: Make sure `TTL` is considered as an acronym (#202506161424)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.9.0" (2025-06-16)
====================

Features
--------

- :sparkles: Add support for POSIX style operands (#20250616100028)


Bugfixes
--------

- Dependency upgrade: alpine-3.22.0 (#20250602050743)
- :bug: Make sure that bindings work when using custom types (#20250603105205)
- Dependency upgrade: golang-1.24.4 (#20250606041601)


Misc
----

- #20250609111531


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.8.3" (2025-05-22)
====================

Bugfixes
--------

- :bug: fix problem with names containing identifier (#20250522155216)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.8.2" (2025-05-20)
====================

Bugfixes
--------

- :bug: update acronyms list to include `HTTP` (#20250520153505)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.8.1" (2025-05-20)
====================

Bugfixes
--------

- :bug: ensure correct casing and default value (#20250518001822)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.8.0" (2025-05-15)
====================

Features
--------

- :sparkles: Add a way to pass tool version information to each CLI commands (#20250430153155)


Bugfixes
--------

- :book: Add doc strings to better explain the purpose of the different files and the next steps (#20250430153313)
- :bug: Fix links and doc strings issues (#20250430153347)
- Dependency upgrade: golang-1.24.3 (#20250507044720)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.7.0" (2025-04-25)
====================

Features
--------

- :sparkles: Make go version configurable (#20250422163652)
- :sparkles: `[boom]` change interface name to be more explicit (#20250422181756)
- :sparkles: Improve the argument validations (#20250423172034)
- :sparkles: extend the possible types for lang_type (#20250423172050)


Bugfixes
--------

- :gear: Specify required go tools (#20250422160754)
- :recycle: Split CLI code from SDK code (#20250422160826)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.6.2" (2025-04-22)
====================

Bugfixes
--------

- :camel: Upgrade dependencies (#20250422112455)
- :gear: Use `go 1.24 tool` command for generators (#20250422112541)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.6.1" (2025-04-11)
====================

Bugfixes
--------

- :bug: Make sure that when a default value is a function it doesn't wrap it in quotes (#20250411142251)
- :bug: Make sure that the type of an argument in the CLI is respected (#20250411142323)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.6.0" (2025-04-10)
====================

Features
--------

- :sparkles: Add support for using yml directly (#20250409162959)
- :sparkles: Add information on the installation of the tool to DEVELOPMENT.md template (#20250409163025)


Bugfixes
--------

- Dependency upgrade: golang-1.24.2 (#20250402045325)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.5.0" (2025-03-19)
====================

Features
--------

- :sparkles: Add support for conditional validation of fields (#20250319103210)


Bugfixes
--------

- :bug: Make sure that extra imports are included in config.go and types.go (#20250319115417)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.4.1" (2025-03-13)
====================

Bugfixes
--------

- :bug: Fix variadic argument forwarding in function definition (#20250312175334)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.4.0" (2025-03-11)
====================

Features
--------

- :sparkles: Update `mockgen` import to use `go 24` [tool command](https://go.dev/doc/go1.24#tools) (#20250311104426)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.3.0" (2025-03-10)
====================

Features
--------

- :sparkles: Add support for not regenerating gomod files (#20240918125706)
- :sparkles: Add support for config files (#20240918125718)
- :sparkles: Add `version` command for all CLIs (#20250310161007)


Bugfixes
--------

- :bug: Make sure that the go.mod does not have superfluous replace directives (#20240617115401)
- Dependency upgrade: golang-1.22.6 (#20240813150941)
- :arrow_up: Upgrade dependencies (#20250310161018)
- Dependency upgrade: alpine-3.21.3 (#20250310162303)
- Dependency upgrade: golang-1.24.1 (#20250310162310)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.2.0" (2024-08-13)
====================

Features
--------

- :sparkles: Add support for specifying time.Duration as CLI arguments (#20240813154502)
- :sparkles: Remove inactive linters from golangci-lint template (#20240813154532)


Bugfixes
--------

- Dependency upgrade: alpine-3.20.1 (#20240621043617)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.1.2" (2024-06-06)
====================

Bugfixes
--------

- :book: correct the spelling in the generated documentation (#202406060937)
- Dependency upgrade: golang-1.22.3 (#20240528145224)
- Dependency upgrade: golang-1.22.4 (#20240606043443)
- :bug: Update generator to use correct [goreleaser](https://goreleaser.com/) configuration (#20240606084014)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.1.1" (2024-05-28)
====================

Bugfixes
--------

- Dependency upgrade: alpine-3.20.0 (#20240528145151)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Proprietary
-->
"0.1.0" (2024-05-28)
====================

Features
--------

- :sparkles: Create initial SDK generator (#20240422092734)
