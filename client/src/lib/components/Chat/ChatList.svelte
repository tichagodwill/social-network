<script lang="ts">
    import {onMount, onDestroy} from 'svelte';
    import {Avatar, Badge, Input, Spinner} from 'flowbite-svelte';

    import {activeChats} from '$lib/stores/websocket';
    import {get} from 'svelte/store';
    import {auth} from '$lib/stores/auth';
    import defaultProfileImg from '$lib/assets/default-profile.jpg';

    // Props
    export let onSelectChat: (chatId: number, isGroup: boolean) => void;

    const MagnifyingGlass = `
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
      <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
    </svg>
  `;

    // Component state
    let searchQuery = '';
    let loading = true;
    let selectedChatId: number | null = null;

    $: filteredChats = $activeChats.filter(chat =>
        chat.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    // Get current user ID from auth store
    function getCurrentUserId(): number | null {
        const authState = get(auth);
        return authState.user?.id || null;
    }

    // Load contacts/chats list
    async function loadChats() {
        try {
            const currentUserId = getCurrentUserId();
            if (!currentUserId) {
                console.error('No current user ID available');
                return;
            }

            const response = await fetch(
                `http://localhost:8080/chats`,
                {credentials: 'include'}
            );

            if (response.ok) {
                const chats = await response.json();
                console.log("[DEBUG] Loaded chats:", chats);

                if (!Array.isArray(chats)) {
                    console.error("Invalid response format, expected array:", chats);
                    loading = false;
                    return;
                }

                // Process chats - they now have real IDs from the database
                activeChats.set(chats.map((chat: any) => {
                    // Ensure all required properties have default values
                    return {
                        id: chat.id || 0, // Use the real chat ID from database
                        name: chat.name || `${chat.first_name || ''} ${chat.last_name || ''}`.trim() || 'Unknown',
                        avatar: chat.avatar || null,
                        unreadCount: Number(chat.unread_count || 0),
                        isGroup: chat.type === 'group',
                        lastMessage: chat.last_message || '',
                        lastMessageTime: chat.last_message_time || null,
                        recipientId: chat.participant_id || null,
                        potential: Boolean(chat.potential || false) // New flag for potential chats
                    };
                }));
            } else {
                console.error('Error loading chats:', response.status, response.statusText);
            }
        } catch (error) {
            console.error('Error loading chats:', error);
        } finally {
            loading = false;
        }
    }


    // Format timestamp
    function formatLastMessageTime(timestamp?: string): string {
        if (!timestamp) return '';

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
        if (diffDays < 7) return date.toLocaleDateString([], {weekday: 'short'});

        return date.toLocaleDateString([], {month: 'short', day: 'numeric'});
    }

    // Truncate message for preview
    function truncateMessage(message?: string): string {
        if (!message) return '';
        return message.length > 30 ? message.substring(0, 27) + '...' : message;
    }

    function handleChatSelect(chatId: number, isGroup: boolean) {
        if (chatId < 0) {
            // Handle potential chats with negative IDs
            onSelectChat(chatId, isGroup);
        } else {
            selectedChatId = chatId;
            onSelectChat(chatId, isGroup);
        }
    }

    // Load chats on mount
    onMount(async () => {
        await loadChats();
        
        // Add event listener for refreshing the chat list
        document.addEventListener('refresh-chat-list', async () => {
            await loadChats();
        });
    });
    
    onDestroy(() => {
        // Clean up event listener
        document.removeEventListener('refresh-chat-list', async () => {
            await loadChats();
        });
    });
</script>

<style>
    .custom-scrollbar {
        scrollbar-width: thin;
        scrollbar-color: #ccc #f4f4f4;
    }

    .custom-scrollbar::-webkit-scrollbar {
        width: 8px;
    }

    .custom-scrollbar::-webkit-scrollbar-track {
        background: #f4f4f4;
    }

    .custom-scrollbar::-webkit-scrollbar-thumb {
        background-color: #ccc;
        border-radius: 10px;
    }
</style>

<div class="h-full flex flex-col overflow-hidden bg-white dark:bg-gray-800 border-r">
    <!-- Search header -->
    <div class="p-4 border-b">
        <Input
                type="text"
                placeholder="Search conversations..."
                bind:value={searchQuery}
                size="md"
        >
            {@html MagnifyingGlass}
        </Input>
    </div>

    <!-- Chat list -->
    <div class="flex-1 overflow-y-auto custom-scrollbar">
        {#if loading}
            <div class="flex justify-center items-center p-8">
                <Spinner size="6"/>
            </div>
        {:else if filteredChats.length === 0}
            <div class="flex flex-col items-center justify-center p-8 text-center text-gray-500">
                <p>No conversations found</p>
                {#if searchQuery}
                    <p class="text-sm mt-1">Try a different search term</p>
                {:else}
                    <p class="text-sm mt-1">Connect with users to start chatting</p>
                {/if}
            </div>
        {:else}
            <ul class="divide-y">
                {#each filteredChats as chat (chat.id)}
                    {@const isActive = selectedChatId === chat.id}
                    <li>
                        <button
                                class="w-full p-3 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150 {isActive ? 'bg-gray-100 dark:bg-gray-700' : ''}"
                                on:click={() => handleChatSelect(chat.id, chat.isGroup)}
                        >
                            <div class="flex items-start gap-3">
                                <!-- Avatar with online indicator -->
                                <div class="relative">
                                    <Avatar
                                            src={chat.avatar || defaultProfileImg}
                                            alt={chat.name}
                                            class="w-12 h-12"
                                            rounded={chat.isGroup ? false : true}
                                    />
                                    {#if chat.isGroup}
                                        <Badge color="purple"
                                               class="absolute -top-1 -right-1 text-xs w-6 h-6 flex items-center justify-center rounded-full p-0">
                                            <span>G</span>
                                        </Badge>
                                    {/if}
                                </div>

                                <!-- Chat info -->
                                <div class="flex-1 min-w-0">
                                    <div class="flex justify-between items-start">
                                        <h3 class="font-medium truncate max-w-[120px]">{chat.name}</h3>
                                        <span class="text-xs text-gray-500">
                      {formatLastMessageTime(chat.lastMessageTime)}
                    </span>
                                    </div>

                                    <p class="text-sm text-gray-600 dark:text-gray-300 truncate">
                                        {#if chat.potential}
                                            Start a new conversation
                                        {:else}
                                            {truncateMessage(chat.lastMessage) || 'Start a conversation...'}
                                        {/if}
                                    </p>

                                    <!-- Unread counter -->
                                    {#if chat.unreadCount > 0}
                                        <div class="mt-1">
                                            <Badge color="red" class="px-2 py-1">
                                                {chat.unreadCount > 99 ? '99+' : chat.unreadCount}
                                            </Badge>
                                        </div>
                                    {/if}
                                    
                                    <!-- Potential chat indicator -->
                                    {#if chat.potential}
                                        <div class="mt-1">
                                            <Badge color="blue" class="px-2 py-1">New</Badge>
                                        </div>
                                    {/if}
                                </div>
                            </div>
                        </button>
                    </li>
                {/each}
            </ul>
        {/if}
    </div>
</div>