# GO SDK/CLI Generator

[![Go Badge](https://img.shields.io/badge/go-v1.26.0-blue)](https://golang.org/)

SDK/CLI generator from a specification following [schema](https://github.com/Arm-Debug/go-sdk-generator/blob/main/schema.yml)

> :gear: This project also comes with [CLI tools and commands](./DOCUMENTATION.md#cli_tools) for ease of use

You can access the man page for the CLI via:

```sh
man go-sdk-generator
```

## Documentation

To learn more about the SDK and how to use it, please have a look at [our documentation guide](./DOCUMENTATION.md)

## Capabilities

### Formats

As well as `golang` built-in types, special types format are also supported by the genetator for ease of use:

- duration
- uuid
- json
- url
- version, semver
- port
- host
- email
