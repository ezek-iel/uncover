import { fetchEventSource, type EventSourceMessage } from '@microsoft/fetch-event-source';

export function fetchAIResponse(
  payload: object,
  onMessage: (ev: EventSourceMessage) => void,
  onDone: () => void
): AbortController { // or return a cleanup function
  const controller = new AbortController();

  fetchEventSource('http://localhost:4000/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    signal: controller.signal,
    onmessage: onMessage,
    onerror(err) {
      console.error('SSE error:', err);
    },
    onclose: onDone
  });

  return controller; // caller can call .abort()
}

export type AIResponse = { type: "reasoning" | "content" | "tool-request" | "tool-response", message: string }[]