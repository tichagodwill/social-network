<script lang="ts">
    import "../app.css";
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import { auth } from '$lib/stores/auth';
    import { Spinner } from 'flowbite-svelte';
    import Navbar from '$lib/components/Navbar.svelte';
    import WebSocketInitializer from '$lib/components/WebSocketInitializer.svelte';
    import { requestNotificationPermission } from '$lib/stores/websocket';
    import type { LayoutData } from './$types';

    let loading = true;
    export const data: {} = {}; // Initialize with an appropriate value

    onMount(async () => {
        try {
            await auth.initialize();

            // Request notifications permission if user is authenticated
            if (browser && $auth.isAuthenticated) {
                requestNotificationPermission();
            }
        } finally {
            loading = false;
        }
    });
</script>

{#if browser && $auth.isAuthenticated && !loading}
    <!-- Only initialize WebSocket when user is authenticated and page is loaded -->
    <WebSocketInitializer />
{/if}

<div class="min-h-100 bg-gray-50 dark:bg-gray-900">
    {#if loading}
        <div class="flex items-center justify-center min-h-screen">
            <Spinner size="12" />
        </div>
    {:else}
        <Navbar />
        <slot />
    {/if}
</div>