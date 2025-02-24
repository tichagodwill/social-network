<script lang="ts">
    import '../app.css'
    import { onMount } from 'svelte'
    import { browser } from '$app/environment'
    import { auth } from '$lib/stores/auth'
    import { Spinner } from 'flowbite-svelte'
    import Navbar from '$lib/components/Navbar.svelte'
    import WebSocketProvider from '$lib/components/WebSocketProvider.svelte'
    import Toast from '$lib/components/Toast.svelte'

    let loading = true;
    export const data: {} = {}; // Initialize with an appropriate value

    onMount(async () => {
        try {
            await auth.initialize();
            // Note: We don't need to request notification permission here anymore
            // as it's now handled by the WebSocketProvider component
        } finally {
            loading = false;
        }
    });
</script>

<!-- Always include Toast component for all pages -->
<Toast />

{#if browser && $auth.isAuthenticated && !loading}
    <!-- Only initialize WebSocket when user is authenticated and page is loaded -->
    <WebSocketProvider />
{/if}

<div class="min-h-100 bg-gray-50 dark:bg-gray-900">
    {#if loading}
        <div class="flex items-center justify-center min-h-screen">
            <Spinner size="12" />
        </div>
    {:else}
        <Navbar />
        <main class="container mx-auto px-4 py-8">
            <slot />
        </main>
    {/if}
</div>