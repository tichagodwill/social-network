<script lang="ts">
    import { notifications } from '$lib/stores/notifications';
    import { Button, Dropdown, DropdownItem } from 'flowbite-svelte';
    import { BellSolid } from 'flowbite-svelte-icons';
    import { getFormattedDate } from '$lib/dateFormater';
    import { goto } from '$app/navigation'

    async function handleNotificationClick(notification: any) {
        await notifications.markAsRead(notification.id);
        
        // Handle different notification types
        switch (notification.type) {
            case 'follow_request':
                window.location.href = `/profile/${notification.fromUserId}`;
                break;
            case 'group_invite':
                window.location.href = `/groups/${notification.groupId}`;
                break;
            case 'group_event':
                window.location.href = `/groups/${notification.groupId}/events/${notification.eventId}`;
                break;
            case 'chat':
                goto(`/chat/${notification.fromUserId}`)
                break;
            default:
                break;
        }
    }
</script>

<div class="relative">
    <Button color="alternative" class="!p-2">
        <div class="relative">
            <BellSolid class="w-5 h-5" />
            {#if $notifications.unreadCount > 0}
                <div class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs">
                    {$notifications.unreadCount}
                </div>
            {/if}
        </div>
    </Button>
    
    <Dropdown class="w-80">
        <div class="py-2">
            <h6 class="px-4 py-2 font-semibold text-sm">Notifications</h6>
            {#if $notifications.notifications.length === 0}
                <div class="px-4 py-2 text-sm text-gray-500">
                    No notifications
                </div>
            {:else}
                {#each $notifications.notifications as notification}
                    <DropdownItem
                        class="flex items-start gap-3 cursor-pointer"
                        on:click={() => handleNotificationClick(notification)}
                    >
                        <div class:font-bold={!notification.isRead} class="flex-1">
                            <p>{notification.content}</p>
                            <p class="text-xs text-gray-500">
                                {getFormattedDate(new Date(notification.createdAt)).diff}
                            </p>
                        </div>
                    </DropdownItem>
                {/each}
            {/if}
        </div>
    </Dropdown>
</div> 