<script lang="ts">
    import { Bell } from 'lucide-svelte';
    import { Button, Dropdown, DropdownItem } from 'flowbite-svelte';
    import { onMount, onDestroy } from 'svelte';
    import { goto } from '$app/navigation';
    import { auth } from '$lib/stores/auth'; // Import auth store
    import { handleInvitation } from '$lib/api/groupApi';
    import { fade } from 'svelte/transition';

    let notifications: any[] = [];
    let unreadCount = 0;
    let loading = false;
    let isOpen = false;
    let socket: WebSocket | null = null;

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

    // Function to check if current user is the recipient of the notification
    function isRecipient(notification: any): boolean {
        return notification.user_id === $auth?.user?.id;
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
                console.log('Received notifications:', data);

                // Data is already sorted by created_at DESC from the server
                notifications = data.map((notification: any) => ({
                    ...notification,
                    created_at: notification.createdAt || notification.created_at,
                    is_read: notification.isRead,
                    is_processed: notification.isProcessed
                }));

                console.log('Processed notifications:', notifications);
                
                // Update unread count based on unread notifications
                unreadCount = notifications.filter(n => !n.is_read).length;
                console.log('Unread count:', unreadCount);
            } else {
                console.error('Failed to fetch notifications:', response.status);
                const errorData = await response.json();
                console.error('Error data:', errorData);
            }
        } catch (error) {
            console.error('Error fetching notifications:', error);
            showToast('Failed to load notifications', 'error');
        }
    }

    async function markAsRead(notificationId: number) {
        try {
            const response = await fetch(`http://localhost:8080/notifications/${notificationId}/read`, {
                method: 'GET',
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                
                // Update the notification in the list
                notifications = notifications.map(notification => 
                    notification.id === notificationId 
                        ? { ...notification, is_read: true }
                        : notification
                );
                
                // Update unread count from server response
                unreadCount = data.unreadCount;
                
                // Force a UI update
                notifications = [...notifications];
            } else {
                throw new Error('Failed to mark notification as read');
            }
        } catch (error) {
            console.error('Error marking notification as read:', error);
            showToast('Failed to mark notification as read', 'error');
        }
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
            const readResponse = await fetch(`http://localhost:8080/notifications/${notification.id}/read`, {
                method: 'GET',
                credentials: 'include'
            });

            const data = await readResponse.json();
            
            // Remove this notification from the list regardless of read status
            notifications = notifications.filter(n => n.id !== notification.id);
            
            // Update unread count
            unreadCount = Math.max(0, unreadCount - 1);
            
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

            // Only show read error if it's not a server error (e.g., notification already processed)
            if (!readResponse.ok && readResponse.status !== 404) {
                console.error('Error marking notification as read:', await readResponse.text());
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
            const response = await fetch(`http://localhost:8080/groups/${notification.group_id}/join-requests/${action}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ requestId: notification.request_id })
            });

            if (!response.ok) {
                throw new Error(`Failed to ${action} request`);
            }

            // Remove notification from list first
            notifications = notifications.filter(n => n.id !== notification.id);
            
            // Update unread count if notification was unread
            if (!notification.is_read) {
                unreadCount = Math.max(0, unreadCount - 1);
            }
            
            // Show single success message
            showToast(`Join request ${action}ed successfully`, 'success');

            // Silently mark as read without showing additional messages
            await fetch(`http://localhost:8080/notifications/${notification.id}/read`, {
                method: 'GET',
                credentials: 'include'
            });

        } catch (error: any) {
            showToast(error.message || `Failed to ${action} request`, 'error');
        } finally {
            loading = false;
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

    function initializeWebSocket() {
        socket = new WebSocket('ws://localhost:8080/ws');

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'notification') {
                    // Add new notification to the beginning of the list
                    notifications = [data.data, ...notifications];
                    unreadCount += 1;
                    
                    // Show a toast for new notifications
                    showToast('New notification received', 'info');
                }
            } catch (error) {
                console.error('Error processing WebSocket message:', error);
            }
        };

        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        socket.onclose = () => {
            // Attempt to reconnect after a delay
            setTimeout(initializeWebSocket, 5000);
        };
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

            if (!notification.is_read) {
                console.log('Marking notification as read:', notification.id);
                const response = await fetch(`http://localhost:8080/notifications/${notification.id}/read`, {
                    method: 'GET',
                    credentials: 'include'
                });

                const data = await response.json();
                console.log('Mark as read response:', data);

                if (!response.ok) {
                    throw new Error(data.error || 'Failed to mark notification as read');
                }

                // Update notification in the list
                notifications = notifications.map(n => 
                    n.id === notification.id 
                        ? { ...n, is_read: true }
                        : n
                );

                // Update unread count
                unreadCount = Math.max(0, unreadCount - 1);

                // Show success toast
                showToast('Notification marked as read', 'success');

                // Handle navigation for group invitations
                if (notification.type === 'group_invitation' && !notification.is_processed && notification.groupId) {
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

    onMount(() => {
        fetchNotifications();
        initializeWebSocket();
        
        const interval = setInterval(fetchNotifications, 5000); // Poll every 5 seconds
        
        return () => {
            clearInterval(interval);
            if (socket) {
                socket.close();
            }
        };
    });

    onDestroy(() => {
        if (socket) {
            socket.close();
        }
    });
</script>

<div class="relative">
    <Button
        class="relative"
        color="alternative"
        on:click={() => isOpen = !isOpen}
    >
        <Bell class="w-5 h-5" />
        {#if unreadCount > 0}
            <span 
                class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center"
                transition:fade={{ duration: 200 }}
            >
                {unreadCount}
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
            {#if notifications.length === 0}
                <div class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">
                    No notifications
                </div>
            {:else}
                {#each notifications as notification (notification.id)}
                    <div 
                        class="px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200 cursor-pointer
                               {notification.is_read ? 'opacity-75' : 'bg-blue-50 dark:bg-blue-900/20'}"
                        on:click={(event) => handleNotificationClick(notification, event)}
                    >
                        <div class="text-sm font-medium text-gray-900 dark:text-white">
                            {notification.content}
                        </div>
                        
                        {#if notification.type === 'group_invitation' && !notification.is_processed}
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

                        {#if notification.type === 'join_request' && isGroupAdmin(notification) && !notification.is_processed}
                            <div class="flex gap-2 mt-2">
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
                        {/if}

                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                            {formatDate(notification.created_at)}
                        </div>
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