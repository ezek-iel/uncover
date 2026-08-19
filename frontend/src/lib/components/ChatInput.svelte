<script lang="ts">
    import IconSparklesFill from "~icons/gravity-ui/sparkles-fill";
    import IconPaperPlane from "~icons/gravity-ui/paper-plane";
</script>

<div class="container">
    <textarea class="text-input" placeholder="Ask anything"></textarea>
    <div class="icon">
        <IconSparklesFill  class="text-slate-400"/>
    </div>

    <button class="send-button">
        <IconPaperPlane class="size-6" />
    </button>
</div>

<style>
    @reference "../../app.css";

    .container {
        @apply w-2xl h-40 relative shadow-xl rounded-xl;
    }

    .icon {
        @apply absolute inset-0 w-fit h-fit left-4 top-4;
    }

    .text-input {
        @apply w-full h-full resize-none rounded-xl  pl-13 pr-15 pt-3 border border-slate-300 bg-slate-100 transition-all;
    }

    .send-button {
        @apply p-3 bg-purple-800 text-white rounded-full absolute w-fit h-fit bottom-4 right-4 hover:bg-fuchsia-800 transition-colors hover:cursor-pointer;
    }
</style>
