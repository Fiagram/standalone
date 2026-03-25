---
name: openapi
description: "Write and maintain OpenAPI 3.0 specifications. Use when adding new endpoints, schemas, parameters, or responses to openapi.yml. Use when asking about REST API design conventions."
---

# OpenAPI 3.0 Specification Writing

## When to Use
- Adding new API endpoints to `docs/openapi.yml`
- Adding or modifying request/response schemas
- Adding query parameters, path parameters, or headers
- Reviewing or fixing OpenAPI spec issues
- Designing new REST API resources

## Project Setup
- OpenAPI version: `3.0.0`
- Spec file: `docs/openapi.yml`
- Code generator: `oapi-codegen` (configured in `oapi_codegen.yml`)
- Generated output: `internal/generated/openapi/oapi.gen.go`
- Regenerate command: `make generate`

## Structure Conventions
- Organize paths by tag groupings, separated with comment headers (`# ---- TagName`)
- Every operation MUST have: `tags`, `summary`, `operationId`, `description`, `responses`
- Use `$ref` extensively — never inline reusable schemas, parameters, or responses

## Naming Conventions

| Element | Style | Examples |
|---------|-------|---------|
| `operationId` | camelCase | `signUp`, `getProfileMe`, `deleteProfileWebhook` |
| Schema names | PascalCase | `SignupRequest`, `AccessTokenResponse` |
| Properties | camelCase | `phoneNumber`, `isRememberMe`, `accessToken` |
| Path segments | kebab-case nouns | `/auth/token/refresh`, `/profile/webhooks/{webhookId}` |

## Schema Rules
- Use `additionalProperties: false` on all request and response object schemas
- Mark all required fields in the `required` array
- Use `writeOnly: true` for sensitive input fields (e.g., passwords)
- Use `readOnly: true` for server-generated fields (e.g., IDs)
- Include `pattern` regex for validated string types
- Include `minLength`/`maxLength` constraints on strings
- Include `minimum`/`maximum` constraints on numbers
- Include `example` values on leaf schemas
- Include `description` on non-obvious fields
- Extract reusable field types as standalone schemas (e.g., `Password`, `Username`, `Email`)

## Response Conventions
Use shared response refs for common errors:

| Status | Ref |
|--------|-----|
| `400` | `#/components/responses/BadRequest` |
| `401` | `#/components/responses/Unauthorized` |
| `403` | `#/components/responses/Forbidden` |
| `404` | `#/components/responses/NotFound` |
| `429` | `#/components/responses/TooManyRequests` |
| `500` | `#/components/responses/InternalServerError` |

- All error responses use the `ErrorResponse` schema (`code` + `message` + optional `details`)
- Use shorthand ref syntax: `"400": { $ref: "#/components/responses/BadRequest" }`

## Security Conventions
- Default global security: `bearerAuth` (JWT)
- Public endpoints override with `security: []`
- Cookie-based auth uses `RefreshTokenCookie` scheme
- Document `Set-Cookie` headers where refresh tokens are issued

## Pagination Conventions
- Use shared `Limit` and `Offset` query parameters from `#/components/parameters/`
- `Limit` default: 20, max: 100
- `Offset` default: 0

## Components Organization
Order within `components`:
1. `securitySchemes`
2. `parameters`
3. `responses`
4. `schemas`

## Constraints
- DO NOT use OpenAPI 3.1 syntax (e.g., no `type: [string, null]`)
- DO NOT inline schemas that could be reused — always extract to `#/components/schemas/`
- DO NOT omit `operationId` on any operation
- DO NOT use `additionalProperties: true` on request/response schemas unless explicitly needed
- DO NOT add endpoints without specifying all relevant error responses
- ONLY output valid OpenAPI 3.0 YAML

## Output Format
- Valid OpenAPI 3.0 YAML
- Maintain consistent 2-space indentation
- Use YAML block scalars (`|` or `>`) for multi-line descriptions

## Procedure
1. Read the existing `docs/openapi.yml` to understand current state
2. Follow all conventions listed above when adding or modifying content
3. Place new schemas in `components.schemas`, new parameters in `components.parameters`
4. Add the appropriate tag and group paths with existing tagged sections
5. Validate that all `$ref` pointers resolve to existing definitions
6. Run `make generate` to regenerate Go server code and verify success
