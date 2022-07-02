package pyper

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
)

func EncodeYAMLFromChannel(c chan *yaml.Node, spacesIndent int, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(spacesIndent)
	for {
		node, ok := <-c
		if !ok {
			return nil
		}

		if err := encoder.Encode(node); err != nil {
			return fmt.Errorf("failed to encode node: %w", err)
		}
	}
}
