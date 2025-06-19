# smart-chat - Indian Travellers Chatbot Backend


This repository contains the backend service for the Indian Travellers chatbot application. The chatbot leverages large language model (LLM) technologies, including function calls, to provide travel-related assistance and recommendations. It is integrated with multiple data sources and communication channels to deliver a seamless travel experience.

## Features

- **Chatbot Backend:** Golang Gin-based REST API backend for handling chatbot conversations.
- **LLM Integration:** Utilizes advanced LLM technologies for natural language understanding and function calling.
- **Multi-source Integration:** 
  - Integrated with Indian Travellers Team website to fetch Package and Trips data.
  - Slack integration for error reporting and notifications.
- **Database:** Uses GORM ORM with PostgreSQL for persistent storage of conversations and related data.
- **Conversation Management:** Supports storing and managing human and agent messages within conversations.
- **Scalable Architecture:** Modular design with internal service layers and clear separation of concerns.

## Architecture Overview

- **Golang Gin:** HTTP web framework for building APIs.
- **GORM:** ORM library for database interaction with PostgreSQL.
- **Slack API:** Used for sending error notifications and alerts.
- **Indian Travellers APIs:** Custom API client integration to fetch travel packages and trip details.
- **Internal Services:** HumanService and other service layers handle business logic and data management.
