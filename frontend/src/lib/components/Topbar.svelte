<script lang="ts">
    import IconSearch from "~icons/gravity-ui/magnifier";
    import IconPlus from "~icons/gravity-ui/plus";
</script>

<div class="top-bar">
    <div class="start"></div>
    <div class="end">
        <button class="btn btn-outline"><IconSearch />Search Thread</button>
        <button class="btn btn-primary"><IconPlus />New Thread</button>
    </div>
</div>

<style>
    @reference "../../app.css";

    .top-bar {
        @apply flex items-center justify-between py-4 px-8;
    }

    .end {
        @apply gap-4;
    }
</style>
