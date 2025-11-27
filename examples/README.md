# Examples

This directory contains practical examples demonstrating how to use the go-report-engine with different configuration approaches.

## Directory Structure

```
examples/
├── configs/                    # Example configuration files
│   ├── config.yaml            # Complete YAML config
│   ├── config.minimal.yaml    # Minimal working config
│   └── config.json            # Complete JSON config
├── scripts/                    # Helper scripts
│   └── env_override_demo.sh   # Environment variable demo
├── config_loading/             # Config file loading examples
├── defaults_usage/             # Default config examples
└── integration_patterns/       # Integration patterns
```

## Quick Start

### 1. Config Loading (`config_loading/`)

Demonstrates loading configuration from files with environment variable support.

```bash
cd examples/config_loading
go run main.go
```

**What you'll learn:**

- Loading YAML configurations
- Loading JSON configurations
- Environment variable overrides

### 2. Default Configs (`defaults_usage/`)

Shows how to use preset configurations and builder patterns.

```bash
cd examples/defaults_usage
go run main.go
```

**What you'll learn:**

- Using default configurations
- Development vs production presets
- Builder pattern for config construction
- Combining file configs with defaults

### 3. Integration Patterns (`integration_patterns/`)

Demonstrates various integration patterns for real-world usage.

```bash
cd examples/integration_patterns
go run main.go
```

**What you'll learn:**

- One-step load and build
- Must variants for initialization
- Fallback patterns
- Building from raw bytes

## Configuration Files

All example configurations are in the `configs/` directory:

- **`config.yaml`** - Complete configuration showing all options
- **`config.minimal.yaml`** - Simplest working configuration
- **`config.json`** - JSON format configuration

See [`configs/README.md`](configs/README.md) for detailed configuration documentation.

## Environment Variables

Use the demo script to see environment variable overrides in action:

```bash
cd examples/scripts
./env_override_demo.sh
```

Then run any example to see the overrides take effect.

## Usage Patterns

### Production Pattern

```go
// Load from standard location with env overrides
engine, err := config.LoadAndBuildWithEnv("/etc/myapp/config.yaml")
if err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}
engine.Run()
```

### Development Pattern

```go
// Use default if config not found
cfg, _ := config.LoadOrDefault("config.yaml")
engine, _ := config.ValidateAndBuild(*cfg)
engine.Run()
```

### Testing Pattern

```go
// Use testing preset
engine := config.MustBuildFromTesting()
engine.Run()
```

### Initialization Pattern

```go
func init() {
    // Panic on error - appropriate for init
    engine := config.MustLoadAndBuild("config.yaml")
}
```

## Running Examples

All examples can be run from their respective directories:

```bash
# From repository root
go run examples/config_loading/main.go
go run examples/defaults_usage/main.go
go run examples/integration_patterns/main.go
```

Or run all examples:

```bash
for dir in config_loading defaults_usage integration_patterns; do
    echo "=== Running $dir ==="
    go run examples/$dir/main.go
done
```

## Next Steps

1. Start with `config_loading/` to understand basic file loading
2. Explore `defaults_usage/` to learn about presets and builders
3. Study `integration_patterns/` for production-ready patterns
4. Review the config files in `configs/` to understand structure
5. Try the environment variable demo in `scripts/`

## See Also

- [Integration Functions Reference](../docs/INTEGRATION_REFERENCE.md)
- [Main README](../README.md)
- Project documentation in `docs/`
