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
