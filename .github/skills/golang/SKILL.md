---
name: golang
description: "Write Go code for the Fiagram project. Use when creating new packages, modules, interfaces, structs, DAO accessors, logic layers, HTTP handlers, tests, or any Go source files. Use when asked about Go conventions, architecture, or patterns in this project."
---

# Go Language — Fiagram Project Conventions

## When to Use
- Creating new Go source files or packages
- Adding DAO accessors, logic layers, HTTP handlers, or cache accessors
- Writing interfaces, structs, constructors, or module registrations
- Writing or modifying tests
- Reviewing Go code for convention compliance

## Project Setup
- Module: `github.com/Fiagram/standalone`
- Go version: 1.25+
- DI framework: Uber `go.uber.org/fx`
- HTTP framework: `github.com/gin-gonic/gin`
- Logger: `go.uber.org/zap`
- Testing: `github.com/stretchr/testify` (require, not assert)
- OpenAPI codegen: `oapi-codegen` (generates `internal/generated/openapi/oapi.gen.go`)
- Build: `make build`, `make test`, `make lint`, `make generate`

## Architecture Layers

```
cmd/standalone/main.go          → CLI entry (cobra)
internal/app/standalone.go      → fx.Module composition
internal/configs/               → Configuration structs + fx providers
internal/logger/                → Zap logger init
internal/dao/                   → Data Access Objects (database, cache)
internal/logic/                 → Business logic
internal/handler/               → HTTP handlers + webhook servers
internal/utils/                 → Generic utility functions
internal/generated/             → Auto-generated code (do NOT edit)
```

Dependency flow: `configs → logger → dao → logic → handler`

## Uber fx Module Pattern

Every layer MUST export a `Module` variable using `fx.Module()`:

```go
package mypackage

import (
    "go.uber.org/fx"
)

var Module = fx.Module(
    "mypackage",
    fx.Provide(
        NewMyService,
    ),
)
```

- Module name matches the package purpose (e.g., `"dao"`, `"logic"`, `"handler"`, `"config"`)
- Use `fx.Provide()` for constructor functions
- Use `fx.Invoke()` only in the handler layer for lifecycle hooks (start/stop servers)
- Register new modules in `internal/app/standalone.go`

### Lifecycle Hooks (Handler Layer Only)

```go
fx.Invoke(
    func(lc fx.Lifecycle, server MyServer) {
        var cancel context.CancelFunc
        lc.Append(fx.Hook{
            OnStart: func(_ context.Context) error {
                var ctx context.Context
                ctx, cancel = context.WithCancel(context.Background())
                go server.Start(ctx)
                return nil
            },
            OnStop: func(_ context.Context) error {
                cancel()
                return nil
            },
        })
    },
)
```

## Naming Conventions

### Packages
- Lowercase, single-word when possible: `configs`, `logger`, `handler`, `dao`
- Sub-packages: `dao/cache`, `dao/database`, `logic/account`, `logic/http`, `handler/http`
- Package name in `package` declaration uses underscore aliases when nested: `package dao_database`, `package logic_account`, `package http_handler`

### Imports
- Alias nested internal packages with underscores matching `<parent>_<child>`:
```go
import (
    dao_cache "github.com/Fiagram/standalone/internal/dao/cache"
    dao_database "github.com/Fiagram/standalone/internal/dao/database"
    logic_account "github.com/Fiagram/standalone/internal/logic/account"
    http_handler "github.com/Fiagram/standalone/internal/handler/http"
    http_logic "github.com/Fiagram/standalone/internal/logic/http"
    token_logic "github.com/Fiagram/standalone/internal/logic/token"
)
```
- Standard library imports first, then external, then internal (groups separated by blank lines)

### Files
- `snake_case.go`: `account_password.go`, `refresh_token.go`, `chatbot_webhook.go`
- Dedicated files per concern: `account.go` (interface + impl), `account_types.go` (param/output types), `errors.go` (sentinel errors)
- `module.go` in every layer package for fx registration

### Types
| Element | Convention | Example |
|---------|-----------|---------|
| Exported interface | PascalCase, noun/role | `AccountAccessor`, `Client`, `Token`, `HttpServer` |
| Unexported impl struct | camelCase, matches interface | `accountAccessor`, `ramClient`, `account` |
| Param structs | `<Action>Params` | `CreateAccountParams`, `GetAccountParams` |
| Output structs | `<Action>Output` | `CreateAccountOutput`, `CheckAccountValidOutput` |
| DAO entity structs | PascalCase singular | `Account`, `AccountPassword`, `ChatbotWebhook` |
| Enum types | PascalCase type alias | `type Role uint8`, `type CacheClientType string` |
| Enum constants | PascalCase values | `None`, `Admin`, `Member`, `CacheTypeRam`, `CacheTypeRedis` |

### Functions & Methods
| Pattern | Convention | Example |
|---------|-----------|---------|
| Constructor | `New<Type>` | `NewAccountAccessor()`, `NewHash()`, `NewDaoCache()` |
| Config extractor | `GetConfig<Name>` | `GetConfigHttp()`, `GetConfigAuth()` |
| CRUD | Verb-first | `CreateAccount()`, `GetAccount()`, `UpdateAccount()`, `DeleteAccount()` |
| Boolean query | `Is<Noun>` or `Has` | `IsUsernameTaken()`, `IsMember()` |
| Init + setup | `InitAnd<Action>` | `InitAndMigrateUpDatabase()` |

### Variables
- Short names in tight scopes: `c`, `db`, `acc`, `tx`, `err`, `ctx`
- Receiver names: single letter matching the type — `func (a accountAccessor)`, `func (c ramClient)`
- Package-level const keys: `camelCase` — `usernamesTakenKey = "usernames_taken"`

## Interface Design

- Small, focused interfaces with single responsibility
- `context.Context` as first parameter in every method
- `error` as last return value
- Use Param/Output structs for methods with >2 parameters or return values

```go
type AccountAccessor interface {
    CreateAccount(ctx context.Context, account Account) (uint64, error)
    GetAccount(ctx context.Context, id uint64) (Account, error)
    UpdateAccount(ctx context.Context, account Account) error
    DeleteAccount(ctx context.Context, id uint64) error
    WithExecutor(exec Executor) AccountAccessor
}
```

### Compile-time Interface Checks
Verify implementations satisfy interfaces:
```go
var _ Executor = (*sql.DB)(nil)
var _ Executor = (*sql.Tx)(nil)
```

## Constructor Pattern

Constructors return the interface type, not the concrete struct:

```go
func NewAccountAccessor(
    exec Executor,
    logger *zap.Logger,
) AccountAccessor {
    return &accountAccessor{
        exec:   exec,
        logger: logger,
    }
}
```

- Accept dependencies as parameters (injected by fx)
- Return interface, not `*struct`
- Unexported struct, exported interface

## Error Handling

### Sentinel Errors
Declare in dedicated `errors.go` file per package:
```go
package logic_account

import "fmt"

var (
    ErrTxCommitFailed = fmt.Errorf("failed to commit transaction")
    ErrTxBeginFailed  = fmt.Errorf("failed to begin transaction")
)
```

- Use `var Err<Name> = fmt.Errorf("...")` or `errors.New("...")`
- Prefix with `Err`

### Error Wrapping
Always wrap errors with context:
```go
if err != nil {
    return fmt.Errorf("failed to create account: %w", err)
}
```

### Error Logging
Log before returning, with structured context:
```go
logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("account", acc))
if err != nil {
    logger.With(zap.Error(err)).Error("failed to create account")
    return 0, err
}
```

### Input Validation
Validate at system boundaries (DAO, logic entry points):
```go
if acc.Username == "" && acc.RoleId == 0 {
    return 0, ErrLackOfInfor
}
```

## DAO Layer Patterns

### Database Accessors
- Accept `Executor` interface (works with both `*sql.DB` and `*sql.Tx`)
- Provide `WithExecutor(exec Executor) <Interface>` for transaction support
- Use parameterized queries (`?` placeholders)
- Scan rows into structs
- Check `RowsAffected()` for mutation validation
- Use `LastInsertId()` for auto-increment IDs

```go
const query = `INSERT INTO accounts 
        (username, fullname, email, phone_number, of_role_id) 
        VALUES (?, ?, ?, ?, ?)`
result, err := a.exec.ExecContext(ctx, query,
    acc.Username,
    acc.Fullname,
    acc.Email,
    acc.PhoneNumber,
    acc.RoleId,
)
```

### Cache Accessors
- Generic `Client` interface with `Set`, `Get`, `Delete`, `Has` methods
- Specialized accessors wrap the generic `Client` with domain-specific keys
- Factory pattern: switch on config type (`CacheTypeRam` / `CacheTypeRedis`)
- Key formatting with prefixes: `fmt.Sprintf("refresh_token:%s", token)`
- Const keys for set-based caches: `usernamesTakenKey = "usernames_taken"`

### Transaction Pattern
```go
tx, err := a.db.BeginTx(ctx, &sql.TxOptions{})
if err != nil {
    return ..., ErrTxBeginFailed
}
defer tx.Rollback()

// Use accessor with transaction executor
accessor := a.accessor.WithExecutor(tx)
// ... perform operations ...

if err = tx.Commit(); err != nil {
    return ..., ErrTxCommitFailed
}
```

## Logic Layer Patterns

- Interface defines business operations with Param/Output types
- Implementation struct holds DAO accessors, other logic, logger, and `*sql.DB` for transactions
- Transactions managed at the logic layer, NOT in DAO
- Map between DAO entity types and logic domain types explicitly

## HTTP Handler Patterns

- Implement `oapi.ServerInterface` generated from OpenAPI spec
- Method signature: `func (h *handler) OperationName(c *gin.Context)`
- Extract `accountId` from Gin context (set by auth middleware): `c.MustGet("accountId").(uint64)`
- Return JSON with explicit status codes: `c.JSON(http.StatusOK, response)`
- Error responses use OpenAPI-generated types: `Unauthorized`, `BadRequest`, `InternalServerError`

### Middleware
- Bearer token extraction: `strings.CutPrefix(authHeader, "Bearer ")`
- Set values in Gin context: `c.Set("accountId", claims.AccountId)`
- Call `c.Next()` on success, `c.AbortWithStatusJSON()` on failure

## Logging Patterns

- Use `zap.Logger` everywhere (injected via fx)
- Always context-aware: `logger := logger.LoggerWithContext(ctx, l.logger)`
- Chain structured fields with `.With()`:
```go
logger.With(zap.String("key", val)).
    With(zap.Uint64("id", id)).
    Error("operation failed", zap.Error(err))
```
- Use type-specific zap fields: `zap.String()`, `zap.Uint64()`, `zap.Int()`, `zap.Any()`, `zap.Error()`
- Use `zap.NewNop()` in tests

## Struct Tags

| Context | Tag Style | Example |
|---------|----------|---------|
| JSON (DAO entities) | `snake_case` | `json:"phone_number"` |
| YAML (config) | `camelCase` | `yaml:"phoneNumber"` |
| Omit empty | with `omitempty` | `json:"field,omitempty"` |

## Configuration Pattern

- Struct-based config with YAML tags
- `//go:embed local.yaml` for default config fallback
- Type aliases for config values: `type ConfigFilePath string`, `type CacheClientType string`
- Method receivers for computed properties on config types:
```go
func (c ConfigHttpCookie) SameSite() http.SameSite { ... }
```
- Individual `GetConfig<Name>(config Config) Config<Name>` functions for fx injection

## Test Patterns

### Test File Location
Tests live in a separate `test/` directory tree mirroring `internal/`:
```
test/dao/database/     → tests for internal/dao/database/
test/logic/account/    → tests for internal/logic/account/
test/logic/http/       → tests for internal/logic/http/
```

### Package Naming
Test packages use `_test` suffix: `package dao_database_test`, `package logic_account_test`

### Test Setup
Use `TestMain(m *testing.M)` for global setup (config, DB, migrations):
```go
func TestMain(m *testing.M) {
    config, err := configs.NewConfig("")
    if err != nil {
        log.Fatal("failed to init config")
    }
    logger = zap.NewNop()
    db, cleanup, err := dao_database.InitAndMigrateUpDatabase(config.DatabaseClient, logger)
    if err != nil {
        log.Fatal("failed to init database")
    }
    defer cleanup()
    sqlDb = db
    os.Exit(m.Run())
}
```

### Test Functions
- Individual functions per scenario (not table-driven):
  - `TestCreateAccount(t *testing.T)`
  - `TestGetAccountById(t *testing.T)`
- Create-test-delete pattern for isolation
- Use `require` (not `assert`) from testify:
  - `require.NoError(t, err)`
  - `require.Equal(t, expected, actual)`
  - `require.NotEmpty(t, value)`

### Test Utilities
Random data generators in `utils_test.go`:
- `RandomString(length uint) string`
- `RandomVnPersonName() string`
- `RandomGmailAddress() string`
- `RandomAccount() dao_database.Account`

### Run Tests
```bash
make test
```

## Utility Functions

Generic helpers in `internal/utils/quick_func.go`:
```go
func If[T any](condition bool, trueVal, falseVal T) T
func Ptr[T any](v T) *T
```

## Constraints
- DO NOT edit files in `internal/generated/` — these are auto-generated
- DO NOT use `assert` from testify — use `require` (fails immediately)
- DO NOT put transactions in the DAO layer — manage them in logic
- DO NOT use `fmt.Println` for logging — use `zap.Logger`
- DO NOT inline complex schemas — use Param/Output struct types
- DO NOT export implementation structs — export interfaces, keep structs unexported
- DO NOT skip `context.Context` as first parameter on any DAO or logic method
- DO NOT import packages from `cmd/` into `internal/`

## Procedure for New Feature
1. Read existing code in the relevant layer(s) to understand current patterns
2. If new DAO accessor: create interface + unexported impl in `internal/dao/<sub>/`, register in `internal/dao/module.go`
3. If new logic: create interface + types + impl in `internal/logic/<sub>/`, register in `internal/logic/module.go`
4. If new HTTP endpoint: update `docs/openapi.yml`, run `make generate`, implement handler method
5. If new config: add struct + YAML tags in `internal/configs/`, add `GetConfig<Name>` and register in module
6. Register all new constructors in the appropriate `module.go`
7. Write tests in `test/` mirroring the internal path
8. Run `make test` and `make lint` to verify
