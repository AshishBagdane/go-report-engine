# Integration Reference

This guide details the helper functions available in the `config` package to simplify initializing the Report Engine.

## Loading and Building

### `LoadAndBuild(path string)`
Loads configuration from a YAML or JSON file and builds the engine.
```go
engine, err := config.LoadAndBuild("config.yaml")
```

### `LoadAndBuildWithEnv(path string)`
Loads configuration from a file and applies environment variable overrides.
```go
// Overrides: ENGINE_PROVIDER_TYPE=postgres, ENGINE_FORMATTER_PARAM_INDENT=4
engine, err := config.LoadAndBuildWithEnv("config.yaml")
```

### `BuildFromBytes(data []byte, format string)`
Builds the engine from raw configuration data.
```go
data := []byte(`provider: {type: mock}`)
engine, err := config.BuildFromBytes(data, "yaml")
```

## "Must" Variants
These functions panic on error, making them suitable for `init()` or application startup where failure is fatal.

### `MustLoadAndBuild(path string)`
Same as `LoadAndBuild` but panics on error.
```go
engine := config.MustLoadAndBuild("config.yaml")
```

## Presets
Quickly get a configured engine for standard environments.

### `BuildFromDevelopment()` / `MustBuildFromDevelopment()`
- **Provider**: Mock
- **Formatter**: JSON (indented)
- **Output**: Console
- **Logging**: Debug level

### `BuildFromProduction()` / `MustBuildFromProduction()`
- **Provider**: Mock (Replace with real)
- **Formatter**: JSON (Compact)
- **Output**: File
- **Logging**: Info level (JSON format)

### `BuildFromTesting()` / `MustBuildFromTesting()`
- **Provider**: Mock
- **Formatter**: JSON
- **Output**: Console
- **Logging**: Error level

## Fallback Pattern

### `LoadOrDefault(path string)`
Attempts to load config from `path`. If the file doesn't exist, returns a default configuration structure instead of an error.
```go
cfg, err := config.LoadOrDefault("config.yaml")
// If config.yaml is missing, cfg is the default config
engine, err := config.ValidateAndBuild(*cfg)
```
