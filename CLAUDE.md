# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go CLI application that wraps Google's Gemini Pro model as a streaming multi-turn chat interface. The system prompt primes the model to act as an expert in fundamental analysis of Indian stock market companies (NSE & BSE), covering financials, holdings, quarterly results, and sector context.

## Commands

```bash
# Run the application
go run main.go

# Build binary
go build -o ai-trading .

# Tidy dependencies
go mod tidy
```

There are no tests in this repository.

## Architecture

The entire application is a single file (`main.go`) with no packages beyond `main`. The flow is:

1. A hardcoded `apiKey` string in `main()` must be replaced with a real Gemini API key before running.
2. A `genai.Client` connects to the Gemini API via `google.golang.org/api/option.WithAPIKey`.
3. `model.StartChat()` creates a stateful `ChatSession` — conversation history is retained across turns automatically by the SDK.
4. Each iteration of the `for` loop reads a line from stdin, sends it via `chat.SendMessageStream`, and streams the response back to stdout using the iterator pattern (`response.Next()` until `iterator.Done`).

## Key Conventions

- **API key**: The placeholder `<API-KEY>` on line 17 of `main.go` must be replaced with a live Google AI Studio / Vertex AI key. Do not commit a real key.
- **Model**: Currently hardcoded to `"gemini-pro"`. To switch models (e.g. `gemini-1.5-pro`), change the string passed to `client.GenerativeModel(...)`.
- **Module name**: The Go module is named `gnai` (see `go.mod`), not `ai-trading`.
- **Streaming**: Responses use `SendMessageStream` + iterator — avoid replacing with the non-streaming `SendMessage` unless batching is explicitly required, as streaming provides better UX for long financial analyses.
