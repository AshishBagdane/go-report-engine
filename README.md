# 🚀 go-report-engine

A **modular, pluggable reporting engine for Go**.
Built using **Strategy**, **Factory**, **Template Method**, and **Chain of Responsibility** patterns.

**Fetch → Process → Format → Output — fully customizable.**
A clean, extensible architecture for building reporting pipelines in Go.

---

## ✨ Features

- 🔌 **Pluggable Providers**
  Fetch data from any source — DB, CSV, API, mock providers, etc.

- ♻️ **Processing Pipeline (Chain of Responsibility)**
  Add processors like Filter, Validate, Transform — fully customizable.

- 🧾 **Formatters**
  JSON, CSV, YAML (coming soon) format strategies.

- 📤 **Output Adapters**
  Console, Slack, Email, Files, Memory output.

- 🧱 **Enterprise-grade Architecture**
  SOLID principles + Go interfaces → clean, testable, extensible.

- 🧪 **Test-friendly Design**
  All components mockable via interfaces.

- 🌱 **Built in Public**
  Follow the real-time development journey.

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

### **1. Provider**

Responsible for fetching data.

Example:

- DBProvider
- CSVProvider
- APIProvider
- MockProvider

### **2. Processor Chain**

A Chain of Responsibility that transforms data step-by-step.

Examples:

- FilterProcessor
- ValidateProcessor
- TransformProcessor

### **3. Formatter**

Converts data into the desired output format.

Examples:

- JSONFormatter
- CSVFormatter
- YAMLFormatter

### **4. Output**

Delivers the final formatted report.

Examples:

- ConsoleOutput
- SlackOutput
- EmailOutput
- FileOutput

---

## 📁 Folder Structure

```
go-report-engine/
 ├── cmd/
 │    └── example/
 │         └── main.go
 ├── internal/
 │    ├── engine/
 │    ├── provider/
 │    ├── formatter/
 │    ├── output/
 │    ├── processor/
 │    └── factory/
 └── go.mod
```

---

## 🧰 Example Usage

Below is a minimal working pipeline:

```go
package main

import (
    "fmt"
    "github.com/AshishBagdane/go-report-engine/internal/engine"
    "github.com/AshishBagdane/go-report-engine/internal/provider"
    "github.com/AshishBagdane/go-report-engine/internal/processor"
    "github.com/AshishBagdane/go-report-engine/internal/formatter"
    "github.com/AshishBagdane/go-report-engine/internal/output"
)

func main() {
    mockProvider := provider.NewMockProvider()
    jsonFormatter := formatter.NewJSONFormatter()
    consoleOutput := output.NewConsoleOutput()

    // Simple processor chain: just passes data forward
    baseProcessor := &processor.BaseProcessor{}

    eng := engine.ReportEngine{
        Provider:  mockProvider,
        Processor: baseProcessor,
        Formatter: jsonFormatter,
        Output:    consoleOutput,
    }

    if err := eng.Run(); err != nil {
        fmt.Println("Error:", err)
    }
}
```

---

## 🔌 Providers (Data Sources)

Each provider implements:

```go
type ProviderStrategy interface {
    Fetch() ([]map[string]interface{}, error)
}
```

Available:

- [x] MockProvider
- [ ] CSVProvider
- [ ] DBProvider
- [ ] APIProvider

---

## ⚙️ Processors (Chain of Responsibility)

Processors implement:

```go
type ProcessorHandler interface {
    SetNext(next ProcessorHandler)
    Process(data []map[string]interface{}) ([]map[string]interface{}, error)
}
```

Built-in processors (coming soon):

- [ ] FilterProcessor
- [ ] ValidateProcessor
- [ ] TransformProcessor

---

## 🧾 Formatters

Formatter interface:

```go
type FormatStrategy interface {
    Format(data []map[string]interface{}) ([]byte, error)
}
```

Available:

- [x] JSONFormatter
- [ ] CSVFormatter
- [ ] YAMLFormatter

---

## 📤 Outputs

Output interface:

```go
type OutputStrategy interface {
    Send(data []byte) error
}
```

Available:

- [x] ConsoleOutput
- [ ] SlackOutput
- [ ] EmailOutput
- [ ] FileOutput

---

## 🏗️ Factories

Factories help generate the correct provider/formatter/output based on config.

Folders:

```
internal/factory/provider
internal/factory/formatter
internal/factory/output
```

---

## 🧪 Testing

Every component is interface-driven → fully unit-testable.

Example test pattern:

```go
func TestMockProvider(t *testing.T) {
    p := provider.NewMockProvider()
    data, err := p.Fetch()

    if err != nil || len(data) == 0 {
        t.Fatal("expected mock data")
    }
}
```

---

## 🗺️ Roadmap

### **MVP Phase**

- [x] Architecture skeleton
- [x] Core interfaces
- [x] BaseProcessor
- [ ] First provider (CSV)
- [ ] First formatter (JSON full)
- [ ] Console output
- [ ] Example usage

### **Phase 2**

- [ ] Filters + Validators
- [ ] File output
- [ ] YAML + CSV formatters
- [ ] DB provider

### **Phase 3**

- [ ] Slack + Email connectors
- [ ] Config system
- [ ] Logging

### **Phase 4 — Pro / Enterprise**

- [ ] Dashboard UI (SaaS)
- [ ] Scheduling
- [ ] AI enrichment processor
- [ ] Caching layer
- [ ] BigQuery / Snowflake providers

---

## 💬 Community & Contribution

PRs are welcome!
Please open:

- issues for bugs
- discussions for ideas
- PRs for improvements

Join the build-in-public journey 🎉

---

## 🪪 License

MIT License — free for personal & commercial use.

---

## ⭐ Support the Project

If you find this useful:

- Star ⭐ the repo
- Share it
- Contribute
