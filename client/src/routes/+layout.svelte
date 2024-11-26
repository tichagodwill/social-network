<script lang="ts">
    import "../app.css";
    import { onMount } from 'svelte';
    import { auth } from '$lib/stores/auth';
    import { Spinner } from 'flowbite-svelte';

    let loading = true;

    onMount(async () => {
        try {
            await auth.initialize();
        } finally {
            loading = false;
        }
    });
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    {#if loading}
        <div class="flex items-center justify-center min-h-screen">
            <Spinner size="12" />
        </div>
    {:else}
        <slot />
    {/if}
</div>
