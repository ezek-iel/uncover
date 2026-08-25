package main

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"os"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// modelName returns the DeepSeek model to use, overridable via DEEPSEEK_MODEL.
// Model IDs change frequently; prefer the env var over a hardcoded name.
func modelName() string {
	if m := os.Getenv("DEEPSEEK_MODEL"); m != "" {
		return m
	}
	return "deepseek/deepseek-v4-flash"
}

// newGenerateFlow wraps the generation in a streaming flow so every model call,
// tool request, and latency is traced (Dev UI, `genkit flow:run`, trace:get).
func newGenerateFlow(app *application) *core.Flow[AIRequest, string, EventMessage] {
	return genkit.DefineStreamingFlow(app.genkit, "generate",
		func(ctx context.Context, req AIRequest, sendChunk core.StreamCallback[EventMessage]) (string, error) {
			stream := genkit.GenerateStream(ctx, app.genkit,
				ai.WithPrompt(req.Question),
				ai.WithModelName(modelName()),
				ai.WithTools(app.searchTool),
			)
			return streamGenerateEvents(ctx, stream, sendChunk)
		},
	)
}

// streamGenerateEvents turns model chunks into client events and streams them
// via sendChunk. It returns the final response text once the stream completes.
func streamGenerateEvents(ctx context.Context, stream iter.Seq2[*ai.ModelStreamValue, error], sendChunk core.StreamCallback[EventMessage]) (string, error) {
	buffers := make(map[string]*strings.Builder)

	for result, err := range stream {
		if err != nil {
			return "", err
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
				var ok bool
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

// accumulateToolRequest buffers streaming tool-input fragments per call Ref and
// emits one complete tool-request event once the JSON input has fully arrived.
// It prefers the SDK's Partial flag; json.Valid is a fallback for providers that
// don't set it.
func accumulateToolRequest(tr *ai.ToolRequest, buffers map[string]*strings.Builder) (EventMessage, bool) {
	if tr == nil {
		return EventMessage{}, false
	}

	frag, ok := tr.Input.(string)
	if !ok {
		// Non-fragmented input (e.g. an object); emit as-is.
		return EventMessage{eventType: "tool-request", data: fmt.Sprint(tr.Input)}, true
	}

	buf := buffers[tr.Ref]
	if buf == nil {
		buf = &strings.Builder{}
		buffers[tr.Ref] = buf
	}
	buf.WriteString(frag)

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
		return EventMessage{eventType: "tool-request", data: buf.String()}, true
	}

	var x strings.Builder
	for _, q := range value.SearchQueries {
		fmt.Fprintf(&x, "search(%s)\t", q)
	}
	return EventMessage{eventType: "tool-request", data: x.String()}, true
}

type EventMessage struct {
	eventType string
	data      string
}

func (e *EventMessage) String() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "event: %s\n", e.eventType)
	// SSE allows one logical event to carry multiple `data:` lines; per the spec
	// the client joins them with a single `\n`. This preserves newlines in the
	// payload losslessly, so the frontend can simply concatenate streamed tokens
	// instead of trying to reconstruct where newlines were split out.
	for _, line := range strings.Split(e.data, "\n") {
		fmt.Fprintf(&builder, "data: %s\n", strings.TrimSuffix(line, "\r"))
	}
	builder.WriteString("\n")

	return builder.String()
}
