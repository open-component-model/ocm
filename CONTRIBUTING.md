# Contributing to the OCM CLI

For the general contribution process (fork-and-pull workflow, commit requirements, code of conduct, and more), see the
[central contributing guide](https://ocm.software/community/contributing/) on the project website.

This document covers repository-specific development details.

## Prerequisites

- **Go 1.26+**
- **Make**
- **Docker** - required for integration tests

## Project Structure

```text
.
├── api/           # Core Go library (OCM, OCI, credentials, datacontext)
├── cmds/          # CLI entry-points and plugins
├── components/    # Component definitions for building OCM component versions
├── docs/          # CLI reference, plugin reference, ADRs
├── examples/      # Usage examples
├── hack/          # Development scripts (generate, format, install, cross-build)
├── Makefile       # Build automation
└── VERSION        # Current version
```

## Common Tasks

```bash
# Build all binaries
make build

# Build ocm cli
make bin/ocm

# Run all tests (unit + integration; requires Docker)
make test

# Run unit tests only
make unit-test

# Format code (gci + gofumpt)
make format

# Lint
make check

# Lint with auto-fix
make check-fix

# Full pipeline: generate, format, generate-deepcopy, build, test, lint
make prepare

# Install dev dependencies (vault, oci-registry, ...)
make install-requirements
```

## Testing

Tests use [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/).

| Build tag | Purpose | Requirements |
|-----------|---------|--------------|
| *(none)* | Unit tests | Go only |
| `integration` | Integration tests | Docker, vault, oci-registry |
| `unix` | Unix-specific tests | Linux / macOS |

`make test` runs with the `integration` tag. `make unit-test` runs without tags.

## Linting

- **Linter**: golangci-lint, configured in `.github/config/golangci.yaml`

## Code Generation

```bash
make generate            # go generate
make generate-deepcopy   # controller-gen for api/ocm/compdesc/
```

Run these before committing if you change types or CLI commands.

## Pull Requests

PR titles must follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>(<optional scope>): <description>
```

Allowed types: `feat`, `fix`, `chore`, `docs`, `test`, `perf`

## Questions?

- [Project issues](https://github.com/open-component-model/ocm-project/issues)
- [Repository issues](https://github.com/open-component-model/ocm/issues)
- [Community engagement](https://ocm.software/community/engagement/)
- [NeoNephos Code of Conduct](https://github.com/neonephos/.github/blob/main/CODE_OF_CONDUCT.md)
