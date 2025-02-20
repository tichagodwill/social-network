<!-- src/routes/chat/+page.svelte -->
<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { Button, Card, Spinner } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { get } from 'svelte/store';

    // SVG icons
    const UsersGroupIcon = `
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 0 0 3.741-.479 3 3 0 0 0-4.682-2.72m.94 3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0 1 12 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0 1 6 18.719m12 0a5.971 5.971 0 0 0-.941-3.197m0 0A5.995 5.995 0 0 0 12 12.75a5.995 5.995 0 0 0-5.058 2.772m0 0a3 3 0 0 0-4.681 2.72 8.986 8.986 0 0 0 3.74.477m.94-3.197a5.971 5.971 0 0 0-.94 3.197M15 6.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm6 3a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Zm-13.5 0a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Z" />
        </svg>
    `;
    const UsersIcon = `
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />
        </svg>
    `;

    import ChatWindow from '$lib/components/Chat/ChatWindow.svelte';
    import ChatList from '$lib/components/Chat/ChatList.svelte';
    import EmptyState from '$lib/components/UI/EmptyState.svelte';
    import { initializeWebSocket, requestNotificationPermission } from '$lib/stores/websocket';

    // Component state
    let selectedChat: {
        id: number;
        isGroup: boolean;
        recipientId?: number;
        name: string;
        avatar?: string;
    } | null = null;

    let isMobileView = false;
    let showChatList = true;
    let loading = true;

    // Handle chat selection
    async function handleSelectChat(chatId: number, isGroup: boolean) {
        loading = true;

        try {
            if (isGroup) {
                // Fetch group details
                const response = await fetch(`http://localhost:8080/groups/${chatId}`, {
                    credentials: 'include'
                });
                if (response.ok) {
                    const groupData = await response.json();
                    selectedChat = {
                        id: chatId,
                        isGroup: true,
                        name: groupData.name,
                        avatar: groupData.avatar
                    };
                }
            } else {
                // For direct chats, we need to extract the other user's ID
                const currentUserId = getCurrentUserId();
                const id1 = Math.floor(chatId / 1000000);
                const id2 = chatId % 1000000;
                const otherUserId = id1 === currentUserId ? id2 : id1;

                // Fetch user details
                const response = await fetch(`http://localhost:8080/user/${otherUserId}`, {
                    credentials: 'include'
                });
                if (response.ok) {
                    const userData = await response.json();
                    selectedChat = {
                        id: chatId,
                        isGroup: false,
                        recipientId: otherUserId,
                        name: `${userData.firstName} ${userData.lastName}`,
                        avatar: userData.avatar
                    };
                }
            }

            // Update URL
            const newParams = new URLSearchParams();
            newParams.set('id', chatId.toString());
            newParams.set('type', isGroup ? 'group' : 'direct');
            goto(`/chat?${newParams.toString()}`, { replaceState: true });

            // On mobile, show the chat and hide the list
            if (isMobileView) {
                showChatList = false;
            }
        } catch (error) {
            console.error('Error selecting chat:', error);
        } finally {
            loading = false;
        }
    }

    // Get current user ID from auth store
    function getCurrentUserId(): number {
        const authState = get(auth);
        return authState.user?.id || 0;
    }

    // Toggle between chat list and chat on mobile
    function toggleView() {
        showChatList = !showChatList;
    }

    // Handle window resize to check for mobile view
    function handleResize() {
        isMobileView = window.innerWidth < 768;
        if (!isMobileView) {
            showChatList = true;
        }
    }

    // Check URL for chat parameters on page load
    async function checkUrlParams() {
        const params = $page.url.searchParams;
        const chatId = params.get('id');
        const chatType = params.get('type');

        if (chatId && chatType) {
            await handleSelectChat(
                parseInt(chatId, 10),
                chatType === 'group'
            );
        }

        loading = false;
    }

    onMount(() => {
        // Initialize websocket connection
        initializeWebSocket();

        // Request notification permission
        requestNotificationPermission();

        // Check for mobile view
        handleResize();
        window.addEventListener('resize', handleResize);

        // Check URL parameters
        checkUrlParams();
    });

    onDestroy(() => {
        window.removeEventListener('resize', handleResize);
    });
</script>

<svelte:head>
    <title>Chat | Social Network</title>
</svelte:head>

<!-- Adjusted height to work with navbar and layout -->
<div class="container mx-auto p-4 flex-1 flex items-stretch" style="height: calc(100vh - 70px);">
    <Card class="w-full p-0 overflow-hidden shadow-md">
        <div class="flex h-full">
            <!-- Chat list (hidden on mobile when viewing a chat) -->
            {#if !isMobileView || (isMobileView && showChatList)}
                <div class="w-full md:w-80 lg:w-96 h-full border-r flex-shrink-0">
                    <ChatList onSelectChat={handleSelectChat} />
                </div>
            {/if}

            <!-- Chat window or empty state -->
            <div class="hidden md:flex md:flex-1 flex-col {isMobileView && !showChatList ? '!flex' : ''}">
                {#if loading}
                    <div class="h-full flex items-center justify-center">
                        <Spinner size="8" />
                    </div>
                {:else if selectedChat}
                    <div class="relative h-full flex flex-col">
                        {#if isMobileView}
                            <Button
                                    size="xs"
                                    class="absolute top-2 left-2 z-10"
                                    on:click={toggleView}
                            >
                                Back
                            </Button>
                        {/if}

                        <ChatWindow
                                chatId={selectedChat.id}
                                isGroup={selectedChat.isGroup}
                                recipientId={selectedChat.recipientId}
                                recipientName={selectedChat.name}
                                recipientAvatar={selectedChat.avatar}
                        />
                    </div>
                {:else}
                    <div class="h-full flex items-center justify-center">
                        <EmptyState
                                title="Select a conversation"
                                description="Choose a chat from the list or start a new conversation"
                                icon="chat"
                        >
                            <div class="flex justify-center gap-4 mt-4">
                                <Button color="light" on:click={() => goto('/users')}>
                                    {@html UsersIcon}
                                    <span class="ml-2">Find Users</span>
                                </Button>
                                <Button color="light" on:click={() => goto('/groups')}>
                                    {@html UsersGroupIcon}
                                    <span class="ml-2">Browse Groups</span>
                                </Button>
                            </div>
                        </EmptyState>
                    </div>
                {/if}
            </div>
        </div>
    </Card>
</div>