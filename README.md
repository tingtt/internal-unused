# internal-unused

`internalunused` detects unused exported declarations in the `internal` packages of a Go module.

Go's `internal` import restriction means every caller must live inside the same module, so a module-wide static analysis can determine with confidence whether an exported declaration is actually used.

## Install

```bash
go install github.com/tingtt/internal-unused/cmd/internalunused@latest
```

## Quickstart

```bash
# Analyse the module in the current directory (reports all unused declarations)
internalunused

# Also report declarations that are only referenced from test code
internalunused -mode=production

# Analyse a module in another directory
internalunused -dir /path/to/mymodule
```

### Example output

```text
internal/parser/parser.go:12:6: exported function example.com/mymodule/internal/parser.Parse is unused
internal/parser/parser.go:34:6: exported function example.com/mymodule/internal/parser.ParseBytes is only used by tests
```

### Exit codes

| Code | Meaning                                                 |
| ---- | ------------------------------------------------------- |
| `0`  | No unused declarations found                            |
| `1`  | One or more unused declarations detected (lint failure) |
| `2`  | Analysis could not be completed (execution failure)     |

## Detection modes

| Flag                  | Behaviour                                                       |
| --------------------- | --------------------------------------------------------------- |
| `-mode=all` (default) | Report declarations unused in **both** production and test code |
| `-mode=production`    | Also report declarations used **only** by test code             |

## What is detected

Exported declarations inside any `internal/` subtree of the module:

- Package-level functions
- Named types (including type aliases)
- Package-level variables and constants
- Methods on named types
- Exported struct fields
- Explicit interface methods

### What is not detected

- Declarations outside `internal/` packages
- Usage through `reflect`, struct tags, JSON/XML serialisation, or other runtime mechanisms
- Declarations in modules outside the target module

## Usage in CI

```yaml
# GitHub Actions example
- run: go install github.com/tingtt/internal-unused/cmd/internalunused@latest
- run: internalunused -mode=production
```

`internalunused` exits with code `1` on findings and `0` when clean, making it a drop-in lint step.

## Links

- [DesignDoc](DesignDoc.md)
