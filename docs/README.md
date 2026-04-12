# Smart Chat Docs

This folder documents the smart-chat backend used for Indian Travellers conversational flows.

## What This Service Does

The service is a Gin-based Go API that supports two generations of chat APIs:

- `v1` legacy chat endpoints backed by an in-memory/store-based auth flow.
- `v2` session-backed chat endpoints backed by PostgreSQL and GORM.
- `v2/client` operations for internal/admin use cases such as conversation history, analytics, agent lookup, manual agent replies, assignment linking, and assignment tracking.

Core capabilities include:

- conversation start and reply flows
- LLM-backed message generation with function/tool calls
- integration with Indian Travellers APIs for packages, trips, and workflow data
- notification fan-out to a separate notification service
- Slack alerts and notifications
- analytics and conversation-history retrieval for internal users

## Entry Point

The application starts in `cmd/main.go`.

At startup it:

1. initializes memcache
2. loads runtime configuration
3. opens the PostgreSQL connection
4. runs SQL migrations from the `migrations/` directory
5. runs GORM automigrations for key models
6. wires routers, services, middleware, and external clients
7. schedules a daily cron job to push conversations to S3
8. starts the HTTP server on port `8080`

## API Surface

### `v1`

- `GET /v1/chat/ping`
- `POST /v1/chat/send`
- `GET /v1/chat/receive`
- `v1/auth/*` routes registered from the `auth/` package

This path is the older flow and uses `internal/middlewares/authMiddleware.go` with `store.GetConversation(...)`.

### `v2/chat`

- `POST /v2/chat/start`
- `GET /v2/chat/messages`
- `POST /v2/chat/message`

This path is backed by `internal/services/conversation` and uses `internal/middlewares/authSessionMiddleware.go` to resolve a persisted session from the `Authorization` header.

### `v2/client`

- `POST /v2/client/login`
- `GET /v2/client/conversation/:id`
- `GET /v2/client/conversations`
- `GET /v2/client/analytics/dashboard/conversations-summary`
- `GET /v2/client/analytics/conversations/last-30-days`
- `GET /v2/client/agents`
- `GET /v2/client/userdetails`
- `POST /v2/client/add-message`
- `POST /v2/client/conversations/link`
- `PATCH /v2/client/conversations/tracking`

The last two endpoints are backed by the `auth_user_conversation` assignment table. Tracking fields currently include `started`, `resolved`, and `comments`.

## Main Packages

- `cmd/`: application bootstrap
- `config/`: config loading and production SSM lookup
- `auth/`: auth handlers and route registration
- `external/indian_travellers/`: client for packages, trips, workflow, and booking-related calls
- `external/notification/`: client for notification side effects
- `internal/handlers/`: HTTP handlers
- `internal/services/conversation/`: main conversation pipeline orchestration
- `internal/services/conversation_history/`: admin list/detail queries and filtering
- `internal/services/auth_user_conversation/`: assignment linking, agent lookup, and tracking state updates
- `internal/services/human/`: manual agent message injection into a conversation
- `internal/services/analytics/`: reporting and dashboard data
- `internal/services/slack/`: Slack notifications and alerts
- `internal/llm_service/`: model execution and tool/function-call definitions
- `internal/models/`: GORM models
- `internal/routes/`: route registration
- `migrations/`: SQL migrations applied at startup
- `tests/test_handlers/`: handler-level test coverage

## Conversation Pipeline Summary

The `ConversationService` constructs a receiver composed of:

- builder
- executor
- state manager
- history loader

The high-level flow is:

1. auth middleware resolves a session
2. handler calls `ConversationService.HandleSession(...)`
3. receiver loads history and state
4. LLM layer chooses either plain assistant output or a function/tool call
5. downstream services persist message pairs and related records
6. optional notification jobs and Slack updates run around the chat lifecycle

WhatsApp-specific behavior is toggled via the `whatsapp=true` query parameter on relevant `v2/chat` endpoints.

## Configuration Notes

Configuration is loaded from `config/config.go`.

- non-production mode uses local defaults and Gin debug mode
- production mode fetches parameters from AWS SSM in `ap-south-1`
- notable config areas: database, OpenAI, notification service, auth service, Indian Travellers API, Slack, email

For future changes, prefer externalized secrets and environment-specific configuration rather than adding more inline defaults.

## Database and Migrations

The service applies every `.sql` file under `migrations/` during startup and then runs GORM automigrations for selected models.

Recent schema work includes:

- performance indexes
- renaming `auth_user_conversation_links` to `auth_user_conversation`
- adding assignment tracking fields: `started`, `resolved`, `comments`

Because migrations are executed at process start, migration files should be idempotent and safe to re-run.

## External Dependencies

- PostgreSQL via GORM
- Memcache
- OpenAI via `github.com/sashabaranov/go-openai`
- Indian Travellers API
- notification service
- Zitadel-backed auth-service token validation
- Slack webhooks
- AWS SSM in production
- cron for scheduled jobs

## Local Development

Useful commands inferred from the repo:

```bash
go build cmd/main.go
go test ./tests/test_handlers
```

The Bitbucket pipeline currently runs those two checks in parallel.

## Where To Read Next

- `ARCHITECTURE.md` for component relationships and request flow
- `CLAUDE.md` for repo-specific development and contribution guidance