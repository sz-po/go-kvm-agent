# Claude Guidelines

- Ensure all documentation and source code remain in English.
- Generate user-facing output in the same language as the original question.
- Any change to `AGENTS.md` requires synchronized updates to `CLAUDE.md` and `GEMINI.md` unless explicitly waived by the user.
- Keep commit messages brief, clear, and in English.
- **AI-DEV Comments**: Always respect special directives in code comments prefixed with `AI-DEV:`, such as:
  - `AI-DEV: never modify this interface`
  - `AI-DEV: always ask user before modifying this interface`
  - Follow these directives strictly when working with the codebase.
- **When Uncertain - Ask First**: Before making changes or implementing features, **always ask the user for clarification** when:
  - You don't fully understand what the user expects or wants to achieve
  - The user's request is ambiguous or could be interpreted in multiple ways
  - You're unsure about the purpose or intended behavior of existing code you're working with
  - You need to make architectural decisions that could impact system design
  - You're uncertain whether your proposed solution aligns with user expectations
  - Multiple valid approaches exist and the best choice depends on user preferences or requirements
  - **Never guess or assume** - it's always better to ask than to implement the wrong solution. Present your understanding and ask for confirmation before proceeding with significant changes.
- **Enum Convention**: When defining integer-based enums in Go (using `iota`), always include an `Unknown` or equivalent zero-value constant as the first element (position 0). This ensures a safe default state for uninitialized or invalid values.
- **Structured Logging**: When using Go's `slog`, always wrap attribute values with the appropriate helpers (for example `slog.String`, `slog.Int`). Prefer camelCase over snake_case for field names in contextual loggers (e.g., `slog.String("userId", id)` instead of `slog.String("user_id", id)`). Log messages should be complete sentences with proper capitalization and punctuation (e.g., `logger.Info("Failed to connect to server.")` instead of `logger.Info("failed to connect")`).
- **Error Convention**: When it makes sense, define errors as variables with the `Err` prefix (for example `var ErrNotFound = errors.New("not found")`). This is the standard Go convention for sentinel errors.
- **Error Wrapping**: When returning errors, always add clear context with `fmt.Errorf` and the `%w` verb (for example `fmt.Errorf("create machine from config: %w", err)` or `fmt.Errorf("read file %s: %w", filePath, err)`) instead of passing the original error unchanged.
- **Testing Assertions**: When writing tests in Go, use the `testify/assert` package directly (for example `assert.NoError(t, err)`, `assert.Equal(t, expected, actual)`) instead of creating custom assertion helpers, so failures produce descriptive messages out of the box.
- **Coverage Verification**: After adding or updating tests, run them with coverage enabled and inspect the report (for example `go test -coverprofile` followed by `go tool cover -func`) to confirm the affected files are exercised.
- **Mock Testing**: When using testify/mock with expecters (mockery generated mocks), follow these conventions:
  - **Mock Structure**: Mock structs embed `mock.Mock` without additional fields for storing values (for example `type MyMock struct { mock.Mock }` instead of `type MyMock struct { mock.Mock; value string }`).
  - **Use Generated Constructors**: Always create mocks using the generated constructor that accepts `*testing.T` (for example `m := NewMockRequester(t)`). The constructor automatically registers expectations and cleanup.
  - **Configure with EXPECT()/Return**: Set mock behavior using the expecter pattern `mockObj.EXPECT().MethodName(args...).Return(values...)` instead of `On()`. For example:
    ```go
    m := NewMockRequester(t)
    m.EXPECT().Get("foo").Return("bar", nil).Once()
    retString, err := m.Get("foo")
    ```
  - See [testify/mock documentation](https://pkg.go.dev/github.com/stretchr/testify/mock) for details on expecter pattern, `Return()`, `Once()`, and argument matchers like `mock.Anything` or `mock.MatchedBy()`.
- **Peripheral SDK Alias**: Always import `github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral` using the alias `peripheralSDK` to avoid naming conflicts with internal packages.
- **Naming Conventions**:
  - **JSON Tags**: Always use camelCase for field names in JSON tags (for example `json:"userName"` instead of `json:"user_name"`).
  - **Enum and Config Values**: Always use kebab-case for enum values and configuration option values (for example `"auto-detect"`, `"usb-device"`, `"high-performance"`).
  - **Receivers**: Receiver names must be clear and descriptive, not single or two-letter abbreviations. Use meaningful names that indicate the receiver's role (for example `(manager *Manager)` is acceptable, but prefer `(displayManager *Manager)` when there are multiple manager types. Avoid generic `(m *Manager)` or `(d *Display)`).
  - **Variables**: Avoid single-letter variable names and meaningless abbreviations. Use full, descriptive names (for example `config` instead of `cfg`, `connection` instead of `conn`). Exceptions are allowed for universally accepted short names in limited scopes (like `i` in simple for-loops iterating over indices, `err` for errors, `ok` for boolean checks, `ctx` for context.Context, and `wg` for sync.WaitGroup).
- **Code Readability**: Code readability is the top priority when writing code. Follow these principles:
  - Write self-documenting code with clear, descriptive variable and function names that express intent
  - Avoid abbreviations and single-letter variable names (exceptions listed in Naming Conventions above)
  - Structure code in logical blocks that are easy to follow
  - Add comments only where truly necessary - when the code cannot be made self-explanatory or when documenting complex business logic, algorithms, or non-obvious decisions
  - Prefer refactoring unclear code over adding explanatory comments
  - When comments are needed, write them as complete sentences explaining "why" rather than "what"
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
