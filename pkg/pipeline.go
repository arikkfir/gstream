package pyper

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"sync"
)

const (
	inputNodesBufferSize = 100
	processingWorkersNum = 100
)

type PipeInput func(context.Context, chan *yaml.Node) error

type PipeProcessor func(context.Context, *yaml.Node) error

func Pipe(ctx context.Context, inputs []PipeInput, processor PipeProcessor) error {
	nodes := make(chan *yaml.Node, inputNodesBufferSize)

	// must have enough space in exit channel for theoretically all goroutines to fail
	// Each goroutine, if it fails, will emit its error into this channel. Since the goroutines are running in parallel,
	// and since theoretically more than one goroutine can fail, we must make sure this channel has enough room for all
	// of them to send an error in order for them to exit (even if in failure).
	goroutinesCount := len(inputs) + 1 + processingWorkersNum
	exitCh := make(chan error, goroutinesCount)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// GENERATE NODES
	// --------------
	// Invoke each input function in a separate goroutine, and let it push nodes into our nodes channel.
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	inputsWG := &sync.WaitGroup{}
	for index, input := range inputs {
		inputsWG.Add(1)
		go func(index int, input PipeInput) {
			defer inputsWG.Done()
			if err := input(ctx, nodes); err != nil {
				exitCh <- fmt.Errorf("input %d failed: %w", index, err)
			}
		}(index, input)
	}

	// Create a wait-group that tracks all root goroutines that must exit for the pipeline to finish.
	wg := &sync.WaitGroup{}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// CLOSE NODES CHANNEL WHEN INPUTS ARE DONE
	// ----------------------------------------
	// After all input functions goroutines exit, we should close the nodes channel, to make sure downstream readers
	// can finish.
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	wg.Add(1)
	go func() {
		defer wg.Done()
		inputsWG.Wait()
		close(nodes)
	}()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// PROCESS NODES
	// -------------
	// Create processing worker goroutines, each one creating its own processor and consumes nodes and processes them.
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	for i := 0; i < processingWorkersNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				node, ok := <-nodes
				if !ok {
					// no more nodes
					return
				}

				if err := processor(ctx, node); err != nil {
					exitCh <- fmt.Errorf("processor failed: %w", err)
					return
				}
			}
		}()
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// CONCLUDE
	// --------
	// Wait for all goroutines to finish (even if they fail, they should finish) and then check if an error occurred or
	// not. If one did, that error is returned.
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	wg.Wait()
	select {
	case err := <-exitCh:
		return err
	default:
		return nil
	}
}
