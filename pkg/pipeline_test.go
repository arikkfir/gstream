package pyper

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"strconv"
	"testing"
)

func TestPipeline(t *testing.T) {
	const nodeCount = 1_000_000

	// Create an Input that generates many nodes
	input := func(ctx context.Context, target chan *yaml.Node) error {
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
	}

	// Target channel for nodes to be sent to, and YAML paths for validation
	c := make(chan *yaml.Node, 1000)
	include1Path, err := yamlpath.NewPath("$.metadata.include1")
	if err != nil {
		t.Fatalf("Failed to create include1 path: %v", err)
	}
	include2Path, err := yamlpath.NewPath("$.metadata.include2")
	if err != nil {
		t.Fatalf("Failed to create include2 path: %v", err)
	}

	// Start a validation thread to verify only selected nodes were sent
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()
		for {
			node, ok := <-c
			if !ok {
				return
			} else if matches, err := include1Path.Find(node); err != nil {
				t.Errorf("Failed to inspect for include1 path: %v", err)
			} else if len(matches) != 1 {
				t.Errorf("Expected 1 match of 'include1' property, got %d", len(matches))
			} else if n := matches[0]; n.Kind != yaml.ScalarNode {
				t.Errorf("Expected scalar match of 'include1' property, got %d", n.Kind)
			} else if n.Value != "true" {
				t.Errorf("Node with 'include1!=true' found!")
			}
			if matches, err := include2Path.Find(node); err != nil {
				t.Errorf("Failed to inspect for include2 path: %v", err)
			} else if len(matches) != 1 {
				t.Errorf("Expected 1 match of 'include2' property, got %d", len(matches))
			} else if n := matches[0]; n.Kind != yaml.ScalarNode {
				t.Errorf("Expected scalar match of 'include2' property, got %d", n.Kind)
			} else if n.Value != "true" {
				t.Errorf("Node with 'include2!=true' found!")
			}
		}
	}()

	// Pipeline execution, reading all nodes from input, and sending to our channel only selected nodes
	processor := MustFilterProcessor(
		"$[?(@.metadata.include1==true)]",
		MustFilterProcessor(
			"$[?(@.metadata.include2==true)]",
			MustChannelSendProcessor(c),
		),
	)
	if err := Pipe(context.Background(), []PipeInput{input}, processor); err != nil {
		t.Errorf("failed executing pipeline: %v", err)
	}
	close(c)

	// Wait for validation to finish
	<-done
}
