<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { browser } from "$app/environment";
  import { auth } from '$lib/stores/auth';
  import {
    initializeWebSocket,
    cleanupWebSocketResources,
    connectionState,
    ConnectionState,
    requestNotificationPermission
  } from '$lib/stores/websocket';
  import { toast } from '$lib/stores/toast';  // Import the toast store instead of individual methods

  // Reconnection logic
  let reconnectInterval: number | null = null;
  let lastConnectionState = ConnectionState.CLOSED;
  let notificationPermissionRequested = false;

  let currentState: ConnectionState;
  connectionState.subscribe(state => {
    currentState = state;
  });

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      // Try to reconnect if the connection is closed or in error state
      if (
        currentState === ConnectionState.CLOSED ||
        currentState === ConnectionState.ERROR
      ) {
        initializeWebSocket();
      }
    }
  }

  // Watch connection state and show appropriate notifications
  $: if (currentState !== lastConnectionState) {
    if (currentState === ConnectionState.OPEN && lastConnectionState !== ConnectionState.CONNECTING) {
      toast.success('Connected to server');
    } else if (currentState === ConnectionState.ERROR) {
      toast.error('Connection to server failed');
    } else if (currentState === ConnectionState.CLOSED && lastConnectionState === ConnectionState.OPEN) {
      toast.info('Disconnected from server. Reconnecting...');
    }

    lastConnectionState = currentState;
  }

  // Watch connection state for reconnection logic
  $: if (currentState === ConnectionState.ERROR || currentState === ConnectionState.CLOSED) {
    startReconnectTimer();
  } else if (currentState === ConnectionState.OPEN) {
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

  // Request notification permission
  async function setupNotifications() {
    if (!notificationPermissionRequested && browser) {
      notificationPermissionRequested = true;
      const granted = await requestNotificationPermission();
      if (granted) {
        toast.success('Notification permission granted');
      } else {
        toast.info('Enable notifications to receive updates when the app is in the background');
      }
    }
  }

  // Initialize WebSocket on authentication change
  $: if (browser && $auth.isAuthenticated && currentState === ConnectionState.CLOSED) {
    initializeWebSocket();
    setupNotifications();
  }

  onMount(() => {
    // Only initialize if user is logged in
    if (browser && $auth.isAuthenticated) {
      initializeWebSocket();
      setupNotifications();

      // Add visibility change listener for reconnection
      document.addEventListener('visibilitychange', handleVisibilityChange);
    }
  });

  onDestroy(() => {
    document.removeEventListener('visibilitychange', handleVisibilityChange);
    stopReconnectTimer();
    if (currentState === ConnectionState.OPEN) {
      cleanupWebSocketResources();
    }
  });
</script>

{#if currentState === ConnectionState.CONNECTING}
  <div>Connecting...</div>
{:else if currentState === ConnectionState.ERROR}
  <div>Connection error. Attempting to reconnect...</div>
{/if}

<!-- This is a utility component with no UI -->