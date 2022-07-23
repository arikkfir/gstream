package processing

import (
	"context"
	"github.com/arikkfir/gstream/internal"
	. "github.com/arikkfir/gstream/pkg/types"
	"gopkg.in/yaml.v3"
)

func AnnotateK8sResource(name string, value interface{}) NodeProcessor {
	return func(ctx context.Context, n *yaml.Node) error {
		return internal.SetAnnotation(n, name, value)
	}
}
