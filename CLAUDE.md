# Claude Guidelines

- Ensure all documentation and source code remain in English.
- Generate user-facing output in the same language as the original question.
- Any change to `AGENTS.md` requires synchronized updates to `CLAUDE.md` and `GEMINI.md` unless explicitly waived by the user.
- Keep commit messages brief, clear, and in English.
- **AI-DEV Comments**: Always respect special directives in code comments prefixed with `AI-DEV:`, such as:
  - `AI-DEV: never modify this interface`
  - `AI-DEV: always ask user before modifying this interface`
  - Follow these directives strictly when working with the codebase.
- **Enum Convention**: When defining integer-based enums in Go (using `iota`), always include an `Unknown` or equivalent zero-value constant as the first element (position 0). This ensures a safe default state for uninitialized or invalid values.
- **Structured Logging**: When using Go's `slog`, always wrap attribute values with the appropriate helpers (for example `slog.String`, `slog.Int`).
