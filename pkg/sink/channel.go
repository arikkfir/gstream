package sink

import (
	"context"
	. "github.com/arikkfir/gstream/pkg/types"
	"gopkg.in/yaml.v3"
)

type channelNodeSink struct {
	c chan *yaml.Node
}

func (s *channelNodeSink) Process(_ context.Context, node *yaml.Node) error {
	s.c <- node
	return nil
}

func (s *channelNodeSink) Close() error {
	return nil
}

func ToChannel(c chan *yaml.Node) NodeSink {
	return &channelNodeSink{c}
}
