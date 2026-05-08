# prom-label-enforcer

Prometheus middleware that validates and enforces required label sets on incoming metrics before storage.

---

## Installation

```bash
go get github.com/yourusername/prom-label-enforcer
```

---

## Usage

```go
package main

import (
    "github.com/yourusername/prom-label-enforcer/enforcer"
    "github.com/prometheus/client_golang/prometheus"
)

func main() {
    required := []string{"env", "service", "region"}

    enc := enforcer.New(enforcer.Config{
        RequiredLabels: required,
        OnViolation:    enforcer.DropMetric, // or enforcer.RejectWithError
    })

    registry := prometheus.NewRegistry()
    wrapped := enc.Wrap(registry)

    // Use wrapped registry as you would a standard prometheus.Registry
    counter := prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests.",
    }, []string{"env", "service", "region", "method"})

    wrapped.MustRegister(counter)

    // This will pass validation
    counter.WithLabelValues("production", "api", "us-east-1", "GET").Inc()

    // This will trigger the violation handler (missing required labels)
    counter.WithLabelValues("", "", "", "POST").Inc()
}
```

### Violation Strategies

| Strategy          | Behavior                                      |
|-------------------|-----------------------------------------------|
| `DropMetric`      | Silently drops metrics missing required labels |
| `RejectWithError` | Returns an error and rejects the metric        |

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

---

## License

This project is licensed under the [MIT License](LICENSE).