# Configuration File Examples

This directory contains example configuration files demonstrating different config formats and use cases.

## Files

### `config.yaml`

Complete configuration example with all available options:

- Provider configuration with parameters
- Multiple processors in the pipeline
- Formatter configuration
- Output configuration

**Use case:** Reference for all configuration options

### `config.minimal.yaml`

Minimal valid configuration:

- Basic provider, formatter, and output
- No processors
- Simplest working configuration

**Use case:** Quick start, testing, simple deployments

### `config.json`

Same as `config.yaml` but in JSON format:

- Demonstrates JSON configuration support
- Identical functionality to YAML version

**Use case:** When JSON is preferred over YAML

## Usage

### From Code

```go
// Load YAML config
config, err := config.LoadFromFile("examples/configs/config.yaml")

// Load JSON config
config, err := config.LoadFromFile("examples/configs/config.json")

// Load minimal config
config, err := config.LoadFromFile("examples/configs/config.minimal.yaml")
```

### With Environment Overrides

```bash
# Set environment variables
export ENGINE_PROVIDER_TYPE=postgres
export ENGINE_PROVIDER_PARAM_HOST=localhost
export ENGINE_PROVIDER_PARAM_PORT=5432

# Load config with overrides
config, err := config.LoadFromFileWithEnv("examples/configs/config.yaml")
```

## Format Support

Both YAML and JSON formats are supported. The loader automatically detects the format based on file extension:

- `.yaml`, `.yml` → YAML parser
- `.json` → JSON parser

## Creating Your Own Config

1. Copy `config.minimal.yaml` as a starting point
2. Add your provider configuration
3. Add processors as needed
4. Configure formatter options
5. Set output destination

See the main README for detailed configuration documentation.
