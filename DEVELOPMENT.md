# Development Guide

This document explains the core abstractions used by the agent and outlines the
expected development workflow.

## Display Interfaces

Display devices are split into `DisplaySource` and `DisplaySink` implementations
defined under `pkg/peripherals`.

- `DataChannel(ctx)` and `ControlChannel(ctx)` may be obtained before a source or
  sink is started. Both channels stay open regardless of `Start`/`Stop` calls;
  canceling the supplied context is the canonical way to signal stream teardown.
- `Start(ctx, info)` / `Stop(ctx)` on sources configure or release the capture
  hardware without side effects on previously acquired channels. Sources emit
  data events (frame start/chunk/end) and control events (metrics, mode changes,
  lifecycle notifications).
- `Start(ctx)` / `Stop(ctx)` on sinks configure or release the rendering target.
  For HDMI grabbers this is where negotiated display modes are applied. Sinks
  continue to read from their control channel until the context is canceled.
- Frame and control events have dedicated constructor helpers in
  `pkg/peripherals/display_events.go` to guarantee timestamps are populated.

These contracts ensure that channel ownership remains with the caller while
hardware lifecycle management is explicit. Implementations should be careful to
avoid closing caller-provided channels and to respect context cancellation.

## Keyboard Interfaces

Keyboard peripherals mirror the source/sink split and live alongside display
contracts in `pkg/peripherals`.

- `EventChannel(ctx)` / `ControlChannel(ctx)` on sources and sinks follow the
  same rules as display components: channels may be fetched ahead of
  `Start`/`Stop`, and context cancellation is the supported teardown path.
- `KeyboardKeyEvent` instances (defined in `keyboard_events.go`) include the
  physical scan code, HID usage, logical key metadata, modifier bitmask, and an
  optional text payload so downstream code can choose the abstraction layer it
  needs.
- Layout negotiation relies on `KeyboardInfo`, `KeyboardLayout`, and
  `KeyboardLayoutChangedEvent`. Sources expose their current layout via
  `GetCurrentLayout()` and publish changes on the control channel; sinks apply
  updates through `SetLayout`.
- Control events also carry LED state transitions (`KeyboardLEDState`, emitted
  by sinks) as well as metrics and lifecycle notifications. Data channels stay
  reserved for user-generated key transitions.
- Constructor helpers ensure timestamps are always set on keyboard data and
  control events, making telemetry ordering deterministic.

## Development Workflow

- Add devices by extending interfaces in `pkg/peripherals` and creating matching
  `Source`/`Sink` pairs under the relevant runtime package before registering
  them with the router.
- Extend the router with configuration intents that map logical endpoints (e.g.,
  "LLM control" or "API passthrough") to concrete sinks.
- Keep transport code in `internal/` focused on orchestration; expose only
  narrowly scoped interfaces from `pkg/peripherals` to avoid leaky abstractions.
- Validate latency-critical paths with integration tests or profilers that
  replay captured sessions before merging substantial changes.
- Document new API surface areas alongside example client flows so LLM-driven
  automation can exercise them reliably.
