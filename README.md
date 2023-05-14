# GoCovRpt

A `go` coverage reporter.

```text
gocovrpt is a CLI application that creates code coverage reports in multiple
formats like HTML, JSON, XML, TEXT, etc. with convenience options for generating
summaries, badges, and an isolated value useful in CI/CD control.

The input file MUST always be the last argument, and can support multiples separated by a space.

Usage:
  gocovrpt [flags]

Examples:
  $ gocovrpt -f html -l [full|summary] -o ./coverage -i ./.build/coverage.raw
  $ gocovrpt -f json -l [full|summary] -o ./coverage.json -i ./.build/coverage.raw
  $ gocovrpt -f xml -l [full|summary] -o ./coverage.xml -i ./.build/coverage.raw
  $ gocovrpt -f badge -o ./coverage.svg -i ./.build/coverage.raw
  $ gocovrpt -f value -o ./covered -i ./.build/coverage.raw

Flags:
  -f, --format string       Report format. Available formats: html, badge, value (default "html")
  -h, --help                help for gocovrpt
  -i, --input stringArray   One or more coverage.raw files to read from. (default [./.build/coverage.raw])
  -l, --level string        Report level. Available levels: full, summary (default "full")
  -o, --output string       Output file or directory. For badges, the default is ./.build/coverage.svg. (default "./.build/coverage")
  -p, --project string      The name of the project.
  -s, --source string       The directory containing the covered source files. (default $PWD)
```
