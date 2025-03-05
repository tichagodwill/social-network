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
            // Handle potential chats (negative IDs)
            if (chatId < 0) {
                // Convert to positive user ID
                const userId = Math.abs(chatId);
                
                // Create a chat room with this user
                const response = await fetch('http://localhost:8080/chat/direct', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({ userId })
                });
                
                if (response.ok) {
                    const data = await response.json();
                    // Use the newly created chat ID
                    chatId = data.id;
                    isGroup = false;
                } else {
                    console.error('Failed to create chat room');
                    loading = false;
                    return;
                }
            }

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
                // For direct chats, we need to get the other user's details
                const currentUserId = getCurrentUserId();

                // First, get the chat participants
                const chatResponse = await fetch(`http://localhost:8080/chat/${chatId}/participants`, {
                    credentials: 'include'
                });

                if (chatResponse.ok) {
                    const participants = await chatResponse.json();

                    // Find the other user (not the current user)
                    const otherUser = participants.find((p: any) => p.id !== currentUserId);

                    if (otherUser) {
                        // Fetch user details if needed
                        const response = await fetch(`http://localhost:8080/user/${otherUser.id}`, {
                            credentials: 'include'
                        });

                        if (response.ok) {
                            const userData = await response.json();
                            selectedChat = {
                                id: chatId,
                                isGroup: false,
                                recipientId: otherUser.id,
                                name: `${userData.user.first_name} ${userData.user.last_name}`,
                                avatar: userData.user.avatar
                            };
                        }
                    } else {
                        console.error('Could not find other participant in chat');
                    }
                } else {
                    // Fallback - if you don't have a participants endpoint yet
                    // We can create an alternative approach

                    // Create a temporary chat connection to get the recipient ID
                    const tempChatResponse = await fetch(`http://localhost:8080/messages/${currentUserId}/temp`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        credentials: 'include',
                        body: JSON.stringify({ chatId: chatId })
                    });

                    if (tempChatResponse.ok) {
                        const chatInfo = await tempChatResponse.json();
                        const otherUserId = chatInfo.recipientId;

                        // Now fetch the user details
                        const userResponse = await fetch(`http://localhost:8080/user/${otherUserId}`, {
                            credentials: 'include'
                        });

                        if (userResponse.ok) {
                            const userData = await userResponse.json();
                            selectedChat = {
                                id: chatId,
                                isGroup: false,
                                recipientId: otherUserId,
                                name: `${userData.user.first_name} ${userData.user.last_name}`,
                                avatar: userData.user.avatar
                            };
                        }
                    }
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

    function handleChatIdChange({ oldId, newId }: { oldId: number, newId: number }): void {
        if (selectedChat && selectedChat.id === oldId) {
            selectedChat = { ...selectedChat, id: newId };
            // Update URL params
            const newParams = new URLSearchParams();
            newParams.set('id', newId.toString());
            newParams.set('type', selectedChat.isGroup ? 'group' : 'direct');
            goto(`/chat?${newParams.toString()}`, { replaceState: true });
        }
    }
</script>

<svelte:head>
    <title>Chat | SocialNet</title>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            background-color: #f4f4f4;
            color: #333;
        }
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
        .navbar-spacer {
            height: 40px; /* Adjust this value based on your navbar height */
        }
        .chat-container {
            height: calc(90vh - 40px); /* Adjust 64px to match your navbar height */
            display: flex;
            flex-direction: column;
        }
    </style>
</svelte:head>
<div class="navbar-spacer"></div>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 max-h-screen chat-container">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden h-full flex flex-col mt-4">
        <div class="flex-1 flex overflow-hidden">
            {#if !isMobileView || (isMobileView && showChatList)}
                <div class="w-full md:w-80 lg:w-96 border-r border-gray-200 dark:border-gray-700 flex-shrink-0 h-full">
                    <ChatList onSelectChat={handleSelectChat} />
                </div>
            {/if}

            <div class="flex-1 {isMobileView ? (showChatList ? 'hidden' : 'flex') : 'flex'} flex-col">
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
                          on:chatIdChanged={(event) => handleChatIdChange(event.detail)}
                        />
                    </div>
                {:else}
                    <div class="h-full flex items-center justify-center p-4">
                        <EmptyState
                          title="Select a conversation"
                          description="Choose a chat from the list or start a new conversation"
                          icon="chat"
                        >
                            <div class="flex justify-center gap-4 mt-4">
                                <Button color="light" on:click={() => goto('/explore')}>
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
    </div>
</div>