# gstream

![Maintainer](https://img.shields.io/badge/maintainer-arikkfir-blue)
![GoVersion](https://img.shields.io/github/go-mod/go-version/arikkfir/gstream.svg)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/arikkfir/gstream)
[![GoReportCard](https://goreportcard.com/badge/github.com/arikkfir/gstream)](https://goreportcard.com/report/github.com/arikkfir/gstream)
[![codecov](https://codecov.io/gh/arikkfir/gstream/branch/main/graph/badge.svg?token=QP3OAILB25)](https://codecov.io/gh/arikkfir/gstream)

Golang YAML node pipeline.

## Status

This is currently alpha, with some features still in development, and not full test coverage. We're on it! 💪

## Example

This program does the following:
- Creates a pipeline input that just provides a YAML from a constant
  - More realistic examples would pipe YAML from a file or a stream
- Creates a pipeline of two steps:
  - Accepts only nodes whose `include` property equals `true`
  - Sends those nodes to a node channel
- Pipes nodes sent to the node channel into `stderr`

```go
package main

import (
	"context"
	. "github.com/arikkfir/gstream/pkg"
	. "github.com/arikkfir/gstream/pkg/generate"
	. "github.com/arikkfir/gstream/pkg/processing"
	. "github.com/arikkfir/gstream/pkg/sink"
	. "github.com/arikkfir/gstream/pkg/types"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const yml = `
key: value1
include: true
---
key: value2
include: false
---
key: value3
include: true`

func main() {
    s := NewStream().
        Generate(FromReader(strings.NewReader(yml))).
        Transform(YAMLPathFilter("$[?(@.include==true)]")).
        Sink(ToWriter(os.Stdout))
    if err := s.Execute(context.Background()); err != nil {
        panic(err)
    }
}
```

This will print the following:

```yaml
key: value1
include: true
---
key: value3
include: true
```
