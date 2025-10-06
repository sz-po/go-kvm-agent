# AGENTS Guidelines

- All project documentation and source code must be written in English.
- User-facing output should match the language of the incoming request.
- If you modify `AGENTS.md`, you must also update `CLAUDE.md` and `GEMINI.md` unless the user explicitly instructs otherwise.
- Commit messages must be short, concise, and written in English.
- **AI-DEV Comments**: Always respect special directives in code comments prefixed with `AI-DEV:`, such as:
  - `AI-DEV: never modify this interface`
  - `AI-DEV: always ask user before modifying this interface`
  - Follow these directives strictly when working with the codebase.
- **Enum Convention**: When defining integer-based enums in Go (using `iota`), always include an `Unknown` or equivalent zero-value constant as the first element (position 0). This ensures a safe default state for uninitialized or invalid values.
