# go-kvm-agent

Go-based KVM agent designed for pixel-perfect remote control of Linux workstations.

## Overview

- Displays video from capture cards and NVIDIA raw output with ultra-low latency.
- Provides an API server to inject inputs and retrieve live output.
- Operates in an agent mode orchestrated by an LLM for autonomous or assisted control loops.
- Targets Linux environments and emphasizes deterministic, lossless rendering.

## Project Layout

- `build/` – tooling, CI/CD scripts, packaging artifacts.
- `cmd/` – application entrypoints (e.g., `cmd/go-kvm-agent`).
- `internal/` – private application modules.

## Getting Started

1. Ensure compatible Linux host with supported video capture hardware and NVIDIA drivers.
2. Install Go 1.22+ and clone the repository.
3. Adjust `go.mod` module path, then run `go mod tidy` to sync dependencies.
4. Build the agent with `go build ./cmd/go-kvm-agent`.


## Machine Configuration

Machine definitions live in standalone JSON or YAML files. Provide a directory via `--machines-dir` and the agent will load every `*.json`, `*.yaml`, or `*.yml` file found there. Each file may define a single machine or a list under `machines`.

Example files are available in `examples/machines` and cover two separate workstations:

```
go run ./cmd/go-kvm-agent --machines-dir=./examples/machines
```

Within each file, `type` and `role` use the integer constants from `pkg/peripherals/peripheral_common.go` (`display`=1, `keyboard`=2, `mouse`=3; `source`=1, `sink`=2). Drivers and params are passed as-is to the corresponding peripheral implementations.
## Architecture

The agent is organized around modular device abstractions and real-time routing:

- Device interfaces live in `pkg/peripherals` (keyboard, mouse, display); each interface defines a `Source` that emits events and a `Sink` that applies them, while concrete devices implement those contracts in their own packages.
- The event router binds active sources to sinks according to the current session profile, letting the agent redirect control dynamically.
- Video capture pipes feed display sinks to maintain pixel-perfect output backed by NVIDIA raw buffers.
- Input sinks translate API or LLM-issued actions into device-level events with deterministic timing.
- The control plane (API server + LLM agent loop) orchestrates configuration changes, manages sessions, and supervises error handling.

The `Source`/`Sink` naming mirrors stream-processing terminology and keeps direction explicit: sources publish events, sinks consume them. Alternative labels (producer/consumer, emitter/receiver) were considered, but source/sink best fits the mixed media + input domain and aligns with Go streaming conventions.
For detailed development guidance and display interface semantics, see
`DEVELOPMENT.md`.

## Next Steps

- Implement media capture pipelines feeding the pixel-perfect renderer.
- Extend the API surface for input control and output streaming.
- Integrate LLM-driven orchestration logic for agentic workflows.
