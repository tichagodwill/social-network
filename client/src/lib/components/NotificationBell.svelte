<script lang="ts">
    import { Bell } from 'lucide-svelte';
    import { Button, Dropdown, DropdownItem } from 'flowbite-svelte';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { auth } from '$lib/stores/auth'; // Import auth store

    let notifications: any[] = [];
    let unreadCount = 0;
    let loading = false;
    let isOpen = false;

    // Add a toast notification system
    let toast = { message: '', type: '', visible: false };

    // Function to check if current user is the invitee
    function isInvitee(notification: any): boolean {
        return notification.type === 'group_invitation' && 
               notification.user_id === $auth?.user?.id;
    }

    // Function to check if current user is admin/creator of the group
    function isGroupAdmin(notification: any): boolean {
        return notification.type === 'join_request' && 
               (notification.user_role === 'creator' || notification.user_role === 'admin');
    }

    function showToast(message: string, type: 'success' | 'error' | 'info') {
        toast = { message, type, visible: true };
        setTimeout(() => {
            toast = { ...toast, visible: false };
        }, 3000);
    }

    async function fetchNotifications() {
        try {
            const response = await fetch('http://localhost:8080/notifications', {
                credentials: 'include'
            });
            if (response.ok) {
                const data = await response.json();
                notifications = data.notifications || [];
                unreadCount = notifications.filter(n => !n.read).length;
            }
        } catch (error) {
            console.error('Error fetching notifications:', error);
        }
    }

    async function handleInvitation(notification: any, action: 'accept' | 'reject') {
        try {
            loading = true;
            const response = await fetch(
                `http://localhost:8080/groups/${notification.group_id}/invitations/${notification.invitation_id}/${action}`, 
                {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                }
            );

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || `Failed to ${action} invitation`);
            }

            // Remove this notification
            notifications = notifications.filter(n => n.id !== notification.id);
            unreadCount = notifications.filter(n => !n.read).length;

            // Show success message
            const message = action === 'accept' ? 'Invitation accepted! Redirecting...' : 'Invitation rejected';
            showToast(message, action === 'accept' ? 'success' : 'info');

            if (action === 'accept') {
                // Redirect to the group page after a short delay
                setTimeout(() => {
                    goto(`/groups/${notification.group_id}`);
                }, 1500);
            }
        } catch (error) {
            console.error('Error handling invitation:', error);
            showToast(error.message || `Failed to ${action} invitation`, 'error');
        } finally {
            loading = false;
        }
    }

    // Function to handle join requests
    async function handleJoinRequest(notification: any, action: 'accept' | 'reject') {
        try {
            loading = true;
            const response = await fetch(
                `http://localhost:8080/groups/${notification.group_id}/requests/${notification.request_id}/${action}`, 
                {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                }
            );

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || `Failed to ${action} request`);
            }

            // Remove this notification
            notifications = notifications.filter(n => n.id !== notification.id);
            unreadCount = notifications.filter(n => !n.read).length;

            showToast(`Join request ${action}ed successfully`, 'success');
        } catch (error) {
            console.error('Error handling join request:', error);
            showToast(error.message || `Failed to ${action} request`, 'error');
        } finally {
            loading = false;
        }
    }

    async function markAsRead(notificationId: number) {
        try {
            const response = await fetch(`http://localhost:8080/notifications/${notificationId}/read`, {
                method: 'GET',
                credentials: 'include'
            });
            if (response.ok) {
                notifications = notifications.map(n => 
                    n.id === notificationId ? { ...n, read: true } : n
                );
                unreadCount = notifications.filter(n => !n.read).length;
            }
        } catch (error) {
            console.error('Error marking notification as read:', error);
        }
    }

    // Function to check if notification should be shown
    function shouldShowNotification(notification: any): boolean {
        if (notification.type === 'group_invitation') {
            // Show invitation notifications only to the invitee
            return notification.user_id === $auth?.user?.id;
        } else if (notification.type === 'join_request') {
            // Show join request notifications only to admins/creators
            return notification.user_role === 'creator' || notification.user_role === 'admin';
        }
        return true; // Show other types of notifications to everyone
    }

    onMount(() => {
        fetchNotifications();
        // Set up polling for new notifications
        const interval = setInterval(fetchNotifications, 30000); // Poll every 30 seconds
        return () => clearInterval(interval);
    });
</script>

<div class="relative">
    <Button class="!p-2" color="light" on:click={() => isOpen = !isOpen}>
        <Bell class="w-5 h-5" />
        {#if unreadCount > 0}
            <span class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
                {unreadCount}
            </span>
        {/if}
    </Button>
    <Dropdown 
        class="w-80 max-h-[80vh] overflow-y-auto" 
        open={isOpen} 
        trigger="click"
        placement="bottom-end"
    >
        <div class="py-2">
            <h6 class="px-4 py-2 font-medium text-gray-900 dark:text-white">
                Notifications ({notifications.length})
            </h6>
            {#if notifications.length === 0}
                <div class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">
                    No notifications
                </div>
            {:else}
                {#each notifications as notification}
                    {#if notification.type === 'group_invitation'}
                        <div class="notification-item p-3 border-l-4 border-blue-500 bg-blue-50 dark:bg-blue-900/20 rounded">
                            <p class="text-sm font-medium text-blue-800 dark:text-blue-200 mb-2">
                                {notification.content}
                            </p>
                            {#if notification.userId === $auth?.user?.id}
                                <div class="flex gap-2">
                                    <Button 
                                        size="xs" 
                                        color="green"
                                        disabled={loading}
                                        on:click={() => handleInvitation(notification, 'accept')}
                                    >
                                        {loading ? 'Processing...' : 'Accept'}
                                    </Button>
                                    <Button 
                                        size="xs" 
                                        color="red"
                                        disabled={loading}
                                        on:click={() => handleInvitation(notification, 'reject')}
                                    >
                                        {loading ? 'Processing...' : 'Reject'}
                                    </Button>
                                </div>
                            {/if}
                        </div>
                    {:else if notification.type === 'join_request' && (notification.userRole === 'creator' || notification.userRole === 'admin')}
                        <div class="notification-item p-3 border-l-4 border-yellow-500 bg-yellow-50 dark:bg-yellow-900/20 rounded">
                            <p class="text-sm font-medium text-yellow-800 dark:text-yellow-200 mb-2">
                                {notification.content}
                            </p>
                            <div class="flex gap-2">
                                <Button 
                                    size="xs" 
                                    color="green"
                                    disabled={loading}
                                    on:click={() => handleJoinRequest(notification, 'accept')}
                                >
                                    {loading ? 'Processing...' : 'Accept Request'}
                                </Button>
                                <Button 
                                    size="xs" 
                                    color="red"
                                    disabled={loading}
                                    on:click={() => handleJoinRequest(notification, 'reject')}
                                >
                                    {loading ? 'Processing...' : 'Reject Request'}
                                </Button>
                            </div>
                        </div>
                    {:else}
                        <div class="text-sm {notification.read ? 'text-gray-600' : 'text-gray-900 font-medium'}">
                            {notification.content}
                        </div>
                    {/if}
                    <div class="text-xs text-gray-500 mt-1">
                        {new Date(notification.created_at).toLocaleString()}
                    </div>
                {/each}
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
    >
        {toast.message}
    </div>
{/if}

<style>
    .notification-enter {
        animation: slideIn 0.3s ease-out;
    }

    @keyframes slideIn {
        from {
            transform: translateY(-20px);
            opacity: 0;
        }
        to {
            transform: translateY(0);
            opacity: 1;
        }
    }

    .notification-item {
        @apply transition-all duration-300;
    }

    .notification-item:hover {
        @apply transform -translate-y-0.5;
    }
</style> 