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
