# panicscan

`panicscan` is a sophisticated static analysis tool for the Go programming language. Its primary purpose is to detect a wide range of potential runtime panics by analyzing Go source code, helping developers write more robust and reliable applications.

## Features

-   **Comprehensive Panic Detection**: Identifies numerous panic categories including:
    -   Explicit `panic()` calls
    -   Integer division by zero
    -   Nil pointer dereferences
    -   Writes to nil maps and sends to nil channels
    -   Operations on closed channels
    -   Out-of-bounds access on fixed-size arrays
    -   Calls to nil function variables
-   **Intelligent Analysis**: Utilizes data flow analysis and state tracking to understand variable states (`nil`, `non-nil`, `closed`), significantly reducing false positives.
-   **Concurrency Aware**: Intelligently handles common concurrency patterns to avoid flagging safe, idiomatic code.
-   **Project-Wide Scanning**: Correctly loads and analyzes entire Go projects, including those with multiple packages and complex dependencies, supporting the `./...` pattern.
-   **Configurable**: Allows excluding specific directories from the scan.

## Installation

```sh
go install [github.com/user/panicscan/cmd/panicscan@latest](https://github.com/user/panicscan/cmd/panicscan@latest)