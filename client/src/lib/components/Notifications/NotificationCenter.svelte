<!-- src/lib/components/Notifications/NotificationCenter.svelte -->
<script lang="ts">
    import { onMount } from 'svelte';
    import { slide } from 'svelte/transition';
    import { goto } from '$app/navigation';
    import { Avatar, Button, Card, Badge, Popover, ToggleSwitch } from 'flowbite-svelte';
    import { BellOutline, CheckSolid, ChatBubbleOvalOutline } from 'flowbite-svelte-icons';

    import {
        notifications,
        unreadNotificationsCount,
        markAllNotificationsAsRead,
        markNotificationAsRead,
        type NotificationMessage
    } from '$lib/stores/websocket';

    // Component state
    let isOpen = false;
    let popover: HTMLElement;
    let showOnlyUnread = true;

    // Filtered notifications
    $: filteredNotifications = showOnlyUnread 
        ? $notifications.filter(n => !n.isRead)
        : $notifications;

    // Format notification time
    function formatTime(timestamp: string): string {
        const date = new Date(timestamp);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMins / 60);
        const diffDays = Math.floor(diffHours / 24);

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        if (diffDays === 1) return 'Yesterday';
        if (diffDays < 7) return date.toLocaleDateString([], { weekday: 'short' });

        return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }

    // Handle notification click
    function handleNotificationClick(notification: NotificationMessage) {
        // Mark as read
        if (!notification.isRead && notification.id) {
            markNotificationAsRead(notification.id);
        }

        // Navigate if there's a link
        if (notification.link) {
            goto(notification.link);
        }

        // Close popover
        isOpen = false;
    }

    // Mark all as read
    function handleMarkAllAsRead() {
        markAllNotificationsAsRead();
    }

    // Close popover when clicking outside
    function handleClickOutside(event: MouseEvent) {
        if (isOpen && popover && !popover.contains(event.target as Node)) {
            isOpen = false;
        }
    }

    onMount(() => {
        document.addEventListener('click', handleClickOutside);

        return () => {
            document.removeEventListener('click', handleClickOutside);
        };
    });
</script>

<div class="relative" bind:this={popover}>
    <Button pill color="light" class="relative" on:click={() => isOpen = !isOpen}>
        <BellOutline class="w-5 h-5" />
        {#if $unreadNotificationsCount > 0}
            <Badge
                    color="red"
                    class="absolute -top-1 -right-1 text-xs min-w-5 h-5 flex items-center justify-center rounded-full"
            >
                {$unreadNotificationsCount > 99 ? '99+' : $unreadNotificationsCount}
            </Badge>
        {/if}
    </Button>

    {#if isOpen}
        <div
                class="absolute right-0 mt-2 w-80 z-50"
                transition:slide={{ duration: 200 }}
        >
            <Card padding="sm" class="!p-0 max-h-[80vh] flex flex-col">
                <!-- Header -->
                <div class="flex items-center justify-between p-4 border-b">
                    <h3 class="font-semibold">Notifications</h3>
                    <div class="flex items-center space-x-2">
                        {#if $unreadNotificationsCount > 0}
                            <Button size="xs" color="light" class="text-xs" on:click={handleMarkAllAsRead}>
                                <CheckSolid class="w-3 h-3 mr-1" />
                                Mark all as read
                            </Button>
                        {/if}
                    </div>
                </div>
                
                <!-- Filter toggle -->
                <div class="px-4 py-2 border-b flex items-center justify-between">
                    <span class="text-sm text-gray-600">Show only unread</span>
                    <ToggleSwitch bind:checked={showOnlyUnread} size="sm" />
                </div>

                <!-- Notification list -->
                <div class="overflow-y-auto custom-scrollbar flex-1">
                    {#if filteredNotifications.length === 0}
                        <div class="p-4 text-center text-gray-500">
                            <p>{showOnlyUnread ? 'No unread notifications' : 'No notifications yet'}</p>
                        </div>
                    {:else}
                        <ul class="divide-y">
                            {#each filteredNotifications as notification (notification.id)}
                                <li>
                                    <button
                                            class="w-full p-3 text-left hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150 {notification.isRead ? '' : 'bg-primary-50 dark:bg-primary-900'}"
                                            on:click={() => handleNotificationClick(notification)}
                                    >
                                        <div class="flex gap-3">
                                            <!-- Notification icon/avatar -->
                                            <div class="flex-shrink-0">
                                                {#if notification.type === 'chat'}
                                                    <div class="w-10 h-10 bg-blue-100 text-blue-600 flex items-center justify-center rounded-full">
                                                        <ChatBubbleOvalOutline class="h-5 w-5" />
                                                    </div>
                                                {:else if notification.link && notification.link.includes('/profile/')}
                                                    <Avatar rounded size="md" />
                                                {:else if notification.link && notification.link.includes('/groups/')}
                                                    <div class="w-10 h-10 bg-purple-100 text-purple-600 flex items-center justify-center rounded-md">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                                        </svg>
                                                    </div>
                                                {:else}
                                                    <div class="w-10 h-10 bg-primary-100 text-primary-600 flex items-center justify-center rounded-full">
                                                        <BellOutline class="h-5 w-5" />
                                                    </div>
                                                {/if}
                                            </div>

                                            <!-- Notification content -->
                                            <div class="flex-1 min-w-0">
                                                <p class="text-sm font-medium">
                                                    {notification.content}
                                                </p>
                                                <p class="text-xs text-gray-500 mt-1">
                                                    {formatTime(notification.createdAt)}
                                                </p>
                                            </div>

                                            {#if !notification.isRead}
                                                <div class="flex-shrink-0 self-center">
                                                    <div class="w-2 h-2 bg-primary-500 rounded-full"></div>
                                                </div>
                                            {/if}
                                        </div>
                                    </button>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                </div>
            </Card>
        </div>
    {/if}
</div>