# Suggestions: `cmd/api/handlers.go`

Based on the `developing-genkit-go` skill (`.agents/skills/developing-genkit-go/SKILL.md`).

## TL;DR

The current `generateAIResponseHandler` is doing three jobs at once — SSE framing,
stream consumption, and hand-rolled tool-request JSON reassembly — and none of it is
traceable. The skill's core guidance is:

> **Wrap AI logic in flows.** Flows give you tracing, observability, HTTP deployment via
> `genkit.Handler`, and the ability to test from the Developer UI and CLI. Any generation
> call worth keeping should live in a flow.

Recommended shape:

1. Move the entire generate loop into a **`DefineStreamingFlow`** that streams
   `EventMessage` values (the flow becomes the traced, testable unit).
2. Make the Echo handler **thin**: bind the request, set SSE headers, and pump the flow's
   stream into the response.
3. Replace the **error-string sniffing** tool-request buffer with a robust accumulator
   (prefer the SDK's `Partial` flag, fall back to `json.Valid`).

---

## 1. Wrap the AI logic in a streaming flow (primary change)

Define the flow in `ai.go` (replacing/augmenting `generateResponse`). The flow owns the
model call, the tool-request reassembly, and emits `EventMessage` events to the stream
callback. `EventMessage` stays exactly as it is.

```go
// ai.go

// Define on the application (or return the *core.Flow) so routes can register it:
func (app *application) generateFlow() *core.Flow[AIRequest, string, EventMessage] {
	return genkit.DefineStreamingFlow(app.genkit, "generate",
		func(ctx context.Context, req AIRequest, sendChunk core.StreamCallback[EventMessage]) (string, error) {
			stream := genkit.GenerateStream(ctx, app.genkit,
				ai.WithPrompt(req.Question),
				ai.WithModelName("deepseek/deepseek-v4-flash"),
				ai.WithTools(app.searchTool),
			)
			return streamGenerateEvents(ctx, stream, sendChunk)
		},
	)
}

// streamGenerateEvents: one place that turns model chunks into client events.
// This is now testable standalone and shows up in Dev UI traces as a flow.
func streamGenerateEvents(ctx context.Context,
	stream iter.Seq2[*ai.ModelStreamValue, error],
	sendChunk core.StreamCallback[EventMessage],
) (string, error) {

	buffers := make(map[string]*strings.Builder)

	for result, err := range stream {
		if err != nil {
			return "", err // flow fails -> visible in trace, handler emits "error" event
		}
		if result.Done {
			return result.Response.Text(), nil
		}

		for _, part := range result.Chunk.Content {
			var ev EventMessage
			switch {
			case part.IsReasoning():
				ev = EventMessage{eventType: "reasoning", data: part.Text}
			case part.IsText():
				ev = EventMessage{eventType: "content", data: part.Text}
			case part.IsToolRequest():
				ev, ok = accumulateToolRequest(part.ToolRequest, buffers)
				if !ok {
					continue // still assembling JSON fragments
				}
			case part.IsToolResponse():
				ev = EventMessage{eventType: "tool-response", data: part.ToolResponse.Name}
			default:
				return "", fmt.Errorf("unexpected part kind: %v", part.Kind)
			}
			if err := sendChunk(ctx, ev); err != nil {
				return "", err
			}
		}
	}
	return "", nil
}
```

Why: `GenerateStream` in a flow gives you traces of every model call, tool request, and
latency in the Dev UI (`genkit start -- go run .`), and lets you invoke/iterate on it
non-interactively with `genkit flow:run generate '{"data":{"question":"..."}}' -- go run .`
— no HTTP round-trip needed while developing.

## 2. Make the Echo handler thin

```go
// handlers.go
func (app *application) generateAIResponseHandler(c *echo.Context) error {
	var req AIRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
	}
	if req.Question == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "question is required"})
	}

	rsp := c.Response()
	rsp.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	rsp.Header().Set("Cache-Control", "no-cache")
	rsp.Header().Set("Connection", "keep-alive")
	rsp.Header().Set("X-Accel-Buffering", "no")
	rc := http.NewResponseController(rsp)

	// Cancel the flow when the client disconnects so we stop hitting the model.
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	for v, err := range app.generateFlow().Stream(ctx, req) {
		if err != nil {
			c.Logger().Error("generate flow failed", "err", err)
			fmt.Fprint(rsp, (&EventMessage{eventType: "error", data: err.Error()}).String())
			rc.Flush()
			return nil
		}
		if v.Done {
			break
		}
		fmt.Fprint(rsp, v.Stream.String())
		if err := rc.Flush(); err != nil {
			return nil // client went away; ctx cancel stops the flow
		}
	}
	return nil
}
```

Notes:

- `Flow.Stream(ctx, input)` returns an iterator that stops when `ctx` is cancelled, so a
  client disconnect actually aborts generation (today it just logs and keeps iterating).
- On failure, send an `event: error` frame instead of `return nil` — the client currently
  gets a silently truncated stream with no explanation.
- Return the `Bind` error as JSON with a message; `c.JSON(400, "bad request")` sends a
  bare string which is awkward for clients.

## 3. Fix tool-request reassembly (remove error-string sniffing)

The current buffer logic matches on `"unexpected end of JSON input"`,
`"invalid character"`, and `"cannot unmarshal"` — brittle, provider-dependent, and it
also has a dead special case (`searchInputString == "}"`). The SDK marks streaming
tool-request fragments with `ToolRequest.Partial` (see `ai.ToolRequest` in gen.go), and
the final chunk has `Partial == false`. Buffer per-`Ref` until the final chunk, then
unmarshal once:

```go
func accumulateToolRequest(tr *ai.ToolRequest, buffers map[string]*strings.Builder) (EventMessage, bool) {
	if tr == nil {
		return EventMessage{}, false
	}
	frag, ok := tr.Input.(string)
	if !ok {
		return EventMessage{eventType: "tool-request", data: fmt.Sprint(tr.Input)}, true
	}

	buf := buffers[tr.Ref]
	if buf == nil {
		buf = &strings.Builder{}
		buffers[tr.Ref] = buf
	}
	buf.WriteString(frag)

	// Prefer the SDK's Partial flag; fall back to json.Valid for providers that
	// don't set it.
	complete := !tr.Partial
	if complete && !json.Valid([]byte(buf.String())) {
		complete = false
	}
	if !complete {
		return EventMessage{}, false // keep accumulating
	}

	value, err := FromString(buf.String())
	delete(buffers, tr.Ref)
	if err != nil {
		return EventMessage{}, true // let the client see the raw input
	}

	var x strings.Builder
	for _, q := range value.SearchQueries {
		fmt.Fprintf(&x, "search(%s)", q)
	}
	return EventMessage{eventType: "tool-request", data: x.String()}, true
}
```

(Keep `FromString` in `tools.go` — nice reuse.)

## 4. Optional: skip the hand-rolled SSE entirely with `genkit.Handler`

If the client can consume Genkit's SSE format (streaming flows are already served as
SSE), you don't need any of the framing code in the handler. Echo can wrap it directly:

```go
// in setupRoutes()
app.echo.POST("/generate", echo.WrapHandler(genkit.Handler(app.generateFlow())))

// request body shape changes to: {"data": {"question": "..."}}
// response is: {"result": ...} / SSE chunks
```

`genkit.Handler` also accepts `genkit.WithContextProviders(...)` if you later need to
inject auth headers into flow context. You lose the custom `reasoning`/`tool-request`
event types in this mode, so only adopt it if the client can be updated.

## 5. Smaller fixes

- **Model ID**: `deepseek/deepseek-v4-flash` is hardcoded. The skill says model names
  change frequently — read it from an env var (`DEEPSEEK_MODEL`) with the current ID as
  default, and check provider docs before bumping.
- **Drop the `getRoutes`/`postRoutes` maps**: three lines of indirection for two routes;
  register them directly. Not AI-related, but it reduces the surface area.
- **`generateResponse`**: remove it once the flow exists — raw `iter.Seq2` returned to a
  handler is what made the handler fat in the first place.
- **Good already**: `SearchInput` has `jsonschema:"description=..."` tags (the skill
  explicitly calls this out as required for structured quality), and `g` is passed
  explicitly rather than stashed in a global. Keep both.

## 6. How to run and verify

```bash
genkit start -- go run .            # capture traces; Dev UI at :4000 shows the "generate" flow
genkit flow:run generate '{"data": {"question": "test"}}' -- go run .   # headless check
genkit trace:list && genkit trace:get <traceId>   # inspect model I/O and tool calls
```

`go run .` alone is "debugging blind" — flows only show up in the Dev UI/traces when
started under `genkit start`.
