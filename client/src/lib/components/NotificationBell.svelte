<script lang="ts">
  import { Bell, Calendar } from 'lucide-svelte';
  import { Button, Dropdown, DropdownItem } from 'flowbite-svelte';
  import { onMount, onDestroy } from 'svelte';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import {
    notifications,
    unreadNotificationsCount,
    markNotificationAsRead,
    markAllNotificationsAsRead
  } from '$lib/stores/websocket';
  import { fade } from 'svelte/transition';

  // Add a toast notification system
  let toastMessage = '';
  let toastVisible = false;
  let toastType: 'success' | 'error' | 'info' = 'info';
  let loading = false;
  let isOpen = false;

  // Function to check if current user is the invitee
  function isInvitee(notification: any): boolean {
    return notification.type === 'group_invitation' &&
      notification.userId === $auth?.user?.id;
  }

  // Function to check if current user is admin/creator of the group
  function isGroupAdmin(notification: any): boolean {
    return notification.type === 'join_request' &&
      (notification.userRole === 'creator' || notification.userRole === 'admin');
  }

  // Function to check if current user is the recipient of the notification
  function isRecipient(notification: any): boolean {
    return notification.userId === $auth?.user?.id;
  }

  function showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
    toastMessage = message;
    toastType = type;
    toastVisible = true;
    setTimeout(() => {
      toastVisible = false;
    }, 3000);
  }

  async function handleInviteResponse(notification: any, action: 'accept' | 'reject') {
    try {
      loading = true;
      
      // First handle the invitation response
      const response = await fetch(`http://localhost:8080/groups/${notification.groupId}/invitations/${notification.invitationId}/${action}`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Failed to ${action} invitation`);
      }

      // Update the notifications list
      notifications.update(notes => 
        notes.map(note => 
          note.id === notification.id 
            ? { ...note, isProcessed: true }
            : note
        )
      );

      showToast(`Successfully ${action}ed invitation`, 'success');

      // Refresh the page if we're on the group page
      if (window.location.pathname.includes(`/groups/${notification.groupId}`)) {
        window.location.reload();
      }

    } catch (error) {
      console.error('Error handling invitation response:', error);
      showToast(error instanceof Error ? error.message : 'Failed to process invitation', 'error');
    } finally {
      loading = false;
    }
  }

  async function handleJoinRequest(notification: any, action: 'accept' | 'reject') {
    try {
      loading = true;
      const response = await fetch(`http://localhost:8080/groups/${notification.groupId}/join-requests/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ requestId: notification.requestId })
      });

      if (!response.ok) {
        throw new Error(`Failed to ${action} request`);
      }

      // Mark notification as read
      await markNotificationAsRead(notification.id);

      // Show success message
      showToast(`Join request ${action}ed successfully`, 'success');

    } catch (error: any) {
      showToast(error.message || `Failed to ${action} request`, 'error');
    } finally {
      loading = false;
    }
  }

  function formatDate(dateString: string): string {
    try {
      // Parse the ISO date string
      const date = new Date(dateString);
      // Check if date is valid
      if (isNaN(date.getTime())) {
        return 'Date not available';
      }
      // Format the date
      return date.toLocaleString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch (error) {
      console.error('Error formatting date:', error);
      return 'Date not available';
    }
  }

  async function handleNotificationClick(notification: any) {
    try {
      if (!notification.isRead) {
        await markNotificationAsRead(notification.id);
      }

      // Handle navigation or other actions based on notification type
      if (notification.link) {
        goto(notification.link);
      }

      // Close dropdown after handling notification
      isOpen = false;
    } catch (error) {
      console.error('Error handling notification click:', error);
      showToast('Failed to mark notification as read', 'error');
    }
  }
</script>

<div class="relative">
  <Button
    class="relative"
    color="alternative"
    on:click={() => isOpen = !isOpen}
  >
    <Bell class="w-5 h-5" />
    {#if $unreadNotificationsCount > 0}
            <span
              class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center"
              transition:fade={{ duration: 200 }}
            >
                {$unreadNotificationsCount}
            </span>
    {/if}
  </Button>

  <Dropdown
    bind:open={isOpen}
    class="w-80 max-h-96 overflow-y-auto custom-scrollbar"
  >
    <div class="py-2">
      <h3 class="px-4 py-2 text-sm font-semibold text-gray-900 dark:text-white">
        Notifications
      </h3>
      {#if $notifications.length === 0}
        <div class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">
          No notifications
        </div>
      {:else}
        {#each $notifications as notification (notification.id)}
          <!-- Debug notification -->
          <!-- {JSON.stringify(notification)} -->
          <div
            class="px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200 cursor-pointer
                               {notification.isRead ? 'opacity-75' : 'bg-blue-50 dark:bg-blue-900/20'}"
            on:click={(event) => handleNotificationClick(notification)}
          >
            <div class="text-sm font-medium text-gray-900 dark:text-white">
              {notification.content}
            </div>

            {#if notification.type === 'group_invitation' && !notification.isProcessed}
              <div class="flex gap-2 mt-2">
                <Button
                  size="xs"
                  color="green"
                  disabled={loading}
                  on:click={(e) => {
                                        e.stopPropagation(); // Prevent notification click
                                        handleInviteResponse(notification, 'accept');
                                    }}
                >
                  {loading ? 'Processing...' : 'Accept'}
                </Button>
                <Button
                  size="xs"
                  color="red"
                  disabled={loading}
                  on:click={(e) => {
                                        e.stopPropagation(); // Prevent notification click
                                        handleInviteResponse(notification, 'reject');
                                    }}
                >
                  {loading ? 'Processing...' : 'Reject'}
                </Button>
              </div>
            {/if}

            {#if notification.type === 'join_request' && isGroupAdmin(notification) && !notification.isProcessed}
              <div class="flex gap-2 mt-2">
                <Button
                  size="xs"
                  color="green"
                  disabled={loading}
                  on:click={(e) => {
                                        e.stopPropagation();
                                        handleJoinRequest(notification, 'accept');
                                    }}
                >
                  {loading ? 'Processing...' : 'Accept Request'}
                </Button>
                <Button
                  size="xs"
                  color="red"
                  disabled={loading}
                  on:click={(e) => {
                                        e.stopPropagation();
                                        handleJoinRequest(notification, 'reject');
                                    }}
                >
                  {loading ? 'Processing...' : 'Reject Request'}
                </Button>
              </div>
            {/if}

            {#if notification.type === 'group_event'}
              <div class="flex items-center space-x-2">
                <p class="text-sm text-gray-700 dark:text-gray-300">
                  {notification.content}
                </p>
              </div>
            {/if}
          </div>
        {/each}
        <div class="px-4 py-2 border-t border-gray-200 dark:border-gray-700">
          <Button
            size="xs"
            color="alternative"
            class="w-full"
            on:click={() => markAllNotificationsAsRead()}
          >
            Mark all as read
          </Button>
        </div>
      {/if}
    </div>
  </Dropdown>
</div>

{#if toastVisible}
  <div
    class="fixed bottom-4 right-4 p-4 rounded-lg shadow-lg z-50"
    class:bg-green-100={toastType === 'success'}
    class:bg-red-100={toastType === 'error'}
    class:bg-blue-100={toastType === 'info'}
    class:text-green-800={toastType === 'success'}
    class:text-red-800={toastType === 'error'}
    class:text-blue-800={toastType === 'info'}
    transition:fade={{ duration: 200 }}
  >
    {toastMessage}
  </div>
{/if}

<style>
    .notification-item {
        @apply transition-all duration-300;
    }

    .notification-item:hover {
        @apply transform -translate-y-0.5;
    }
</style>