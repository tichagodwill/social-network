<!-- src/lib/components/WebSocketInitializer.svelte -->
<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { page } from '$app/stores';
    import { auth } from '$lib/stores/auth';
    import { initializeWebSocket, closeConnection, connectionState, ConnectionState } from '$lib/stores/websocket';
    import {browser} from "$app/environment";

    // Reconnection logic
    let reconnectInterval: number | null = null;

    function handleVisibilityChange() {
        if (document.visibilityState === 'visible') {
            // Try to reconnect if the connection is closed or in error state
            if (
                $connectionState === ConnectionState.CLOSED ||
                $connectionState === ConnectionState.ERROR
            ) {
                initializeWebSocket();
            }
        }
    }

    // Watch connection state
    $: if ($connectionState === ConnectionState.ERROR || $connectionState === ConnectionState.CLOSED) {
        startReconnectTimer();
    } else if ($connectionState === ConnectionState.OPEN) {
        stopReconnectTimer();
    }

    // Start reconnect timer
    function startReconnectTimer() {
        if (reconnectInterval !== null) return;

        reconnectInterval = window.setInterval(() => {
            if (document.visibilityState === 'visible' && $auth.isAuthenticated) {
                console.log('Attempting to reconnect WebSocket...');
                initializeWebSocket();
            }
        }, 10000); // Try every 10 seconds
    }


    // Stop reconnect timer
    function stopReconnectTimer() {
        if (reconnectInterval !== null) {
            window.clearInterval(reconnectInterval);
            reconnectInterval = null;
        }
    }

    // Initialize WebSocket on authentication change
    $: if (browser && $auth.isAuthenticated && $connectionState === ConnectionState.CLOSED) {
        initializeWebSocket();
    }

    onMount(() => {
        // Only initialize if user is logged in
        if (browser && $auth.isAuthenticated) {
            initializeWebSocket();

            // Add visibility change listener for reconnection
            document.addEventListener('visibilitychange', handleVisibilityChange);
        }
    });

    onDestroy(() => {
        document.removeEventListener('visibilitychange', handleVisibilityChange);
        stopReconnectTimer();
        closeConnection();
    });
</script>

<!-- This is a utility component with no UI -->