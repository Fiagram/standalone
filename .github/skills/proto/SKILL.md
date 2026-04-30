---
name: proto
description: "Write and maintain Protocol Buffer (proto3) specifications. Use when adding new RPC methods, services, messages, or enums to .proto files. Use when asked about gRPC API design, proto naming conventions, or how to regenerate gRPC stubs."
---

# Protocol Buffer Specification Writing

## When to Use
- Adding new RPC methods to an existing service in `api/*.proto`
- Adding new request/response message types
- Adding or modifying enum values
- Adding new `optional` or `repeated` fields to existing messages
- Reviewing or fixing proto naming/numbering issues
- Regenerating gRPC stubs after proto changes

## Project Setup
- Proto syntax: `proto3`
- Proto source directory: `api/`
- Main file: `api/strategy.proto`
- Package: `fiagram.strategy`
- Go package option: `go_package = "grpc/strategy"`
- Code generators: `protoc-gen-go` + `protoc-gen-go-grpc`
- Generated output: `internal/generated/grpc/strategy/`
- Regenerate command: `make generate`
- Required imports: `google/protobuf/timestamp.proto` for timestamps

## Naming Conventions

| Element | Style | Examples |
|---------|-------|---------|
| Package | `lowercase.dotted` | `fiagram.strategy` |
| Service | PascalCase | `Strategy` |
| RPC method | PascalCase verb+noun | `CreateAlert`, `GetAlerts`, `UpdateAlert`, `DeleteAlert` |
| Message | PascalCase, no suffix | `Alert`, `CreateAlertRequest`, `GetAlertsResponse` |
| Fields | `snake_case` | `of_account_id`, `created_at`, `alert_id` |
| Enums | PascalCase, **top-level** (not nested) | `Timeframe`, `Operator`, `Price`, `BollingerBand` |
| Enum values | `UPPER_SNAKE_CASE`, prefixed with enum name | `TIMEFRAME_D1`, `OPERATOR_GREATER_THAN`, `PRICE_CLOSE` |

## Request/Response Message Conventions
- Every RPC MUST have a dedicated `XxxRequest` and `XxxResponse` pair
- Never reuse domain messages (e.g., `Alert`) directly as request/response types
- Response messages wrap the domain message (e.g., `CreateAlertResponse { Alert alert = 1; }`)
- List RPCs: response contains `repeated DomainType items = 1;`
- Delete RPCs: response echoes back the IDs that were deleted (`of_account_id`, `alert_id`)
- Update RPCs: response contains the fully updated domain message

## Field Numbering Rules
- Field numbers are permanent — NEVER reuse or renumber existing fields
- Add new fields by continuing the sequence from the last used number
- Start at `1` and increment sequentially
- Reserve removed field numbers with `reserved` keyword to prevent accidental reuse

## Field Type Guidelines
- Use `uint64` for IDs and counts
- Use `uint32` for smaller counts (e.g., pagination `limit`, `offset`)
- Use `int64` for timestamps stored as Unix epoch (e.g., `exp`)
- Use `google.protobuf.Timestamp` for structured timestamps (`created_at`, `updated_at`)
- Use `optional string` for nullable string fields (e.g., `message`)
- Use `repeated MessageType` for lists (e.g., `repeated Alert alerts`)
- Use nested enum types for closed sets of domain values

## Enum Conventions
- Always define enums at the **top level** of the file, not nested inside message types
- Always include a zero-value `NONE` as the first value (proto3 default)
- Zero value MUST be `<PREFIX>_NONE = 0` (e.g., `TIMEFRAME_NONE = 0`, `OPERATOR_NONE = 0`)
- Prefix every enum value with the enum type name in `UPPER_SNAKE_CASE` (e.g., `Price` → `PRICE_OPEN`, `PRICE_CLOSE`)
- Single-value enums (e.g., `RelativeStrengthIndex`, `Volume`) are valid — they allow future extension without breaking the wire format

## Operand / Union Field Pattern
When a field can hold one of several distinct indicator types (plus an optional constant), use a `oneof` message instead of a flat enum:
```protobuf
message Operand {
  oneof value {
    Price                 price                  = 1;
    BollingerBand         bollinger_band          = 2;
    SimpleMovingAverage   simple_moving_average   = 3;
    RelativeStrengthIndex relative_strength_index = 4;
    Volume                volume                  = 5;
    double                const_value             = 6;
  }
}
```
- Each enum type in the `oneof` categorises operands as *price-based* (`Price`, `BollingerBand`, `SimpleMovingAverage`) or *niche-based* (`RelativeStrengthIndex`, `Volume`)
- The `double const_value` arm represents a plain numeric constant
- Cross-field compatibility rules (e.g., price-based op1 can only pair with price-based or const op2) are enforced in the logic layer, not the proto
- Use this pattern whenever a field is semantically a "one of several indicator categories"

## Service Design
- One service per domain area (e.g., `Strategy` owns all alert RPCs)
- Provide full CRUD: `Create`, `Get` (single), `GetXxx` (list), `Update`, `Delete`
- List RPCs should accept `of_account_id`, `limit`, `offset` for pagination
- Single-item RPCs should accept both `of_account_id` and `<entity>_id` for authorization

## Constraints
- DO NOT use `proto2` syntax — always `proto3`
- DO NOT use `string` for IDs — use `uint64`
- DO NOT define RPC methods that accept or return domain types directly
- DO NOT reuse or renumber removed field numbers (use `reserved` instead)
- DO NOT add fields without a comment if the purpose is non-obvious
- ONLY use well-known types from `google/protobuf/` (timestamp, duration, empty, etc.)

## Procedure
1. Read the existing `api/strategy.proto` to understand current service and message structure
2. Identify the correct service to extend, or create a new service file in `api/` if the domain is distinct
3. Add new messages following the naming and field conventions above
4. Add new RPC methods to the service block
5. Verify all field numbers are unique and sequential within each message
6. Run `make generate` to regenerate stubs and confirm no build errors:
   ```bash
   make generate
   ```
7. Check `internal/generated/grpc/strategy/` to confirm the expected Go types and interfaces were generated
8. If a new service was added, implement the client wrapper in `internal/dao/strategy/client.go` using the `WithExecutor` pattern

## Output Format
- Standard `proto3` syntax
- 2-space indentation
- One blank line between message/enum/service blocks
- Keep field comments on the same line or the line above
