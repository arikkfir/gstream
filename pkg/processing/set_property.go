package processing

import (
	"context"
	"fmt"
	"github.com/arikkfir/gstream/pkg/types"
	"gopkg.in/yaml.v3"
	"strconv"
)

func SetProperty(name string, value interface{}) types.NodeProcessor {
	return func(ctx context.Context, node *yaml.Node) error {
		var strValue, tag string
		switch typedValue := value.(type) {
		case int:
			strValue = strconv.Itoa(typedValue)
			tag = "!!int"
		case int8:
			strValue = strconv.Itoa(int(typedValue))
			tag = "!!int"
		case int16:
			strValue = strconv.Itoa(int(typedValue))
			tag = "!!int"
		case int32:
			strValue = strconv.Itoa(int(typedValue))
			tag = "!!int"
		case int64:
			strValue = strconv.Itoa(int(typedValue))
			tag = "!!int"
		case string:
			strValue = typedValue
			tag = "!!str"
		case bool:
			strValue = strconv.FormatBool(typedValue)
			tag = "!!bool"
		default:
			return fmt.Errorf("unsupported type %T", typedValue)
		}
		node.Content = append(
			node.Content,
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: name,
			},
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   tag,
				Value: strValue,
			},
		)
		return nil
	}
}
