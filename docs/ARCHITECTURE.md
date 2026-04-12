# Architecture

## Overview

smart-chat is a modular Go backend that exposes chatbot and internal operations over Gin. The runtime is centered around a PostgreSQL-backed conversation model, with large-language-model execution layered behind service abstractions and multiple external integrations.

At a high level, the system has four main slices:

- public chat APIs
- internal/admin APIs
- conversation and LLM orchestration
- integration and side-effect services

## Runtime Composition

`cmd/main.go` wires the application in this order:

1. memcache initialization
2. config loading
3. database connection
4. SQL migration execution
5. GORM automigration
6. external client construction
7. service construction
8. route registration
9. cron registration
10. HTTP server startup

This means most system behavior is assembled manually in one bootstrap file rather than through dependency injection containers.

## API Layers

### Legacy `v1`

The `v1` layer preserves the earlier chat flow:

- auth routes from `auth/`
- chat endpoints under `/v1/chat`
- middleware that reads a conversation object from `internal/store`

This path appears to exist for backward compatibility and is conceptually separate from the PostgreSQL-backed `v2` session flow.

### Session-backed `v2/chat`

The `v2/chat` layer is the main persisted chat surface:

- `AuthSessionMiddleware` validates the `Authorization` header against the `sessions` table
- handlers convert requests into service calls
- `ConversationService` delegates to a `ConversationReceiver`
- the conversation subsystem builds prompts, loads history, executes LLM/tool logic, persists outputs, and returns a response

WhatsApp mode is a request-level variation that changes downstream behavior and notification side effects.

### Internal `v2/client`

The `v2/client` layer supports internal operations:

- conversation detail and filtered lists
- analytics endpoints
- agent/admin lookup
- manual message insertion by a human agent
- assignment linking between auth users and conversations
- assignment tracking updates with Slack notification side effects

Token validation for these routes is delegated to the auth-service integration in `internal/authservice/zitadel` rather than local JWT parsing.

## Core Domain Model

Important persisted entities include:

- `User`
- `Session`
- `Conversation`
- `MessagePair`
- `FunctionCall`
- `ConvAnalysis`
- `AuthUser`
- `AuthRole`
- `AuthUserConversation`

Key relationships:

- a `Session` belongs to a `User`
- a `Conversation` belongs to a `Session`
- a `Conversation` has many `MessagePair` rows and `FunctionCall` rows
- `AuthUserConversation` links internal auth users to conversations for ownership/assignment

The `AuthUserConversation` table now also stores operational tracking state:

- `started`
- `resolved`
- `comments`

## Conversation Execution Flow

The conversation subsystem is split into several focused types under `internal/services/conversation/`.

Key roles visible from the wiring:

- builder: prepares conversation inputs
- history loader: reconstructs prior context
- executor: runs model and tool logic
- state manager: persists or transitions conversation state
- receiver: orchestrates the above pieces

The request flow is roughly:

1. HTTP handler receives a start or message request.
2. Session middleware resolves the active session.
3. `ConversationService.HandleSession(...)` delegates to the receiver.
4. The receiver prepares history and state.
5. The LLM layer chooses between assistant content and a function/tool call.
6. Tool execution may invoke Indian Travellers APIs.
7. New message pairs and function-call data are persisted.
8. The final response is returned to the handler.
9. Notification and Slack side effects may run asynchronously.

## LLM Layer

The LLM implementation lives in `internal/llm_service/`.

Observed characteristics:

- OpenAI client is initialized from application config
- GPT-4o is used for chat completion
- tool calling is enabled for travel-specific actions
- JSON-schema response formats are used in v2 flows
- WhatsApp output is normalized into text-oriented content after markdown-to-text conversion

Current tool/function schemas include operations such as:

- package detail lookup
- user initial query creation
- final booking creation
- upcoming trips lookup

This design keeps the handler layer free of model-specific branching, but it also means prompt and tool behavior is concentrated in the LLM package and the conversation executor path.

## External Integrations

### Indian Travellers API

`external/indian_travellers` wraps the upstream business system. It provides package, trip, workflow, and booking-related calls used by the LLM tool flow.

### Notification Service

`external/notification` sends message and user events to another service, with simple retry logic built into the client.

### Slack

Slack is used for at least two classes of events:

- alerting on failures or unexpected states
- notifications for new conversations and assignment tracking updates

### Auth Service / Zitadel

Internal endpoints validate bearer tokens through the auth-service integration. This keeps identity resolution centralized outside this repository.

### AWS SSM

Production configuration is loaded from SSM Parameter Store.

## Background Jobs

The application uses `robfig/cron`.

Current bootstrap code schedules a daily call to `PushConversationsToS3`. There are also cron-job-related packages under `internal/cron_jobs/`, which suggests background analysis and notification workflows exist or are planned even if not all are started from `main.go` right now.

## Testing and Quality Gates

The repository contains handler-focused tests under `tests/test_handlers/`.

Current CI behavior in Bitbucket runs:

- `go test ./tests/test_handlers`
- `go build cmd/main.go`

This provides validation for HTTP behavior and basic compilation, but not broad unit or integration coverage across all internal packages.

## Design Constraints And Risks

- startup migrations are run by scanning every SQL file, so migration idempotency matters
- config loading currently mixes environment-driven production behavior with local defaults in code
- multiple API generations coexist, so changes should avoid breaking the legacy `v1` path unless explicitly intended
- several side effects are asynchronous, which is good for latency but can obscure failure handling
- the conversation subsystem is modular but still tightly assembled in `main.go`, so cross-cutting changes often require tracing several packages

## Change Hotspots

If you need to modify behavior, the most likely hotspots are:

- `cmd/main.go` for wiring and lifecycle
- `internal/routes/routes.go` for API surface changes
- `internal/handlers/` for request and response contracts
- `internal/services/conversation/` for chat execution behavior
- `internal/llm_service/` for prompts, models, and tool schemas
- `internal/services/auth_user_conversation/service.go` for agent assignment and tracking behavior
- `migrations/` and `internal/models/` for schema evolution