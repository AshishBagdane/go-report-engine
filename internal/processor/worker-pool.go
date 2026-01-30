package processor

import (
	"context"
	"fmt"
	"sync"
)

// WorkerPool manages a pool of workers for concurrent data processing.
// It provides:
//   - Bounded concurrency to control resource usage
//   - Graceful shutdown with context cancellation
//   - Error propagation from workers
//   - Work distribution across multiple goroutines
//
// The pool processes chunks of data in parallel using a fixed number of workers.
// Each worker executes a task function on its assigned chunk.
//
// Thread-safe: Yes. All methods can be called concurrently.
//
// Resource Management:
// The pool should be closed when no longer needed to ensure goroutine cleanup.
// Use defer pool.Close() or implement CloseableWithContext for timeout support.
//
// Example:
//
//	pool := NewWorkerPool(4) // 4 concurrent workers
//	defer pool.Close()
//
//	chunks := []WorkChunk{...}
//	results, err := pool.ProcessChunks(ctx, chunks, func(ctx context.Context, chunk WorkChunk) (interface{}, error) {
//	    // Process chunk
//	    return processedData, nil
//	})
type WorkerPool struct {
	// workers is the number of concurrent workers
	workers int

	// closeOnce ensures Close is called only once
	closeOnce sync.Once

	// closed indicates if the pool has been closed
	closed bool

	// mu protects the closed flag
	mu sync.RWMutex
}

// WorkChunk represents a chunk of data to be processed by a worker.
// It contains the data slice and metadata about the chunk's position.
type WorkChunk struct {
	// Data is the slice of records to process
	Data []map[string]interface{}

	// Index is the chunk's position in the original dataset
	Index int

	// Start is the starting index in the original dataset
	Start int

	// End is the ending index in the original dataset
	End int
}

// WorkResult holds the result of processing a chunk.
// It maintains the chunk's original position for reassembly.
type WorkResult struct {
	// Data is the processed result
	Data []map[string]interface{}

	// Index is the chunk's position for ordered reassembly
	Index int

	// Error is any error that occurred during processing
	Error error
}

// TaskFunc is the function signature for worker tasks.
// It processes a single chunk and returns the result or error.
type TaskFunc func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error)

// NewWorkerPool creates a new worker pool with the specified number of workers.
// The workers parameter determines the maximum number of concurrent goroutines.
//
// Parameters:
//   - workers: Number of concurrent workers (must be > 0)
//
// Panics if workers <= 0.
//
// Example:
//
//	// Use runtime.NumCPU() for CPU-bound tasks
//	pool := NewWorkerPool(runtime.NumCPU())
//
//	// Use higher count for I/O-bound tasks
//	pool := NewWorkerPool(runtime.NumCPU() * 2)
func NewWorkerPool(workers int) *WorkerPool {
	if workers <= 0 {
		panic("worker pool: workers must be > 0")
	}

	return &WorkerPool{
		workers: workers,
	}
}

// ProcessChunks processes multiple chunks concurrently using the worker pool.
// Work is distributed across workers, and results are collected and returned
// in the original order.
//
// Context handling:
//   - Cancellation stops accepting new work immediately
//   - Active workers complete or abort based on task's context handling
//   - Returns ctx.Err() if context is canceled
//
// Error handling:
//   - First error encountered stops further processing
//   - Partial results up to error point are discarded
//   - All workers are signaled to stop via context cancellation
//
// Ordering:
//   - Results are reassembled in original chunk order
//   - Chunk.Index is used for ordering, not processing order
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - chunks: Work chunks to process
//   - task: Function to execute on each chunk
//
// Returns:
//   - []WorkResult: Results in original chunk order
//   - error: First error encountered, or ctx.Err() if canceled
//
// Example:
//
//	results, err := pool.ProcessChunks(ctx, chunks, func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
//	    return processor.Process(ctx, chunk.Data)
//	})
func (p *WorkerPool) ProcessChunks(ctx context.Context, chunks []WorkChunk, task TaskFunc) ([]WorkResult, error) {
	// Check if pool is closed
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("worker pool is closed")
	}
	p.mu.RUnlock()

	// Check initial context state
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(chunks) == 0 {
		return []WorkResult{}, nil
	}

	// Create buffered channels for work distribution and result collection
	workCh := make(chan WorkChunk, len(chunks))
	resultCh := make(chan WorkResult, len(chunks))

	// Context for canceling workers on first error
	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	// Determine actual worker count (don't spawn more workers than chunks)
	actualWorkers := p.workers
	if actualWorkers > len(chunks) {
		actualWorkers = len(chunks)
	}

	// Use local WaitGroup for this ProcessChunks call
	// This prevents race conditions when pool is used concurrently
	var wg sync.WaitGroup
	wg.Add(actualWorkers)
	for i := 0; i < actualWorkers; i++ {
		go p.worker(workerCtx, workCh, resultCh, task, &wg)
	}

	// Send chunks to workers
	// This is done in a separate goroutine to prevent blocking
	go func() {
		for _, chunk := range chunks {
			select {
			case workCh <- chunk:
			case <-workerCtx.Done():
				// Stop sending work if context is canceled
				close(workCh)
				return
			}
		}
		close(workCh)
	}()

	// Collect results
	results := make([]WorkResult, 0, len(chunks))
	var firstError error

	for i := 0; i < len(chunks); i++ {
		select {
		case result := <-resultCh:
			if result.Error != nil && firstError == nil {
				// Capture first error and cancel workers
				firstError = result.Error
				cancelWorkers()
			}
			results = append(results, result)

		case <-workerCtx.Done():
			// Workers canceled, drain any remaining results
			// This prevents deadlock when workers stop early due to error
			for len(results) < len(chunks) {
				select {
				case result := <-resultCh:
					results = append(results, result)
				default:
					// No more results available, exit
					goto done
				}
			}
			goto done

		case <-ctx.Done():
			// Parent context canceled
			cancelWorkers()
			return nil, ctx.Err()
		}
	}

done:
	// Wait for all workers to finish
	wg.Wait()

	// Check if parent context was canceled during processing
	// This ensures we return context.Canceled even if all results were collected
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Return error if any chunk failed
	if firstError != nil {
		return nil, firstError
	}

	// Sort results by original chunk index for ordered reassembly
	sortWorkResults(results)

	return results, nil
}

// worker is the goroutine function that processes chunks from the work channel.
// It runs until the work channel is closed or context is canceled.
func (p *WorkerPool) worker(ctx context.Context, workCh <-chan WorkChunk, resultCh chan<- WorkResult, task TaskFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case chunk, ok := <-workCh:
			if !ok {
				// Work channel closed, worker done
				return
			}

			// Check context before processing
			select {
			case <-ctx.Done():
				// Context canceled, send error result
				resultCh <- WorkResult{
					Index: chunk.Index,
					Error: ctx.Err(),
				}
				return
			default:
			}

			// Process chunk
			data, err := task(ctx, chunk)

			// Send result
			select {
			case resultCh <- WorkResult{
				Data:  data,
				Index: chunk.Index,
				Error: err,
			}:
			case <-ctx.Done():
				// Context canceled while sending result
				return
			}

		case <-ctx.Done():
			// Context canceled while waiting for work
			return
		}
	}
}

// Close shuts down the worker pool gracefully.
// It marks the pool as closed to reject new work.
//
// This method is idempotent - multiple calls are safe.
//
// Note: This does not wait for active ProcessChunks calls to complete.
// Use context cancellation for immediate termination of active work.
//
// Example:
//
//	pool := NewWorkerPool(4)
//	defer pool.Close()
func (p *WorkerPool) Close() error {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
	})
	return nil
}

// CloseWithContext closes the pool with a timeout.
// If the context expires before workers finish, it returns ctx.Err().
//
// This implements api.CloseableWithContext for integration with
// the engine's resource cleanup system.
//
// Parameters:
//   - ctx: Context with timeout or deadline
//
// Returns:
//   - error: nil if closed successfully, ctx.Err() if timeout
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := pool.CloseWithContext(ctx); err != nil {
//	    log.Printf("Pool shutdown timeout: %v", err)
//	}
func (p *WorkerPool) CloseWithContext(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		_ = p.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Timeout occurred - workers are still running
		// Mark as closed anyway to prevent new work
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
		return ctx.Err()
	}
}

// Workers returns the number of workers in the pool.
// This is useful for metrics and diagnostics.
func (p *WorkerPool) Workers() int {
	return p.workers
}

// IsClosed returns whether the pool has been closed.
// A closed pool will not accept new work.
func (p *WorkerPool) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// sortWorkResults sorts results by chunk index for ordered reassembly.
// This uses insertion sort which is efficient for small, mostly-sorted slices.
func sortWorkResults(results []WorkResult) {
	for i := 1; i < len(results); i++ {
		key := results[i]
		j := i - 1

		// Move elements greater than key one position ahead
		for j >= 0 && results[j].Index > key.Index {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = key
	}
}
