package stream

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	. "github.com/arikkfir/gstream/pkg/generate"
	. "github.com/arikkfir/gstream/pkg/processing"
	. "github.com/arikkfir/gstream/pkg/sink"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestSetProperty(t *testing.T) {
	yamlString := `apiVersion: v1
firstName: John
lastName: Doe
---
firstName: Jane
lastName: Doe
---
firstName: Charlie
lastName: Brown`
	w2 := &bytes.Buffer{}
	s := NewStream().
		Generate(FromReader(strings.NewReader(yamlString))).
		Process(SetProperty("processed", true)).
		Sink(ToWriter(w2))
	if err := s.Execute(context.Background()); err != nil {
		t.Error(fmt.Errorf("failed executing stream: %w", err))
	}

	expectedYAMLString := `apiVersion: v1
firstName: John
lastName: Doe
processed: true
---
firstName: Jane
lastName: Doe
processed: true
---
firstName: Charlie
lastName: Brown
processed: true`
	if actual, err := formatYAML(w2); err != nil {
		t.Fatalf("Failed to format YAML output: %s", err)
	} else if expected, err := formatYAML(strings.NewReader(expectedYAMLString)); err != nil {
		t.Errorf("Failed to format expected YAML output: %s", err)
	} else if strings.TrimSuffix(expected, "\n") != strings.TrimSuffix(actual, "\n") {
		edits := myers.ComputeEdits(span.URIFromPath("expected"), expected, actual)
		diff := fmt.Sprint(gotextdiff.ToUnified("expected", "actual", expected, edits))
		t.Errorf("Incorrect output:\n===\n%s\n===", diff)
	}
}

func TestScale(t *testing.T) {
	const nodeCount = 1_000_000

	// Target channel for nodes to be sent to, and YAML paths for validation
	c := make(chan *yaml.Node, 1_000_000)
	include1Path, err := yamlpath.NewPath("$.metadata.include1")
	if err != nil {
		t.Fatalf("Failed to create include1 path: %v", err)
	}
	include2Path, err := yamlpath.NewPath("$.metadata.include2")
	if err != nil {
		t.Fatalf("Failed to create include2 path: %v", err)
	}
	fooAnnPath, err := yamlpath.NewPath("$.foo")
	if err != nil {
		t.Fatalf("Failed to create foo annotation path: %v", err)
	}

	// Start a validation thread to verify only selected nodes were sent
	done := make(chan int)
	go func() {
		count := 0
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in validation goroutine: %v", r)
				done <- -1
			} else {
				done <- count
			}
		}()
		errors := 0
		for {
			node, ok := <-c
			if !ok {
				return
			} else if matches, err := include1Path.Find(node); err != nil {
				t.Errorf("Failed to inspect for include1 path: %v", err)
				errors++
			} else if len(matches) != 1 {
				t.Errorf("Expected 1 match of 'include1' property, got %d", len(matches))
				errors++
			} else if n := matches[0]; n.Kind != yaml.ScalarNode {
				t.Errorf("Expected scalar match of 'include1' property, got %d", n.Kind)
				errors++
			} else if n.Value != "true" {
				t.Errorf("Node with 'include1!=true' found!")
				errors++
			} else if matches, err := include2Path.Find(node); err != nil {
				t.Errorf("Failed to inspect for include2 path: %v", err)
				errors++
			} else if len(matches) != 1 {
				t.Errorf("Expected 1 match of 'include2' property, got %d", len(matches))
				errors++
			} else if n := matches[0]; n.Kind != yaml.ScalarNode {
				t.Errorf("Expected scalar match of 'include2' property, got %d", n.Kind)
				errors++
			} else if n.Value != "true" {
				t.Errorf("Node with 'include2!=true' found!")
				errors++
			} else if matches, err := fooAnnPath.Find(node); err != nil {
				t.Errorf("Failed to inspect for foo annotation path: %v", err)
				errors++
			} else if len(matches) != 1 {
				t.Errorf("Expected 1 match of 'foo' annotation, got %d", len(matches))
				errors++
			} else if n := matches[0]; n.Kind != yaml.ScalarNode {
				t.Errorf("Expected scalar match of 'foo' annotation, got %d", n.Kind)
				errors++
			} else if n.Value != "bar" {
				t.Errorf("Node with 'annotations.foo!=bar' found!")
				errors++
			} else {
				count++
			}
			if errors >= 10 {
				t.Logf("Stopped after %d errors", errors)
				break
			}
		}
	}()

	s := NewStream().
		Generate(func(ctx context.Context, target chan *yaml.Node) error {
			for i := 0; i < nodeCount; i++ {
				include1 := i%2 == 0
				include2 := i%3 == 0
				yml := `apiVersion: v1
kind: ServiceAccount
metadata:
  include1: ` + strconv.FormatBool(include1) + `
  include2: ` + strconv.FormatBool(include2) + `
  namespace: ns1
  name: sa` + strconv.Itoa(i)
				doc := yaml.Node{}
				if err := yaml.Unmarshal([]byte(yml), &doc); err != nil {
					return fmt.Errorf("failed unmarshalling YAML into node: %w", err)
				}
				target <- &doc
			}
			return nil
		}).
		Transform(YAMLPathFilter("$[?(@.metadata.include1==true)]")).
		Transform(YAMLPathFilter("$[?(@.metadata.include2==true)]")).
		Process(SetProperty("foo", "bar")).
		Sink(ToChannel(c))
	if err := s.Execute(context.Background()); err != nil {
		t.Errorf("failed executing pipeline: %v", err)
	}
	close(c)

	// Wait for validation to finish
	count := <-done
	if count != 166667 {
		t.Errorf("Expected 166667 nodes, got %d", count)
	}
}

func formatYAML(yamlString io.Reader) (string, error) {
	const formatYAMLFailureMessage = `%s: %w
======
%s
======`

	formatted := bytes.Buffer{}
	decoder := yaml.NewDecoder(yamlString)
	encoder := yaml.NewEncoder(&formatted)
	encoder.SetIndent(2)
	for {
		var data interface{}
		if err := decoder.Decode(&data); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf(formatYAMLFailureMessage, "failed decoding YAML", err, yamlString)
		} else if err := encoder.Encode(data); err != nil {
			return "", fmt.Errorf(formatYAMLFailureMessage, "failed encoding struct", err, yamlString)
		}
	}
	encoder.Close()
	return formatted.String(), nil
}
