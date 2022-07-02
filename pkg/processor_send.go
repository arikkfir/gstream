package pyper

import (
	"context"
	"gopkg.in/yaml.v3"
)

func NewChannelSendProcessor(c chan *yaml.Node) (PipeProcessor, error) {
	return func(ctx context.Context, n *yaml.Node) error {
		c <- n
		return nil
	}, nil
}

func MustChannelSendProcessor(c chan *yaml.Node) PipeProcessor {
	processor, err := NewChannelSendProcessor(c)
	if err != nil {
		panic(err)
	}
	return processor
}
