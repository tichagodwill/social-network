<script lang="ts">
  import { Bell } from 'lucide-svelte';
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
  let toast = { message: '', type: '', visible: false };
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

  function showToast(message: string, type: 'success' | 'error' | 'info') {
    toast = { message, type, visible: true };
    setTimeout(() => {
      toast = { ...toast, visible: false };
    }, 3000);
  }

  async function handleInviteResponse(notification: any, action: 'accept' | 'reject') {
    try {
      loading = true;

      // First handle the invitation response
      const response = await fetch(`http://localhost:8080/groups/${notification.groupId}/invitations/${notification.invitationId}/${action}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include'
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Failed to ${action} invitation`);
      }

      // After successful invitation handling, mark the notification as read
      await markNotificationAsRead(notification.id);

      if (action === 'accept') {
        // Show success message for accepting
        showToast('Successfully joined the group!', 'success');

        // Close the notification dropdown
        isOpen = false;

        // Navigate to the group page immediately after accepting
        if (notification.groupId) {
          goto(`/groups/${notification.groupId}`);
        }
      } else {
        // Show message for rejecting
        showToast('Invitation rejected', 'info');
      }
    } catch (error: any) {
      console.error('Error handling invitation response:', error);
      if (error.message.includes('already processed')) {
        showToast('This invitation has already been processed', 'info');
      } else {
        showToast(error.message || `Failed to ${action} invitation`, 'error');
      }
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

  async function handleNotificationClick(notification: any, event: MouseEvent) {
    try {
      event.stopPropagation();
      event.preventDefault();

      if (!notification.isRead) {
        // Use the centralized markNotificationAsRead function
        await markNotificationAsRead(notification.id);

        // Show success toast
        showToast('Notification marked as read', 'success');

        // Handle navigation for group invitations
        if (notification.type === 'group_invitation' && !notification.isProcessed && notification.groupId) {
          isOpen = false;
          goto(`/groups/${notification.groupId}`);
        }
      }
    } catch (error) {
      console.error('Error handling notification click:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to mark notification as read';
      showToast(errorMessage, 'error');
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
          <div
            class="px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200 cursor-pointer
                               {notification.isRead ? 'opacity-75' : 'bg-blue-50 dark:bg-blue-900/20'}"
            on:click={(event) => handleNotificationClick(notification, event)}
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

            <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {formatDate(notification.createdAt)}
            </div>
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

{#if toast.visible}
  <div
    class="fixed bottom-4 right-4 p-4 rounded-lg shadow-lg z-50 transition-all duration-300"
    class:bg-green-100={toast.type === 'success'}
    class:bg-red-100={toast.type === 'error'}
    class:bg-blue-100={toast.type === 'info'}
    class:text-green-800={toast.type === 'success'}
    class:text-red-800={toast.type === 'error'}
    class:text-blue-800={toast.type === 'info'}
    transition:fade={{ duration: 300 }}
  >
    {toast.message}
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