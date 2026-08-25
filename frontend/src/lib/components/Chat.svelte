<script lang="ts">
    import type { EventSourceMessage } from "@microsoft/fetch-event-source";
    import hero from "../../assets/hero.png";
    import IconLoading from "~icons/svg-spinners/3-dots-bounce";
    import { fetchAIResponse, type AIResponse } from "../utils/fetch";
    import ChatInput from "./ChatInput.svelte";
    import MessageBubble from "./MessageBubble.svelte";

    type Message = { role: "user" | "system"; content: AIResponse | string };

    let chatInput = $state("");
    let messages = $state<Message[]>([]);
    // Tokens for the in-flight response; `null` when idle. This is the SAME
    // array pushed into `messages` below, so mutating it streams into the
    // history bubble in place (see the note in onSendRequest).
    let current: AIResponse | null = $state(null);

    const hasMessages = $derived(messages.length > 0);
    const isLoading = $derived(current !== null);

    function handleMessage(e: EventSourceMessage) {
        if (!current) return;

        if (
            ["reasoning", "content", "tool-request", "tool-response"].includes(
                e.event,
            )
        ) {
            current.push({
                type: e.event as
                    | "reasoning"
                    | "content"
                    | "tool-request"
                    | "tool-response",
                message: e.data,
            });
        }
    }

    function onSendRequest() {
        const question = chatInput.trim();
        if (question.length < 5 || current) return;

        // Build the response array first, then push the PROXIED reference into
        // history. Svelte 5's state proxies are not cached per target: `proxy(raw)`
        // creates a fresh proxy on every call, so if we pushed the raw array and
        // streamed into a separate `current`, the bubble would never update.
        // Pushing `current` (the same proxy) keeps every reader on one object.
        const response: AIResponse = [];
        current = response;
        messages.push(
            { role: "user", content: question },
            { role: "system", content: current },
        );
        chatInput = "";

        fetchAIResponse({ question }, handleMessage, () => {
            current = null; // stream finished → spinner clears, history stays
        });
    }
</script>

<div class="chat-container">
    {#if !hasMessages}
        <img src={hero} alt="" class="hero" />
        <div class="title">
            <p>Good Afternoon, Jason.</p>
            <p>What do you want to <span class="gradient">uncover today?</span></p>
        </div>
        <ChatInput bind:text={chatInput} onclick={onSendRequest} />
    {:else}
        <div class="message-list">
            {#each messages as message (message)}
                <MessageBubble
                    messageType={message.role}
                    message={message.content}
                />
            {/each}

            {#if isLoading}
                <div class="loading-row">
                    <IconLoading class="size-6 text-purple-700" />
                </div>
            {/if}
        </div>

        <ChatInput
            bind:text={chatInput}
            onclick={onSendRequest}
            isMinimal
            {isLoading}
        />
    {/if}
</div>

<style>
    @reference "../../app.css";

    .chat-container {
        @apply flex flex-col items-center py-20 gap-8;
    }
    .hero {
        @apply size-48;
    }

    .title {
        @apply text-center font-bold text-4xl;
    }

    .gradient {
        @apply bg-clip-text text-transparent bg-linear-270 from-purple-600 to-80% to-fuchsia-600;
    }

    .message-list {
        @apply w-9/12 space-y-12;
    }

    .loading-row {
        @apply flex items-center;
    }
</style>
