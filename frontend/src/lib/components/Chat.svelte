<script lang="ts">
    import hero from "../../assets/hero.png";
    import ChatInput from "./ChatInput.svelte";
</script>

<div class="chat-container">
    <img src={hero} alt=""  class="hero"/>

    <div class="title">
        <p>Good Afternoon, Jason.</p>
        <p>What do you want to <span class="gradient">uncover today?</span></p>
    </div>

    <ChatInput/>
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
</style>