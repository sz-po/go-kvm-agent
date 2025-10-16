# go-kvm-agent

Go-based KVM agent designed for pixel-perfect remote control of Linux workstations.

## Overview

- Captures and routes video streams from multiple sources (capture cards, test patterns) with ultra-low latency.
- Implements a dynamic display routing system that connects video sources to rendering targets (MPV windows, future: HDMI outputs).
- Provides an HTTP API for runtime control of display routing and peripheral management.
- Designed for remote workstation control with future support for LLM-driven autonomous operation.
- Targets Linux environments with emphasis on deterministic, pixel-perfect rendering.
- Currently implements ffmpeg-based display sources and MPV-based display sinks.

## Project Layout

- `build/` – tooling, CI/CD scripts, packaging artifacts.
- `cmd/` – application entrypoints (e.g., `cmd/go-kvm-agent`).
- `internal/` – private application modules.
- `pkg/` – public SDK interfaces for peripherals and routing.
- `examples/` – example configurations and API requests.

## Requirements

### System Requirements
- Linux operating system (tested on modern distributions)
- Go 1.24.0 or later
- FFmpeg (for display sources)
- MPV (for display sinks)

### Optional Hardware
- Video capture hardware (HDMI capture cards) for production use
- NVIDIA GPU with raw output support (future feature)

### Installation

Install system dependencies:

```bash
# Debian/Ubuntu
sudo apt-get install ffmpeg mpv

# Fedora/RHEL
sudo dnf install ffmpeg mpv

# Arch Linux
sudo pacman -S ffmpeg mpv
```

## Getting Started

1. Clone the repository and sync dependencies:
   ```bash
   git clone <repository-url>
   cd go-kvm-agent
   go mod tidy
   ```

2. Build the agent:
   ```bash
   go build ./cmd/go-kvm-agent
   ```

3. Run with example configurations:
   ```bash
   ./go-kvm-agent --machine.config-path=./examples/config/machines
   ```

   This starts the agent with two example machines:
   - `mpv-ffmpeg`: FFmpeg display source generating a test pattern
   - `mpv-mpv-window`: MPV window display sink for rendering

4. The agent will start an HTTP API server on `http://localhost:8080` (configurable).

5. Connect a display source to a display sink using the API (see HTTP API section below).


## Machine Configuration

Machine definitions live in standalone JSON or YAML files. Provide a directory via `--machine.config-path` and the agent will load every `*.json`, `*.yaml`, or `*.yml` file found there. Each file defines a machine with its peripherals.

Example files are available in `examples/config/machines`:

```bash
./go-kvm-agent --machine.config-path=./examples/config/machines
```

### Configuration Format

Each machine configuration file has the following structure:

```yaml
name: machine-name
peripherals:
  - driver: peripheral-driver-name
    config:
      # Driver-specific configuration
```

## Available Peripheral Drivers

### ffmpeg/display-source

FFmpeg-based display source that captures video from various inputs. Currently supports test pattern generation.

**Configuration Example:**
```yaml
name: test-ffmpeg-source
peripherals:
  - driver: ffmpeg/display-source
    config:
      input:
        testPattern:
          displayMode:
            width: 1920
            height: 1080
            refreshRate: 30
```

**Configuration Options:**
- `input.testPattern.displayMode.width` - Frame width in pixels
- `input.testPattern.displayMode.height` - Frame height in pixels
- `input.testPattern.displayMode.refreshRate` - Frames per second

### mpv/window

MPV-based display sink that renders video in a window on the local display.

**Configuration Example:**
```yaml
name: test-mpv-window
peripherals:
  - driver: mpv/window
    config:
      title: "KVM Agent Display"
      supportedDisplayModes:
        - width: 1920
          height: 1080
          refreshRate: 30
        - width: 1280
          height: 720
          refreshRate: 30
```

**Configuration Options:**
- `title` - Window title (optional)
- `supportedDisplayModes` - List of display modes this sink can handle. The router will configure the sink to match the source's display mode from this list.

## HTTP API

The agent exposes an HTTP API for runtime control. By default, it listens on `http://localhost:8080`.

### Connect Display Source to Display Sink

**Endpoint:** `POST /router/display/connect`

**Request Body:**
```json
{
  "displaySourceId": "ffmpeg-display-source-<uuid>",
  "displaySinkId": "mpv-window-<uuid>"
}
```

**Response:** `204 No Content` on success

**Example:**
```bash
curl -X POST http://localhost:8080/router/display/connect \
  -H "Content-Type: application/json" \
  -d '{
    "displaySourceId": "ffmpeg-display-source-48358aa8-1c50-4ebc-88d9-8fe60e5e86f8",
    "displaySinkId": "mpv-window-15110699-dd22-4c36-ac6c-caa5aba28703"
  }'
```

**Finding Peripheral IDs:**

When the agent starts, it logs the peripheral IDs. Look for lines like:
```
Peripheral created. machineName=mpv-ffmpeg peripheralId=ffmpeg-display-source-48358aa8-1c50-4ebc-88d9-8fe60e5e86f8
Peripheral created. machineName=mpv-mpv-window peripheralId=mpv-window-15110699-dd22-4c36-ac6c-caa5aba28703
```

An example HTTP request file is available at `examples/api/control/display-router-connect.http`.

## Architecture

The agent is organized around modular peripheral abstractions and dynamic routing:

### Core Abstractions

**Peripherals:** All devices (displays, keyboards, mice) implement the `Peripheral` interface defined in `pkg/peripheral`. Each peripheral has:
- A unique ID (auto-generated UUID with driver prefix)
- Capabilities that declare its kind (display/keyboard/mouse) and role (source/sink)
- Lifecycle management (initialization and termination)

**Source/Sink Pattern:** Peripherals are split by data flow direction:
- **Sources** emit events (e.g., `DisplaySource` emits frame data, `KeyboardSource` emits key events)
- **Sinks** consume events (e.g., `DisplaySink` renders frames, `KeyboardSink` injects keystrokes)

This design enables flexible routing where any source can connect to any compatible sink.

**Machines:** A `Machine` represents a physical or virtual workstation and groups related peripherals. Each machine is loaded from a configuration file and manages its peripheral lifecycle.

**Display Router:** The `DisplayRouter` (`pkg/routing`) dynamically connects display sources to sinks at runtime:
- Negotiates display modes between source and sink
- Pipes frame data events from source channels to sink handlers
- Manages connection lifecycle and handles reconnection

### Data Flow

1. **Peripheral Creation:** Machine configurations are loaded and peripherals instantiated by driver-specific factories
2. **Registration:** All peripherals register with a central repository
3. **Routing:** The DisplayRouter builds a registry of available sources and sinks
4. **Connection:** API calls connect specific source/sink pairs
5. **Streaming:** Frame events flow from source channels through the router to sink handlers

### Current Implementations

- **ffmpeg/display-source:** Uses FFmpeg to generate test patterns or capture video
- **mpv/window:** Uses MPV to render frames in a local window
- **LocalDisplayRouter:** In-process routing implementation with goroutine-based event forwarding

### Event Streaming

Display sources expose two channels:
- `DisplayDataChannel`: Emits frame events (start, chunk, end)
- `DisplayControlChannel`: Emits control events (metrics, errors, mode changes)

Display sinks implement handlers for these events, maintaining frame buffers and rendering complete frames.

For detailed development guidance and interface contracts, see `DEVELOPMENT.md`.

## Roadmap

### Near-term
- Implement real video capture sources (HDMI capture cards, V4L2 devices)
- Add keyboard and mouse peripheral implementations
- Implement disconnect operations for display router
- Add metrics and monitoring endpoints to HTTP API

### Mid-term
- Network-based routing (remote sources and sinks)
- NVIDIA raw output support for GPU-accelerated workstations
- WebRTC streaming for browser-based remote access
- Authentication and authorization for HTTP API

### Long-term
- LLM-driven orchestration and autonomous control loops
- Session management and multi-user support
- Advanced input injection with timing control
- Recording and playback of control sessions
