# go-service-kit

A minimal framework and starter kit for building API services in Go, focused on using the standard library and idiomatic Go patterns.

[![CI](https://github.com/Okja-Engineering/go-service-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/Okja-Engineering/go-service-kit/actions/workflows/ci.yml)

## Features

- Simple environment variable helpers
- Structured logging with filtering
- REST API helpers (endpoints, middleware)
- Problem+JSON error responses (RFC-7807)
- JWT authentication utilities
- Built-in metrics and health endpoints
- Linting and CI setup

## Getting Started

This repository is a minimal framework and starter kit for building robust, maintainable API services in Go. It is not an application or server by itself, but a set of packages you can use in your own projects. The included tests serve as validation and usage examples for each package.

To validate the packages and see usage in action:

```sh
git clone https://github.com/Okja-Engineering/go-service-kit.git
cd go-service-kit
go mod tidy
make test   # Run all package tests
make lint   # Lint the codebase
```

## Project Structure

```
pkg/
├── api        # API helpers and endpoints ([docs](pkg/api/README.md))
├── auth       # JWT and auth middleware ([docs](pkg/auth/README.md))
├── env        # Environment variable helpers ([docs](pkg/env/README.md))
├── logging    # Logging utilities ([docs](pkg/logging/README.md))
├── problem    # Problem+JSON error responses ([docs](pkg/problem/README.md))
```

## Usage

This repository is intended as a minimal framework and starter kit for building robust, maintainable API services in Go, with a focus on the standard library and [chi](https://github.com/go-chi/chi) for routing. See each package’s README for details and examples:

- [API](pkg/api/README.md)
- [Auth](pkg/auth/README.md)
- [Env](pkg/env/README.md)
- [Logging](pkg/logging/README.md)
- [Problem](pkg/problem/README.md)

<!-- CONTRIBUTING -->

## Contributing

Please see our [contributing guide](/CONTRIBUTING.md).

## License

MIT (see LICENSE)
