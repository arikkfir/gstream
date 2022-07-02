package pyper

import (
	"context"
	"fmt"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func NewFilterProcessor(expr string, next PipeProcessor) (PipeProcessor, error) {
	path, err := yamlpath.NewPath(expr)
	if err != nil {
		return nil, fmt.Errorf("failed creating filter processor for expression '%s': %w", expr, err)
	}

	return func(ctx context.Context, n *yaml.Node) error {
		matches, err := path.Find(n)
		if err != nil {
			return fmt.Errorf("filter '%s' failed: %w", expr, err)
		}

		for _, node := range matches {
			if err := next(ctx, node); err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func MustFilterProcessor(expr string, next PipeProcessor) PipeProcessor {
	processor, err := NewFilterProcessor(expr, next)
	if err != nil {
		panic(err)
	}
	return processor
}
