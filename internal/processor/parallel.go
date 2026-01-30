package processor

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/AshishBagdane/go-report-engine/pkg/api"
)

// ParallelProcessor wraps any ProcessorHandler to enable concurrent processing.
// It splits input data into chunks and processes them in parallel using a worker pool.
//
// Use cases:
//   - Large datasets that can be processed independently
//   - CPU-bound transformations (filtering, validation, computation)
//   - Multi-core utilization for improved throughput
//
// Limitations:
//   - Records must be processable independently (no inter-record dependencies)
//   - Order is preserved but processing is concurrent
//   - Memory overhead from chunking and buffering
//
// Configuration:
//   - Workers: Number of concurrent goroutines (default: runtime.NumCPU())
//   - ChunkSize: Records per chunk (default: auto-calculated)
//   - MinChunkSize: Minimum records per chunk (default: 100)
//
// Thread-safe: Yes. Can process multiple datasets concurrently.
//
// Resource Management:
// Implements api.CloseableWithContext for automatic cleanup integration.
// The underlying processor and worker pool are properly closed.
//
// Example:
//
//	// Wrap a filter for parallel execution
//	filter := NewFilterWrapper(myFilterStrategy)
//	parallel := NewParallelProcessor(filter)
//	defer parallel.Close()
//
//	// Configure concurrency
//	parallel.Configure(map[string]string{
//	    "workers": "8",
//	    "chunk_size": "1000",
//	})
//
//	// Process data in parallel
//	result, err := parallel.Process(ctx, largeDataset)
type ParallelProcessor struct {
	BaseProcessor

	// processor is the wrapped processor to execute in parallel
	processor ProcessorHandler

	// pool manages worker goroutines
	pool *WorkerPool

	// workers is the number of concurrent workers
	workers int

	// chunkSize is the number of records per chunk
	chunkSize int

	// minChunkSize is the minimum records per chunk
	minChunkSize int

	// closeOnce ensures Close is called only once
	closeOnce sync.Once

	// mu protects configuration changes
	mu sync.RWMutex
}

// ParallelConfig holds configuration for parallel processing.
type ParallelConfig struct {
	// Workers is the number of concurrent workers
	// Default: runtime.NumCPU()
	// Set to 0 for auto-detection
	Workers int

	// ChunkSize is the number of records per chunk
	// Default: auto-calculated based on data size
	// Set to 0 for auto-calculation
	ChunkSize int

	// MinChunkSize is the minimum records per chunk
	// Default: 100
	// Prevents over-chunking small datasets
	MinChunkSize int

	// OrderedResults determines if results maintain input order
	// Default: true (always preserved)
	// Note: Currently always true, may be configurable in future
	OrderedResults bool
}

// DefaultParallelConfig returns the default parallel processing configuration.
func DefaultParallelConfig() ParallelConfig {
	return ParallelConfig{
		Workers:        runtime.NumCPU(),
		ChunkSize:      0, // Auto-calculate
		MinChunkSize:   100,
		OrderedResults: true,
	}
}

// NewParallelProcessor creates a new parallel processor wrapping the given processor.
// It uses default configuration which can be customized via Configure().
//
// The wrapped processor is executed concurrently on data chunks.
// Results are reassembled in original order.
//
// Parameters:
//   - processor: The processor to execute in parallel (must not be nil)
//
// Panics if processor is nil.
//
// Example:
//
//	validator := NewValidatorWrapper(myValidator)
//	parallel := NewParallelProcessor(validator)
//	defer parallel.Close()
func NewParallelProcessor(processor ProcessorHandler) *ParallelProcessor {
	if processor == nil {
		panic("parallel processor: wrapped processor cannot be nil")
	}

	config := DefaultParallelConfig()

	return &ParallelProcessor{
		processor:    processor,
		workers:      config.Workers,
		chunkSize:    config.ChunkSize,
		minChunkSize: config.MinChunkSize,
		pool:         NewWorkerPool(config.Workers),
	}
}

// NewParallelProcessorWithConfig creates a parallel processor with custom configuration.
//
// Parameters:
//   - processor: The processor to execute in parallel (must not be nil)
//   - config: Configuration for parallel execution
//
// Panics if processor is nil or config is invalid.
//
// Example:
//
//	config := ParallelConfig{
//	    Workers: 16,
//	    ChunkSize: 500,
//	    MinChunkSize: 50,
//	}
//	parallel := NewParallelProcessorWithConfig(filter, config)
func NewParallelProcessorWithConfig(processor ProcessorHandler, config ParallelConfig) *ParallelProcessor {
	if processor == nil {
		panic("parallel processor: wrapped processor cannot be nil")
	}

	// Validate and apply defaults
	if config.Workers <= 0 {
		config.Workers = runtime.NumCPU()
	}
	if config.MinChunkSize <= 0 {
		config.MinChunkSize = 100
	}

	return &ParallelProcessor{
		processor:    processor,
		workers:      config.Workers,
		chunkSize:    config.ChunkSize,
		minChunkSize: config.MinChunkSize,
		pool:         NewWorkerPool(config.Workers),
	}
}

// Configure implements api.Configurable for runtime configuration.
// Supported parameters:
//   - "workers": Number of concurrent workers (integer)
//   - "chunk_size": Records per chunk (integer, 0 for auto)
//   - "min_chunk_size": Minimum records per chunk (integer)
//
// Configuration changes require recreating the worker pool.
// Should be called before Process() for best results.
//
// Parameters:
//   - params: Configuration parameters
//
// Returns:
//   - error: Validation error if parameters are invalid
//
// Example:
//
//	err := parallel.Configure(map[string]string{
//	    "workers": "8",
//	    "chunk_size": "1000",
//	})
func (p *ParallelProcessor) Configure(params map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Parse workers
	if workers, ok := params["workers"]; ok {
		w, err := parseInt(workers)
		if err != nil {
			return fmt.Errorf("invalid workers parameter: %w", err)
		}
		if w <= 0 {
			return fmt.Errorf("workers must be > 0, got %d", w)
		}
		p.workers = w
	}

	// Parse chunk_size
	if chunkSize, ok := params["chunk_size"]; ok {
		cs, err := parseInt(chunkSize)
		if err != nil {
			return fmt.Errorf("invalid chunk_size parameter: %w", err)
		}
		if cs < 0 {
			return fmt.Errorf("chunk_size must be >= 0, got %d", cs)
		}
		p.chunkSize = cs
	}

	// Parse min_chunk_size
	if minChunkSize, ok := params["min_chunk_size"]; ok {
		mcs, err := parseInt(minChunkSize)
		if err != nil {
			return fmt.Errorf("invalid min_chunk_size parameter: %w", err)
		}
		if mcs <= 0 {
			return fmt.Errorf("min_chunk_size must be > 0, got %d", mcs)
		}
		p.minChunkSize = mcs
	}

	// Recreate worker pool with new worker count
	if p.pool != nil {
		p.pool.Close()
	}
	p.pool = NewWorkerPool(p.workers)

	// Pass configuration to wrapped processor if it's configurable
	if configurable, ok := p.processor.(api.Configurable); ok {
		return configurable.Configure(params)
	}

	return nil
}

// Process executes the wrapped processor in parallel on data chunks.
// Data is split into chunks, processed concurrently, and reassembled in order.
//
// Processing flow:
//  1. Check context and validate input
//  2. Determine optimal chunk size
//  3. Split data into chunks
//  4. Process chunks concurrently via worker pool
//  5. Reassemble results in original order
//  6. Pass to next processor in chain
//
// Context handling:
//   - Checks context before processing
//   - Propagates context to worker pool and processors
//   - Returns ctx.Err() if canceled during processing
//
// Small dataset optimization:
//   - Datasets below minChunkSize skip parallel processing
//   - Falls back to sequential processing for efficiency
//
// Error handling:
//   - First error stops all workers
//   - Partial results are discarded
//   - Error is wrapped with context
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - data: Input data to process
//
// Returns:
//   - []map[string]interface{}: Processed data in original order
//   - error: Processing error, ctx.Err(), or error from next processor
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	result, err := parallel.Process(ctx, largeDataset)
//	if err != nil {
//	    log.Fatalf("Parallel processing failed: %v", err)
//	}
func (p *ParallelProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Check context before processing
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Handle empty input
	if len(data) == 0 {
		return p.BaseProcessor.Process(ctx, data)
	}

	// Small dataset optimization - skip parallel processing
	// for datasets smaller than minimum chunk size
	if len(data) < p.minChunkSize {
		result, err := p.processor.Process(ctx, data)
		if err != nil {
			return nil, err
		}
		return p.BaseProcessor.Process(ctx, result)
	}

	// Calculate optimal chunk size
	p.mu.RLock()
	chunkSize := p.calculateChunkSize(len(data))
	p.mu.RUnlock()

	// Split data into chunks
	chunks := p.createChunks(data, chunkSize)

	// Process chunks in parallel
	results, err := p.pool.ProcessChunks(ctx, chunks, func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return p.processor.Process(ctx, chunk.Data)
	})

	if err != nil {
		return nil, fmt.Errorf("parallel processing failed at chunk: %w", err)
	}

	// Reassemble results in original order
	assembled := p.assembleResults(results)

	// Pass to next processor in chain
	return p.BaseProcessor.Process(ctx, assembled)
}

// calculateChunkSize determines the optimal chunk size for the dataset.
// It balances parallelism with overhead based on data size and worker count.
//
// Strategy:
//   - If chunkSize is explicitly set (> 0), use it
//   - Otherwise, distribute data evenly across workers
//   - Ensure chunks are not smaller than minChunkSize
//
// Must be called with p.mu held (read or write lock).
func (p *ParallelProcessor) calculateChunkSize(dataSize int) int {
	// Use explicit chunk size if configured
	if p.chunkSize > 0 {
		return p.chunkSize
	}

	// Auto-calculate: distribute evenly across workers
	// This aims for workers * 2 chunks for better load balancing
	targetChunks := p.workers * 2
	autoSize := dataSize / targetChunks

	// Ensure minimum chunk size to avoid excessive overhead
	if autoSize < p.minChunkSize {
		autoSize = p.minChunkSize
	}

	// Ensure at least one record per chunk
	if autoSize < 1 {
		autoSize = 1
	}

	return autoSize
}

// createChunks splits data into work chunks of the specified size.
// Each chunk maintains metadata about its position for ordered reassembly.
func (p *ParallelProcessor) createChunks(data []map[string]interface{}, chunkSize int) []WorkChunk {
	chunks := make([]WorkChunk, 0, (len(data)+chunkSize-1)/chunkSize)

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunks = append(chunks, WorkChunk{
			Data:  data[i:end],
			Index: len(chunks),
			Start: i,
			End:   end,
		})
	}

	return chunks
}

// assembleResults combines processed chunks back into a single slice.
// Results are already sorted by chunk index from the worker pool.
func (p *ParallelProcessor) assembleResults(results []WorkResult) []map[string]interface{} {
	// Pre-calculate total size to avoid reallocations
	totalSize := 0
	for _, result := range results {
		totalSize += len(result.Data)
	}

	// Assemble in order
	assembled := make([]map[string]interface{}, 0, totalSize)
	for _, result := range results {
		assembled = append(assembled, result.Data...)
	}

	return assembled
}

// Close closes the worker pool and underlying processor.
// This implements io.Closer for resource cleanup integration.
//
// Idempotent - multiple calls are safe.
//
// Example:
//
//	parallel := NewParallelProcessor(processor)
//	defer parallel.Close()
func (p *ParallelProcessor) Close() error {
	var poolErr, procErr error

	p.closeOnce.Do(func() {
		// Close worker pool
		if p.pool != nil {
			poolErr = p.pool.Close()
		}

		// Close wrapped processor if it implements Closer
		if closer, ok := p.processor.(interface{ Close() error }); ok {
			procErr = closer.Close()
		}
	})

	// Return first error encountered
	if poolErr != nil {
		return poolErr
	}
	return procErr
}

// CloseWithContext closes the processor with a timeout.
// This implements api.CloseableWithContext for engine integration.
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
//	if err := parallel.CloseWithContext(ctx); err != nil {
//	    log.Printf("Shutdown timeout: %v", err)
//	}
func (p *ParallelProcessor) CloseWithContext(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		done <- p.Close()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Workers returns the current worker count.
// Useful for diagnostics and metrics.
func (p *ParallelProcessor) Workers() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.workers
}

// ChunkSize returns the current chunk size configuration.
// Returns 0 if using auto-calculation.
func (p *ParallelProcessor) ChunkSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.chunkSize
}

// parseInt parses a string to int with error handling.
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
