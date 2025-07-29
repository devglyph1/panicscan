# PanicScan

PanicScan is a static analysis tool for Go that detects potential runtime panics in your code. It helps you identify and fix fragile code before it leads to crashes in production.

## Installation

To install `panicscan`, you can clone the repository and build it from source:

```bash
git clone [https://github.com/devglyph1/panicscan.git](https://github.com/devglyph1/panicscan.git)
cd panicscan
go build ./cmd/panicscan
```

## Usage

Run `panicscan` on your Go project's directory:

```bash
./panicscan /path/to/your/project
```

Or to check all subdirectories:

```bash
./panicscan ./...
```

The tool will scan all `.go` files and report any potential panics it finds. If no issues are found, it will exit with a success message. If panics are detected, it will print them and exit with a non-zero status code, making it suitable for use in CI/CD pipelines.

## Checks Implemented

`panicscan` currently checks for the following potential panics:

- Explicit `panic()` calls
- Integer division or modulo by zero
- Nil pointer dereferences
- Method calls on nil receivers
- Out-of-bounds access on arrays
- Out-of-bounds slicing on arrays
- Sending on a closed channel
- Closing a nil channel
- Closing an already-closed channel
- Writing to a nil map
- Single-value type assertions that can fail
- Calling a nil function variable


// .gitignore
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, build with `go test -c`
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work