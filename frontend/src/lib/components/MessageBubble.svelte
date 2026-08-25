<script lang="ts">
    import type { AIResponse } from "../utils/fetch";
    import { styleAIResponse } from "../utils/actions";

    let {
        messageType,
        message,
    }: { messageType: "user" | "system"; message: AIResponse | string } = $props();

    let inner: HTMLDivElement | undefined = $state();

    // Distance from the bottom the user may be while still "following" the stream.
    const STICKY_THRESHOLD = 64;

    function getScrollableParent(el: HTMLElement): HTMLElement | null {
        let parent: HTMLElement | null = el.parentElement;
        while (parent) {
            const { overflowY } = getComputedStyle(parent);
            if (overflowY === "auto" || overflowY === "scroll" || overflowY === "overlay") {
                return parent;
            }
            parent = parent.parentElement;
        }
        return null;
    }

    $effect(() => {
        // Re-runs whenever the token array grows (streaming), because
        // styleAIResponse iterates `message` inside this effect.
        // User messages are plain strings and never bind `inner`, so they
        // skip the markdown/streaming path entirely.
        if (!inner || typeof message === "string") return;

        const scroller = getScrollableParent(inner);
        // Only keep following the stream if the user hasn't scrolled up to
        // read earlier content (otherwise auto-scroll would yank the view).
        const stickToBottom =
            scroller === null ||
            scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < STICKY_THRESHOLD;

        styleAIResponse(message)(inner);

        // Scroll AFTER the DOM update so scrollHeight reflects the new content.
        if (scroller && stickToBottom) {
            scroller.scrollTop = scroller.scrollHeight;
        }
    });
</script>

<div class="message-bubble-row" data-role={messageType}>
    {#if typeof message === "string"}
        <div class="message-bubble user">
            <p class="user-text">{message}</p>
        </div>
    {:else}
        <div class="message-bubble system">
            <div class="inner" bind:this={inner}></div>
        </div>
    {/if}
</div>

<style>
    @reference "../../app.css";

    .message-bubble.system {
        @apply text-lg;
    }

    .message-bubble.user {
        @apply bg-purple-200 max-w-md p-8 text-lg rounded-4xl;
    }

    .user-text {
        @apply whitespace-pre-wrap;
    }

    .message-bubble-row {
        @apply flex;

        &[data-role="user"] {
            @apply justify-end;
        }
    }

    .inner {
        @apply space-y-8;

        * {
            @apply leading-9;
        }

        /*
         * The markdown HTML is injected via innerHTML, so it has no Svelte scope
         * classes — style it with :global() selectors. Tailwind v4 preflight
         * resets h1-h6 to inherit font-size/weight and zeroes all margins, which
         * is why headings/paragraphs looked like plain text.
         */
        :global(.content h1) { @apply text-3xl font-bold mt-8 mb-4; }
        :global(.content h2) { @apply text-2xl font-bold mt-8 mb-3; }
        :global(.content h3) { @apply text-xl font-bold mt-6 mb-3; }
        :global(.content h4) { @apply text-lg font-semibold mt-6 mb-2; }
        :global(.content h5),
        :global(.content h6) { @apply font-semibold mt-5 mb-2; }

        /* Keep reasoning line breaks visible (the backend preserves them via
           SSE multi-line data fields). */
        :global(.reasoning) { @apply whitespace-pre-wrap text-slate-600 text-base; }

        :global(.content p) { @apply my-3; }
        :global(.content a) { @apply text-purple-700 underline hover:text-fuchsia-700; }

        :global(.content ul) { @apply list-disc pl-6 my-3; }
        :global(.content ol) { @apply list-decimal pl-6 my-3; }
        :global(.content li) { @apply my-1; }

        :global(.content blockquote) { @apply border-l-4 border-slate-300 pl-4 my-3 text-slate-600; }
        :global(.content code) { @apply bg-slate-100 px-1.5 py-0.5 rounded-md text-sm; }
        :global(.content pre) { @apply bg-slate-100 p-4 rounded-lg my-3 overflow-x-auto leading-7; }
        :global(.content pre code) { @apply bg-transparent p-0 text-inherit; }
        :global(.content hr) { @apply my-6 border-slate-300; }
        :global(.content table) { @apply my-3 border-collapse w-full; }
        :global(.content th),
        :global(.content td) { @apply border border-slate-300 px-3 py-1.5; }
    }
</style>
