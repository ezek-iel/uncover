<script lang="ts">
    import IconSparklesFill from "~icons/gravity-ui/sparkles-fill";
    import IconPaperPlane from "~icons/gravity-ui/paper-plane";
    import IconLoading from "~icons/svg-spinners/3-dots-bounce";

    let {
        text = $bindable(""),
        onclick,
        isMinimal = false,
        isLoading = false,
    }: {
        text: string;
        onclick: () => void;
        isMinimal?: boolean;
        isLoading?: boolean;
    } = $props();

    function handleKeydown(event: KeyboardEvent) {
        // Enter sends; Shift+Enter inserts a newline.
        if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault();
            if (!isLoading) onclick();
        }
    }
</script>

<div class={["container", isMinimal ? "h-20" : "h-40"]}>
    <textarea
        class="text-input"
        placeholder="Ask anything"
        bind:value={text}
        onkeydown={handleKeydown}
    ></textarea>
    <div class="icon">
        <IconSparklesFill class="text-slate-400" />
    </div>

    <button
        class="send-button"
        onclick={onclick}
        disabled={isLoading}
        aria-label={isLoading ? "Generating…" : "Send message"}
        title={isLoading ? "Generating…" : "Send"}
    >
        {#if isLoading}
            <IconLoading class="size-6" />
        {:else}
            <IconPaperPlane class="size-6" />
        {/if}
    </button>
</div>

<style>
    @reference "../../app.css";

    .container {
        @apply w-2xl relative shadow-xl rounded-xl;
    }

    .icon {
        @apply absolute inset-0 w-fit h-fit left-4 top-4;
    }

    .text-input {
        @apply w-full h-full resize-none rounded-xl pl-13 pr-15 pt-3 border border-slate-300 bg-slate-100 transition-all;
    }

    .send-button {
        @apply p-3 bg-purple-800 text-white rounded-full absolute w-fit h-fit bottom-4 right-4 transition-colors
               enabled:hover:bg-fuchsia-800 enabled:hover:cursor-pointer disabled:cursor-wait disabled:opacity-80;
    }
</style>
