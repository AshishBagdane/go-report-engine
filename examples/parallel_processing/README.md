# Parallel Processing Example

This example demonstrates how to use parallel processing in the go-report-engine to achieve significant performance improvements on large datasets.

## Overview

The parallel processing feature allows you to process data concurrently across multiple CPU cores, which is especially beneficial for:

- **CPU-intensive operations** (complex filtering, validation, transformations)
- **Large datasets** (thousands to millions of records)
- **Multi-core systems** (automatic scaling to available CPU cores)

## Quick Start

```bash
cd examples/parallel_processing
go run main.go
```

## What You'll Learn

1. **Basic Parallel Processing** - Wrap any processor for concurrent execution
2. **Custom Configuration** - Fine-tune workers, chunk size, and other settings
3. **Type-Safe Registration** - Use parallel filters, validators, and transformers
4. **Performance Comparison** - Sequential vs parallel processing benchmarks
5. **Best Practices** - When and how to use parallel processing effectively

## Examples Included

### Example 1: Basic Parallel Processing

Demonstrates wrapping a base processor with default parallel configuration.

```go
parallelProc := processor.NewParallelProcessor(baseProcessor)
result, err := parallelProc.Process(ctx, data)
```

### Example 2: Configured Parallel Processing

Shows how to customize workers, chunk size, and minimum chunk size.

```go
config := processor.ParallelConfig{
    Workers:      16,
    ChunkSize:    200,
    MinChunkSize: 50,
}
parallelProc := processor.NewParallelProcessorWithConfig(baseProcessor, config)
```

### Example 3-5: Parallel Strategies

Demonstrates parallel versions of filters, validators, and transformers.

```go
registry.RegisterParallelFilter("parallel_filter", &MyFilter{})
registry.RegisterParallelValidator("parallel_validator", &MyValidator{})
registry.RegisterParallelTransformer("parallel_transformer", &MyTransformer{})
```

### Example 6: Performance Comparison

Benchmarks sequential vs parallel processing showing speedup achieved.

## Configuration

### Via Code

```go
// Default configuration (runtime.NumCPU() workers, auto chunk size)
parallel := processor.NewParallelProcessor(baseProcessor)

// Custom configuration
config := processor.ParallelConfig{
    Workers:      8,        // Number of concurrent workers
    ChunkSize:    500,      // Records per chunk (0 = auto)
    MinChunkSize: 100,      // Minimum records per chunk
}
parallel := processor.NewParallelProcessorWithConfig(baseProcessor, config)
```

### Via YAML Configuration

See `config.parallel.yaml` for a complete example:

```yaml
processors:
  - type: parallel_filter
    params:
      workers: "8"
      chunk_size: "500"
      min_chunk_size: "100"
      threshold: "50" # Filter-specific param
```

### Via Registry

```go
// Register with default config
registry.RegisterParallelFilter("my_parallel_filter", &MyFilter{})

// Register with custom config
config := processor.ParallelConfig{Workers: 16, ChunkSize: 500}
registry.RegisterParallelProcessorWithConfig("custom_parallel", config, factory)
```

## Configuration Parameters

| Parameter        | Type | Default            | Description                            |
| ---------------- | ---- | ------------------ | -------------------------------------- |
| `workers`        | int  | `runtime.NumCPU()` | Number of concurrent worker goroutines |
| `chunk_size`     | int  | Auto-calculated    | Records per chunk (0 for auto)         |
| `min_chunk_size` | int  | 100                | Minimum records per chunk              |

### Auto Chunk Size Calculation

When `chunk_size` is 0 (or not specified), it's calculated as:

```
chunk_size = max(data_size / (workers * 2), min_chunk_size)
```

This aims for `workers * 2` chunks to enable good load balancing.

## Performance Characteristics

### When to Use Parallel Processing

✅ **Good for:**

- CPU-intensive operations (complex calculations, validation logic)
- Large datasets (>1000 records)
- Independent record processing (no inter-record dependencies)
- Multi-core systems

❌ **Not ideal for:**

- Small datasets (<100 records) - overhead exceeds benefits
- I/O-bound operations (disk/network already concurrent)
- Operations with inter-record dependencies
- Single-core systems

### Expected Performance

Typical speedup on multi-core systems:

- **4 cores**: 2.5-3.5x faster
- **8 cores**: 4-6x faster
- **16 cores**: 6-10x faster

Actual speedup depends on:

- Dataset size
- Operation CPU intensity
- Memory bandwidth
- Goroutine scheduling overhead

## Small Dataset Optimization

The parallel processor automatically detects small datasets and falls back to sequential processing when:

```
data_size < min_chunk_size (default: 100)
```

This avoids the overhead of parallel processing on small datasets.

## Resource Management

### Cleanup

Always close parallel processors to release resources:

```go
parallel := processor.NewParallelProcessor(baseProcessor)
defer parallel.Close()

// Or with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
parallel.CloseWithContext(ctx)
```

### Context Cancellation

Parallel processing respects context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := parallel.Process(ctx, data)
if err == context.DeadlineExceeded {
    // Processing was canceled
}
```

## Best Practices

### 1. Profile Before Optimizing

```bash
# Run with profiling
go test -cpuprofile=cpu.prof -bench=.

# Analyze profile
go tool pprof cpu.prof
```

### 2. Tune Worker Count

```go
// CPU-bound: Use CPU count
Workers: runtime.NumCPU()

// With hyperthreading: Use 2x CPU count
Workers: runtime.NumCPU() * 2

// I/O-bound: Use higher count
Workers: runtime.NumCPU() * 4
```

### 3. Optimize Chunk Size

```go
// Small chunks: Better load balancing, more overhead
ChunkSize: 50

// Large chunks: Less overhead, potential imbalance
ChunkSize: 1000

// Auto: Let the system decide
ChunkSize: 0  // Recommended starting point
```

### 4. Monitor Memory Usage

Large chunk sizes can increase memory usage. Monitor with:

```bash
# Run with memory profiling
go test -memprofile=mem.prof -bench=.
```

## Integration with Engine

### In Processor Chain

```go
// Build chain with parallel processors
chain, err := factory.BuildProcessorChain([]engine.ProcessorConfig{
    {Type: "filter", Params: map[string]string{...}},
    {Type: "parallel_validator", Params: map[string]string{
        "workers": "8",
        "chunk_size": "500",
    }},
    {Type: "parallel_transformer", Params: map[string]string{
        "workers": "16",
    }},
})
```

### With Full Engine

```go
engine, err := config.LoadAndBuild("config.parallel.yaml")
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
if err := engine.RunWithContext(ctx); err != nil {
    log.Fatal(err)
}
```

## Benchmarking

Run benchmarks to measure performance:

```bash
# Run all benchmarks
go test ./internal/processor/... -bench=Parallel -benchmem

# Run specific benchmark
go test ./internal/processor/... -bench=BenchmarkParallelProcessorProcess_VariableWorkers

# Compare with sequential
go test ./internal/processor/... -bench=. | grep -E "(Sequential|Parallel)"
```

## Troubleshooting

### Poor Performance

**Problem**: Parallel processing slower than sequential

**Solutions**:

- Increase dataset size (below 1000 records may not benefit)
- Check CPU utilization (`top` or `htop`)
- Reduce chunk size for better load balancing
- Ensure operation is CPU-bound, not I/O-bound

### High Memory Usage

**Problem**: Memory consumption too high

**Solutions**:

- Reduce worker count
- Decrease chunk size
- Process data in batches
- Enable garbage collection tuning

### Context Deadline Exceeded

**Problem**: Processing times out

**Solutions**:

- Increase context deadline
- Optimize processor logic
- Reduce dataset size
- Check for blocking operations

## See Also

- [Main README](../../README.md)
- [Processor Documentation](../../internal/processor/)
- [Configuration Guide](../../docs/CONFIG.md)
- [Performance Guide](../../docs/PERFORMANCE.md)

## Running the Example

```bash
# From repository root
go run examples/parallel_processing/main.go

# With race detector
go run -race examples/parallel_processing/main.go

# With profiling
go run examples/parallel_processing/main.go -cpuprofile=cpu.prof
```
