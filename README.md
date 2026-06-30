# yze-cliversion

A [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) analyzer in the [gomatic `yze`](https://github.com/gomatic/yze) suite. It enforces the gomatic CLI standard that a urfave/cli `package main` exposes `--version`: a command literal sets its `Version` field to the package-level `var version` (the ldflags `-X main.version` target).

- **Rule id:** `yze/cliversion`
- **Capability:** `convention:cliversion`

It checks the GO side only; the goreleaser `-X main.version` ldflag↔symbol match is a separate stickler-level YAML check. The rule is a package-level invariant — extra command literals (e.g. an inline subcommand with no `Version`) never cause a false positive.

## Use

```sh
go run github.com/gomatic/yze-cliversion/cmd/yze-cliversion@latest ./...
```
