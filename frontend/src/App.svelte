<script lang="ts">
  import Sidebar from "./lib/components/Sidebar.svelte";
  import Topbar from "./lib/components/Topbar.svelte";
  import Chat from "./lib/components/Chat.svelte";
</script>

<div class="flex">
  <Sidebar />

  <main class="w-full h-screen overflow-y-scroll">
    <Topbar />
    <Chat/>
  </main>
</div>
