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
- **Structured Logging**: When using Go's `slog`, always wrap attribute values with the appropriate helpers (for example `slog.String`, `slog.Int`). Prefer camelCase over snake_case for field names in contextual loggers (e.g., `slog.String("userId", id)` instead of `slog.String("user_id", id)`). Log messages should be complete sentences with proper capitalization and punctuation (e.g., `logger.Info("Failed to connect to server.")` instead of `logger.Info("failed to connect")`).
- **Error Convention**: When it makes sense, define errors as variables with the `Err` prefix (for example `var ErrNotFound = errors.New("not found")`). This is the standard Go convention for sentinel errors.
- **Error Wrapping**: When returning errors, always add clear context with `fmt.Errorf` and the `%w` verb (for example `fmt.Errorf("create machine from config: %w", err)` or `fmt.Errorf("read file %s: %w", filePath, err)`) instead of passing the original error unchanged.
- **Testing Assertions**: When writing tests in Go, use the `testify/assert` package directly (for example `assert.NoError(t, err)`, `assert.Equal(t, expected, actual)`) instead of creating custom assertion helpers, so failures produce descriptive messages out of the box.
- **Peripheral SDK Alias**: Always import `github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral` using the alias `peripheralSDK` to avoid naming conflicts with internal packages.
- **Naming Conventions**:
  - **JSON Tags**: Always use camelCase for field names in JSON tags (for example `json:"userName"` instead of `json:"user_name"`).
  - **Enum and Config Values**: Always use kebab-case for enum values and configuration option values (for example `"auto-detect"`, `"usb-device"`, `"high-performance"`).
  - **Receivers**: Receiver names must be clear and descriptive, not single or two-letter abbreviations. Use meaningful names that indicate the receiver's role (for example `(manager *Manager)` is acceptable, but prefer `(displayManager *Manager)` when there are multiple manager types. Avoid generic `(m *Manager)` or `(d *Display)`).
  - **Variables**: Avoid single-letter variable names and meaningless abbreviations. Use full, descriptive names (for example `config` instead of `cfg`, `connection` instead of `conn`). Exceptions are allowed for universally accepted short names in limited scopes (like `i` in simple for-loops iterating over indices, `err` for errors, `ok` for boolean checks, `ctx` for context.Context, and `wg` for sync.WaitGroup).
- **MCP Servers**: When you perform code operations, prefer using MCP server tools over traditional command-line utilities, as they provide better integration with the development environment. Follow these preferences:

  **File Operations:**
  - Reading files: use `get_file_text_by_path` instead of `cat`, `head`, or `tail`
  - Editing files: use `replace_text_in_file` instead of `sed` or `awk`
  - Creating files: use `create_new_file` with content
  - Formatting: use `reformat_file` instead of manual formatters

  **Search Operations:**
  - Finding files by name: use `find_files_by_name_keyword` (fastest) or `find_files_by_glob` instead of `find` or `ls`
  - Searching in files: use `search_in_files_by_text` or `search_in_files_by_regex` instead of `grep` or `rg`
  - Directory listing: use `list_directory_tree` instead of `ls` or `tree`

  **Refactoring:**
  - Renaming symbols (classes, functions, variables): **always** use `rename_refactoring` instead of text replacement. This tool understands code structure and updates all references project-wide
  - Understanding symbols: use `get_symbol_info` to get declaration details and documentation

  **Code Analysis:**
  - File-level problems: use `get_file_problems` to check errors and warnings in specific files
  - Project-level problems: use `get_project_problems` for global code analysis
  - Project structure: use `get_project_modules` and `get_project_dependencies` to understand architecture

  **Running and Testing:**
  - **Tests**: Always run tests using `execute_run_configuration` with appropriate test run configurations (e.g., "All Tests", "Package Tests"). Never use `go test` directly via terminal
  - Prefer `execute_run_configuration` over manual terminal commands when run configurations exist
  - Use `get_run_configurations` to discover available build/test/run configurations
  - Use `execute_terminal_command` only when MCP-specific tools are not applicable
