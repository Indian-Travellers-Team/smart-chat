# CLAUDE.md

This file documents repository-specific guidance for contributors and coding agents working in smart-chat.

## Primary Objective

Keep the service stable while evolving the `v2` chat and `v2/client` flows. The repository contains both legacy and current behavior, so changes should be deliberate about which surface they affect.

## High-Level Rules

- preserve existing auth flows unless the task explicitly targets auth behavior
- prefer adding focused changes over broad refactors
- keep `v1` and `v2` behavior separate unless the feature intentionally spans both
- make migrations idempotent because all SQL files are applied at startup
- keep external-service contracts backward compatible where possible

## Where To Look First

- bootstrap and wiring: `cmd/main.go`
- route registration: `internal/routes/routes.go`
- HTTP contract changes: `internal/handlers/`
- chat execution behavior: `internal/services/conversation/`
- LLM prompts and tools: `internal/llm_service/`
- assignment and tracking logic: `internal/services/auth_user_conversation/`
- schemas and persistence: `internal/models/` and `migrations/`
- handler-level regression coverage: `tests/test_handlers/`

## Project-Specific Conventions

### Authentication

- `v1` chat uses the legacy middleware path with `internal/store`
- `v2/chat` uses persisted sessions from the database
- `v2/client` auth is validated through the auth-service Zitadel integration
- for token validation changes, prefer the auth-service integration path rather than adding local JWT verification logic

### Database Changes

- add a SQL migration under `migrations/`
- keep the migration safe to re-run
- update the matching GORM model when the schema changes
- check whether admin/client endpoints or tests need updating after schema work

### Conversation Work

- trace changes through handler, conversation service, LLM execution, persistence, and side effects
- maintain compatibility with WhatsApp-specific behavior when changing response generation
- avoid embedding HTTP logic inside lower-level conversation services

### Internal Client Operations

- the `auth_user_conversation` table is the assignment table used by `/v2/client` endpoints
- tracking fields currently include `started`, `resolved`, and `comments`
- tracking updates are scoped to the authenticated agent/admin assignment row and emit Slack notifications on success

## External Dependencies To Respect

- Indian Travellers API for travel data and booking-related side effects
- notification service for outbound message and user events
- Slack for alerts and operational notifications
- AWS SSM for production config loading
- PostgreSQL as the main persistence layer

## Safe Change Strategy

When changing behavior:

1. identify which API generation is affected
2. confirm the persistence model involved
3. review corresponding handlers and services together
4. add or update handler tests when request/response behavior changes
5. run at least build plus relevant tests

## Validation Commands

Use these first unless the task requires more:

```bash
go build cmd/main.go
go test ./tests/test_handlers
```

## Things To Avoid

- do not assume `v1` and `v2` share the same auth model
- do not add non-idempotent startup migrations
- do not break the shape of existing client/admin JSON contracts without a deliberate migration plan
- do not add more inline secrets or environment-specific values to code if configuration can be externalized instead
- do not bypass the centralized auth-service token validation path for internal endpoints unless explicitly required

## Useful Mental Model

Treat the service as three connected systems:

- a public conversational API
- an internal operations API
- an orchestration layer that turns chat inputs into persisted conversation state plus side effects

Most changes are safer when they stay within one of those boundaries and only cross into the next layer through existing service abstractions.