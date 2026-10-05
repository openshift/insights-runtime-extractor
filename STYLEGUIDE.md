# Insights Runtime Extractor Styleguide

## Directory and File Names

- Use lowercase dash-separated names for all files (to avoid git issues with case-insensitive file systems)
- Exceptions are files which have their own naming conventions (e.g. Dockerfile, Makefile, README)
- Automation scripts should be stored into `dotdirs` (eg .githooks, .openshiftci)

## Rust

- All Rust code should be formatted by `cargo fmt`
- Code should pass `cargo clippy -- -D warnings` with no errors
- Run `make lint` (inside the rust-dev container) to check both
- Run `make fix` (inside the rust-dev container) to auto-fix Clippy and formatting issues

## Go

- All go code should be formatted by gofmt
- Import statement pkgs should be separated into 3 groups: stdlib, external dependency, current project.
- TESTS: Should follow the "test tables" convention.
- TESTS: Methods should follow the code style `Test_<FunctionName>` or `Test_<GatherName>_<FunctionName>`

### Recommendations

- **Comparing strings**, its is recommended that you use  `if string != ""` over `if len(string) > 0`, but both are acceptable.
