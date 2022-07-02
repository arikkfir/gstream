package pyper

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
)

func NewReaderInput(r io.Reader) (PipeInput, error) {
	return func(_ context.Context, target chan *yaml.Node) error {
		decoder := yaml.NewDecoder(r)
		for {
			doc := yaml.Node{}
			if err := decoder.Decode(&doc); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				} else {
					return fmt.Errorf("failed parsing input: %w", err)
				}
			} else {
				target <- &doc
			}
		}
	}, nil
}

func MustReaderInput(r io.Reader) PipeInput {
	if input, err := NewReaderInput(r); err != nil {
		panic(err)
	} else {
		return input
	}
}
