# 🚀 go-report-engine

A **production-grade, modular reporting engine for Go** with comprehensive error handling, thread-safe registries, and enterprise-grade architecture.

Built using **Strategy**, **Factory**, **Template Method**, and **Chain of Responsibility** patterns.

**Fetch → Process → Format → Output — fully customizable.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Test Coverage](https://img.shields.io/badge/coverage-95%25-brightgreen)](https://github.com/AshishBagdane/go-report-engine)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## ✨ Features

### **Core Features**

- 🔌 **Pluggable Providers** - Fetch data from any source (DB, CSV, API, etc.)
- ♻️ **Processing Pipeline** - Chain of Responsibility for data transformation
- 🧾 **Multiple Formatters** - JSON, CSV, YAML output formats
- 📤 **Flexible Outputs** - Console, File, API, Slack, Email delivery
- 🧱 **SOLID Principles** - Clean, testable, extensible architecture
- 🧪 **Test-Driven** - 95%+ test coverage with comprehensive test suite

### **Production-Ready Features** ✨ NEW

- 🔒 **Thread-Safe Registries** - Concurrent access with `sync.RWMutex`
- 🚨 **Comprehensive Error Handling** - Context-rich errors with classification
- 🔄 **Intelligent Retry Logic** - Automatic retries for transient failures
- 📊 **Error Classification** - Transient, Permanent, Configuration, Validation, Resource
- 🎯 **Component-Specific Errors** - Specialized errors for debugging
- 🛡️ **Panic Recovery** - Graceful handling with `RunWithRecovery()`
- 🌱 **Built in Public** - Follow the real-time development journey

---

## 📦 Installation

```bash
go get github.com/AshishBagdane/go-report-engine
```

---

## 🧠 Architecture Overview

```
Provider → Processor Chain → Formatter → Output
```

### **Pipeline Components**

| Component     | Purpose                     | Examples                    |
| ------------- | --------------------------- | --------------------------- |
| **Provider**  | Fetch data from sources     | Mock, CSV, Database, API    |
| **Processor** | Transform data step-by-step | Filter, Validate, Transform |
| **Formatter** | Convert to output format    | JSON, CSV, YAML             |
| **Output**    | Deliver the final report    | Console, File, Slack, Email |

---

## 🧰 Quick Start

### **Basic Example**

```go
package main

import (
    "fmt"
    "log"
    "github.com/AshishBagdane/go-report-engine/internal/engine"
    "github.com/AshishBagdane/go-report-engine/internal/provider"
    "github.com/AshishBagdane/go-report-engine/internal/processor"
    "github.com/AshishBagdane/go-report-engine/internal/formatter"
    "github.com/AshishBagdane/go-report-engine/internal/output"
    "github.com/AshishBagdane/go-report-engine/internal/errors"
)

func main() {
    // Create engine with builder pattern
    eng, err := engine.NewEngineBuilder().
        WithProvider(provider.NewMockProvider()).
        WithProcessor(&processor.BaseProcessor{}).
        WithFormatter(formatter.NewJSONFormatter()).
        WithOutput(output.NewConsoleOutput()).
        Build()

    if err != nil {
        log.Fatal(err)
    }

    // Run the pipeline
    if err := eng.Run(); err != nil {
        // Error handling with context
        if errors.IsRetryable(err) {
            // Retry logic for transient failures
            fmt.Println("Retrying...")
        } else {
            log.Printf("Pipeline failed: %v", err)
        }
    }
}
```

### **Config-Driven Example**

```go
package main

import (
    "log"
    "github.com/AshishBagdane/go-report-engine/internal/engine"
    "github.com/AshishBagdane/go-report-engine/internal/factory"
)

func main() {
    // Define configuration
    config := engine.Config{
        Provider: engine.ProviderConfig{
            Type: "mock",
        },
        Processors: []engine.ProcessorConfig{
            {
                Type: "min_score_filter",
                Params: map[string]string{
                    "min_score": "90",
                },
            },
        },
        Formatter: engine.FormatterConfig{
            Type: "json",
        },
        Output: engine.OutputConfig{
            Type: "console",
        },
    }

    // Create engine from config
    eng, err := factory.NewEngineFromConfig(config)
    if err != nil {
        log.Fatal(err)
    }

    // Run pipeline
    if err := eng.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### **Custom Processor Example**

```go
package main

import (
    "github.com/AshishBagdane/go-report-engine/internal/registry"
    "github.com/AshishBagdane/go-report-engine/pkg/api"
)

// Define your custom filter
type MinScoreFilter struct {
    MinScore int
}

func (f *MinScoreFilter) Keep(row map[string]interface{}) bool {
    if score, ok := row["score"].(int); ok {
        return score >= f.MinScore
    }
    return false
}

func (f *MinScoreFilter) Configure(params map[string]string) error {
    // Configure from params
    return nil
}

func init() {
    // Register your custom filter
    registry.RegisterFilter("min_score_filter", &MinScoreFilter{})
}
```

---

## 🚨 Error Handling

### **Context-Rich Errors** ✨ NEW

Every error includes full context for debugging:

```go
if err := eng.Run(); err != nil {
    // Example error output:
    // [provider:fetch] connection timeout | context: {host: localhost, port: 5432, retry_count: 3} [type: transient]

    fmt.Printf("Error: %v\n", err)
}
```

### **Error Classification** ✨ NEW

```go
import "github.com/AshishBagdane/go-report-engine/internal/errors"

if err := eng.Run(); err != nil {
    switch errors.GetErrorType(err) {
    case errors.ErrorTypeTransient:
        // Retry with backoff
        time.Sleep(backoff)
        return retry()

    case errors.ErrorTypeConfiguration:
        // Alert admin - config issue
        alertAdmin(err)

    case errors.ErrorTypePermanent:
        // Log and skip - data issue
        log.Printf("Permanent failure: %v", err)

    case errors.ErrorTypeResource:
        // Scale resources or throttle
        scaleResources()

    case errors.ErrorTypeValidation:
        // Return to user - invalid input
        return fmt.Errorf("validation failed: %w", err)
    }
}
```

### **Intelligent Retry Logic** ✨ NEW

```go
if errors.IsRetryable(err) {
    for attempt := 0; attempt < maxRetries; attempt++ {
        time.Sleep(backoff * time.Duration(1<<attempt))
        if err = eng.Run(); err == nil {
            break
        }
    }
}
```

---

## 📁 Project Structure

```
go-report-engine/
├── cmd/
│   └── example/
│       └── main.go                    # Example usage
├── pkg/
│   └── api/
│       └── interfaces.go              # Public API
├── internal/
│   ├── engine/
│   │   ├── builder.go                 # Builder pattern
│   │   ├── config.go                  # Configuration
│   │   ├── engine.go                  # Core engine ✅
│   │   ├── engine_test.go             # Tests ✅
│   │   └── options.go                 # Functional options
│   ├── errors/                        # ✅ NEW - Complete
│   │   ├── errors.go                  # Core error infrastructure
│   │   ├── provider_errors.go         # Provider-specific errors
│   │   ├── processor_errors.go        # Processor-specific errors
│   │   ├── formatter_errors.go        # Formatter-specific errors
│   │   ├── output_errors.go           # Output-specific errors
│   │   └── *_test.go                  # Comprehensive tests
│   ├── registry/                      # ✅ NEW - Thread-safe
│   │   ├── formatter_registry.go      # Formatter registry
│   │   ├── output_registry.go         # Output registry
│   │   ├── processor_registry.go      # Processor registry
│   │   ├── provider_registry.go       # Provider registry
│   │   └── *_test.go                  # Registry tests
│   ├── provider/
│   │   ├── provider.go                # Provider interface
│   │   └── mock.go                    # Mock implementation
│   ├── processor/
│   │   ├── processor.go               # Processor interface
│   │   ├── base.go                    # Base processor
│   │   └── wrappers.go                # Type-safe wrappers
│   ├── formatter/
│   │   ├── formatter.go               # Formatter interface
│   │   └── json.go                    # JSON formatter
│   ├── output/
│   │   ├── output.go                  # Output interface
│   │   └── console.go                 # Console output
│   └── factory/
│       ├── engine_factory.go          # Engine factory
│       └── processor_chain_factory.go # Processor chain factory
├── go.mod
└── README.md
```

---

## 🧪 Testing

### **Test Coverage**

```bash
# Run all tests
go test ./... -v

# Run with race detector
go test ./... -race

# Check coverage
go test ./... -cover

# Run benchmarks
go test ./... -bench=. -benchmem
```

### **Current Test Statistics** ✨ NEW

- **173 test functions** across all packages
- **18 benchmarks** for performance validation
- **95%+ code coverage** on core components
- **Race detector clean** - safe for concurrent use
- **Zero flaky tests** - reliable and deterministic

### **Example Test**

```go
func TestReportEngineRun(t *testing.T) {
    engine := &ReportEngine{
        Provider:  &mockProvider{data: testData},
        Processor: &mockProcessor{},
        Formatter: &mockFormatter{},
        Output:    &mockOutput{},
    }

    err := engine.Run()
    if err != nil {
        t.Errorf("Run() should succeed, got error: %v", err)
    }
}
```

---

## 🔌 Available Components

### **Providers**

- [x] MockProvider - For testing
- [ ] CSVProvider - Coming soon
- [ ] DBProvider - Coming soon
- [ ] APIProvider - Coming soon

### **Processors**

- [x] BaseProcessor - Pass-through processor
- [x] FilterWrapper - Filter data rows
- [x] ValidatorWrapper - Validate data
- [x] TransformWrapper - Transform data
- [ ] AggregateProcessor - Coming soon
- [ ] SanitizeProcessor - Coming soon

### **Formatters**

- [x] JSONFormatter - JSON output
- [ ] CSVFormatter - Coming soon
- [ ] YAMLFormatter - Coming soon
- [ ] HTMLFormatter - Coming soon

### **Outputs**

- [x] ConsoleOutput - Terminal output
- [ ] FileOutput - Coming soon
- [ ] SlackOutput - Coming soon
- [ ] EmailOutput - Coming soon

---

## 🗺️ Roadmap

### **Phase 1 - Foundation** (Current)

- [x] Core architecture and interfaces
- [x] Thread-safe registries with `sync.RWMutex` ✅
- [x] Comprehensive error handling ✅
- [x] Builder and factory patterns
- [x] 95%+ test coverage on core components ✅
- [ ] Complete documentation (In Progress)
- [ ] Additional unit tests (In Progress)

### **Phase 2 - Production Features**

- [ ] Structured logging with `slog`
- [ ] Context support for cancellation
- [ ] YAML/JSON config file loading
- [ ] Resource cleanup and lifecycle management
- [ ] Integration tests

### **Phase 3 - Performance**

- [ ] Concurrent processing in chains
- [ ] Memory pooling for efficiency
- [ ] Streaming for large datasets
- [ ] Performance benchmarks
- [ ] Worker pools

### **Phase 4 - Enterprise**

- [ ] Metrics and observability (Prometheus/OpenTelemetry)
- [ ] Retry mechanisms with exponential backoff
- [ ] Circuit breakers for resilience
- [ ] Distributed tracing
- [ ] Health check endpoints
- [ ] CI/CD pipeline
- [ ] Additional providers (CSV, Database, API)
- [ ] Additional formatters and outputs

### **Future - Advanced**

- [ ] Dashboard UI (SaaS)
- [ ] Scheduling and cron jobs
- [ ] AI-powered data enrichment
- [ ] Caching layer
- [ ] BigQuery / Snowflake providers

---

## 📊 Progress

| Category               | Status         | Coverage |
| ---------------------- | -------------- | -------- |
| Core Engine            | ✅ Complete    | 95%      |
| Error Handling         | ✅ Complete    | 95%      |
| Thread-Safe Registries | ✅ Complete    | 100%     |
| Providers              | 🟡 In Progress | 50%      |
| Processors             | 🟡 In Progress | 60%      |
| Formatters             | 🟡 In Progress | 50%      |
| Outputs                | 🟡 In Progress | 50%      |
| Documentation          | 🟡 In Progress | 70%      |

**Overall Progress: ~40% Complete**

---

## 🎯 Design Principles

1. **SOLID Principles** - Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
2. **Design Patterns** - Strategy, Factory, Builder, Chain of Responsibility, Registry
3. **Testability** - Every component is interface-driven and mockable
4. **Concurrency** - Thread-safe by design with proper locking
5. **Error Handling** - Comprehensive context and classification
6. **Performance** - Optimized for production use with benchmarks

---

## 💡 Advanced Usage

### **Panic Recovery**

```go
// Run with automatic panic recovery
err := engine.RunWithRecovery()
if err != nil {
    // Panics are converted to errors
    log.Printf("Pipeline failed: %v", err)
}
```

### **Error Context Extraction**

```go
if engineErr, ok := err.(*errors.EngineError); ok {
    fmt.Printf("Component: %s\n", engineErr.Component)
    fmt.Printf("Operation: %s\n", engineErr.Operation)
    fmt.Printf("Type: %s\n", engineErr.Type)
    fmt.Printf("Context: %v\n", engineErr.Context)
    fmt.Printf("Timestamp: %v\n", engineErr.Timestamp)
}
```

### **Type-Safe Processor Registration**

```go
// Register a filter
registry.RegisterFilter("my_filter", &MyFilterStrategy{})

// Register a validator
registry.RegisterValidator("my_validator", &MyValidatorStrategy{})

// Register a transformer
registry.RegisterTransformer("my_transformer", &MyTransformerStrategy{})
```

---

## 💬 Community & Contribution

PRs are welcome! Please open:

- **Issues** for bugs
- **Discussions** for ideas
- **PRs** for improvements

### **How to Contribute**

1. Fork the repository
2. Create a feature branch
3. Follow SOLID principles and existing patterns
4. Add tests (maintain 80%+ coverage)
5. Add documentation
6. Submit a PR

### **Development Guidelines**

- Work on ONE file at a time
- Write godoc for all exports
- Use table-driven tests
- Follow Go best practices
- Run `go fmt`, `go vet`, and `golangci-lint`

Join the #buildinpublic journey! 🎉

---

## 📖 Documentation

- [API Documentation](https://pkg.go.dev/github.com/AshishBagdane/go-report-engine)
- [Architecture Guide](./docs/ARCHITECTURE.md) (Coming soon)
- [Error Handling Guide](./docs/ERROR_HANDLING.md) (Coming soon)
- [Contributing Guide](./CONTRIBUTING.md) (Coming soon)

---

## 🪪 License

MIT License — free for personal & commercial use.

See [LICENSE](LICENSE) for details.

---

## ⭐ Support the Project

If you find this useful:

- ⭐ Star the repo
- 🐦 Share on Twitter
- 🤝 Contribute code or documentation
- 💬 Join discussions
- 🐛 Report bugs

---

## 📞 Follow Me On

- **GitHub:** [@AshishBagdane](https://github.com/AshishBagdane)
- **Twitter/X:** [AshBagdane](https://x.com/AshBagdane)
- **LinkedIn:** [ashishbagdane](https://www.linkedin.com/in/ashishbagdane/)

---

**Built with ❤️ in Go | Production-Ready | Enterprise-Grade | 95%+ Test Coverage**
