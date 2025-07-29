# panicscan

panicscan is a static analysis engine for Go, focused on accurately detecting runtime panics while minimizing false positives and supporting modern Go concurrency patterns.

- Scans Go projects (across multiple packages) and reports possible panics.
- Handles nil dereference, unsafe map use, unsafe channel operations, unsafe type assertions, and explicit panics.
- Intelligent data flow reasoning and concurrency pattern recognition.
- Excludes user-specified directories (tests, mocks, vendor, etc).

## Usage