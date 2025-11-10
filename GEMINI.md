# Gemini Guidelines

- Maintain English for all documentation and source code.
- Reply to users in the language used for their request.
- When `AGENTS.md` changes, update `CLAUDE.md` and `GEMINI.md` as well unless the user explicitly says otherwise.
- Write commit messages that are short, to the point, and in English.
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
- **Defensive Validation**: Do not add unnecessary nil checks or validations for function parameters (e.g., `context.Context`, interface implementations, struct pointers). Trust that callers provide valid parameters - focus on validating business logic and user input, not internal API contracts. If you believe validation is necessary, ask the user first.
- **Context Convention**: When adding methods or functions that require `context.Context`, the context parameter must always be the first parameter in the parameter list (for example `func ProcessRequest(ctx context.Context, userID string, data []byte)` instead of `func ProcessRequest(userID string, ctx context.Context, data []byte)`). This follows the standard Go convention for context propagation.
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
- **Home Directory Expansion**: Use `github.com/mitchellh/go-homedir` whenever you need to resolve `~` in file paths rather than manipulating strings manually.
- **YAML Parsing**: Use `sigs.k8s.io/yaml` for all YAML decoding and encoding tasks in the project.
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
- **MCP Servers**: The following MCP servers are available for enhanced development capabilities:

  **IDE Diagnostics:**
  - Use `mcp__ide__getDiagnostics` to retrieve diagnostic information (errors, warnings) from the IDE for specific files or the entire project

  **Upstash Context7 (Library Documentation):**
  - Use `mcp__upstash-context-7-mcp__resolve-library-id` to search for and retrieve Context7-compatible library IDs (e.g., "/mongodb/docs", "/vercel/next.js")
  - Use `mcp__upstash-context-7-mcp__get-library-docs` to fetch up-to-date documentation and code examples for any library
  - When to use Context7:
    - When you need current documentation for any library (Go, JavaScript, Python, etc.)
    - When looking for usage examples of specific functions or APIs
    - When working with library features and need reference documentation
    - Always call `resolve-library-id` first to obtain the correct library ID, unless the user provides an ID in the format "/org/project" or "/org/project/version"
