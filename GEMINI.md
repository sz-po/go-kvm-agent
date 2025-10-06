# Gemini Guidelines

- Maintain English for all documentation and source code.
- Reply to users in the language used for their request.
- When `AGENTS.md` changes, update `CLAUDE.md` and `GEMINI.md` as well unless the user explicitly says otherwise.
- Write commit messages that are short, to the point, and in English.
- **AI-DEV Comments**: Always respect special directives in code comments prefixed with `AI-DEV:`, such as:
  - `AI-DEV: never modify this interface`
  - `AI-DEV: always ask user before modifying this interface`
  - Follow these directives strictly when working with the codebase.
- **Enum Convention**: When defining integer-based enums in Go (using `iota`), always include an `Unknown` or equivalent zero-value constant as the first element (position 0). This ensures a safe default state for uninitialized or invalid values.
