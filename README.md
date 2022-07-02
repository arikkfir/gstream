# pyper

![Maintainer](https://img.shields.io/badge/maintainer-arikkfir-blue)
![GoVersion](https://img.shields.io/github/go-mod/go-version/arikkfir/pyper.svg)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/arikkfir/pyper)
[![GoReportCard](https://goreportcard.com/badge/github.com/arikkfir/pyper)](https://goreportcard.com/report/github.com/arikkfir/pyper)
[![codecov](https://codecov.io/gh/arikkfir/pyper/branch/main/graph/badge.svg?token=QP3OAILB25)](https://codecov.io/gh/arikkfir/pyper)

Pipeline for YAML nodes in Golang.

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
	"github.com/arikkfir/pyper/pkg"
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
	input := pyper.MustReaderInput(strings.NewReader(yml))
	nodes := make(chan *yaml.Node, 1000)
	processor := pyper.MustFilterProcessor(
		"$[?(@.include==true)]",
		pyper.MustChannelSendProcessor(nodes),
	)
	if err := pyper.Pipe(context.Background(), []pyper.PipeInput{input}, processor); err != nil {
		panic(err)
	}
	close(nodes)
	if err := pyper.EncodeYAMLFromChannel(nodes, 2, os.Stderr); err != nil {
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
