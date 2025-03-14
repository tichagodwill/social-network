<script lang="ts">
    import {onDestroy, onMount, createEventDispatcher} from 'svelte'
    import {scale} from 'svelte/transition'
    import {debounce} from 'lodash-es'
    import {Avatar, Badge, Button, Card, Spinner} from 'flowbite-svelte'
    import {browser} from '$app/environment'
    import {auth} from '$lib/stores/auth'
    import {get} from 'svelte/store'
    import defaultProfileImg from '$lib/assets/default-profile.jpg'
    import {
        type ChatMessage,
        connectionState,
        ConnectionState,
        getChatMessages,
        type GroupChatMessage,
        initializeWebSocket,
        MessageType,
        resetUnreadCount,
        sendMessage,
        currentChatId,
        messages as globalMessages
    } from '$lib/stores/websocket'

    const dispatch = createEventDispatcher();

    let Picker: any
    let emojiPickerLoaded = false;
    let emojiData: any;

    // Props
    export let chatId: number;
    export let isGroup: boolean = false;
    export let recipientId: number | null = null; // For direct chats
    export let recipientName: string = '';
    export let recipientAvatar: string | null = null;

    // Component state
    let messageText: string = '';
    let messages: (ChatMessage | GroupChatMessage)[] = [];
    let loadedHistoricalMessages: (ChatMessage | GroupChatMessage)[] = [];
    let messagesContainer: HTMLElement;
    let loading: boolean = true;
    let isShowingEmojiPicker: boolean = false;
    let isTyping: boolean = false;
    let lastTypingSignalSent: number = 0;

    // Get the current user details from auth store
    function getCurrentUserId(): number {
        const authState = get(auth);
        return authState.user?.id || 0;
    }

    function getCurrentUserName(): string {
        const authState = get(auth);
        if (authState.user) {
            return `${authState.user.firstName} ${authState.user.lastName}`;
        }
        return 'Unknown User';
    }

    // Set current user info
    let currentUserId = getCurrentUserId();
    let currentUserName = getCurrentUserName();

    const FaceSmileOutline = `
  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
    <path stroke-linecap="round" stroke-linejoin="round" d="M15.182 15.182a4.5 4.5 0 0 1-6.364 0M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0ZM9 9h.01M15 9h.01" />
  </svg>
  `;

    // Replace PaperAirplaneOutline with SVG
    const PaperAirplaneOutline = `
  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
    <path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5" />
  </svg>
  `;

    // Helper functions for message types
    function isGroupChatMessage(message: any): message is GroupChatMessage {
        return message.type === MessageType.GROUP_CHAT;
    }

    function isChatMessage(message: any): message is ChatMessage {
        return message.type === MessageType.CHAT;
    }

    // Subscribe to messages for this chat
    const chatMessages = getChatMessages(chatId, isGroup);
    const unsubscribe = chatMessages.subscribe(msgs => {
        // Combine loaded historical messages with new messages from websocket
        const wsMessages = msgs as (ChatMessage | GroupChatMessage)[];

        // Deduplicate messages by ID
        const messageMap = new Map();

        // First add historical messages
        loadedHistoricalMessages.forEach(msg => {
            if (msg.id) {
                messageMap.set(msg.id, msg);
            } else {
                // Use a unique key for messages without ID
                const tempKey = `${msg.createdAt}-${msg.content}`;
                messageMap.set(tempKey, msg);
            }
        });

        // Then add new messages (they will override historical ones with same ID)
        wsMessages.forEach(msg => {
            if (msg.id) {
                messageMap.set(msg.id, msg);
            } else {
                // Use a unique key for messages without ID
                const tempKey = `${msg.createdAt}-${msg.content}`;
                messageMap.set(tempKey, msg);
            }
        });

        // Convert back to array and sort by timestamp
        messages = Array.from(messageMap.values())
            .sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());

        if (messages.length > 0 && messagesContainer) {
            setTimeout(scrollToBottom, 100);
        }
    });

    // Typing indicator debounced function
    const sendTypingIndicator = debounce(() => {
        if (!isTyping) {
            isTyping = true;
            sendTypingSignal(true);
        }

        // Set a timeout to clear the typing indicator
        setTimeout(() => {
            isTyping = false;
            sendTypingSignal(false);
        }, 3000);
    }, 300);

    // Function to send typing signal
    function sendTypingSignal(isTyping: boolean) {
        const now = Date.now();
        // Limit how often we send typing signals (max once per second)
        if (now - lastTypingSignalSent < 1000) {
            return;
        }

        lastTypingSignalSent = now;

        if (!recipientId && !isGroup) {
            console.error('Cannot send typing indicator: recipient ID is missing');
            return;
        }

        sendMessage({
            type: MessageType.TYPING,
            senderId: currentUserId,
            recipientId: recipientId || 0,
            isTyping,
            createdAt: new Date().toISOString()
        });
    }

    // Handle message input changes
    function handleInput() {
        if (messageText.length > 0) {
            sendTypingIndicator();
        }
    }

    // Function to send a chat message
    function sendChatMessage() {
        if (!messageText.trim()) return;

        const now = new Date().toISOString();

        let messageToSend: ChatMessage | GroupChatMessage;

        if (isGroup) {
            messageToSend = {
                type: MessageType.GROUP_CHAT,
                groupId: chatId,
                userId: currentUserId,
                content: messageText.trim(),
                createdAt: now,
                userName: currentUserName
            };
        } else if (recipientId) {
            messageToSend = {
                type: MessageType.CHAT,
                chatId: chatId, // Use the chatId from the server
                senderId: currentUserId,
                recipientId,
                content: messageText.trim(),
                createdAt: now,
                senderName: currentUserName
            };
        } else {
            console.error('Cannot send message: recipient ID is missing');
            return;
        }

        // Send message
        const success = sendMessage(messageToSend);

        if (success) {
            messageText = '';
            isTyping = false;
        }
    }

    // Handle emoji selection
    function handleEmojiSelect(event: CustomEvent) {
        messageText += event.detail.native;
        isShowingEmojiPicker = false;
    }

    // Toggle emoji picker
    function toggleEmojiPicker() {
        isShowingEmojiPicker = !isShowingEmojiPicker;
    }

    // Load historical messages
    async function loadMessages() {
        loading = true;
        try {
            if (!currentUserId) {
                console.error('Missing current user ID for loading messages');
                loading = false;
                return;
            }

            // Clear previous messages
            loadedHistoricalMessages = [];
            messages = [];

            let response;

            if (isGroup) {
                // For group chats, we only need the chatId, not a recipientId
                response = await fetch(
                    `http://localhost:8080/group-chats/${chatId}`,
                    {credentials: 'include'}
                );
            } else {
                // For direct chats, we need both users
                if (!recipientId) {
                    console.error('Missing recipient ID for loading direct messages');
                    loading = false;
                    return;
                }

                response = await fetch(
                    `http://localhost:8080/messages/${currentUserId}/${recipientId}`,
                    {credentials: 'include'}
                );
            }

            if (response.ok) {
                try {
                    const data = await response.json();

                    // Check if we got the expected response format with messages and chatId
                    if (data && data.chatId) {
                        // Don't override the prop directly
                        const newChatId = data.chatId;
                        // Only update the subscription if the ID actually changed
                        if (newChatId !== chatId) {
                            console.log(`Chat ID changed from ${chatId} to ${newChatId}`);
                            // Update the parent component about this change
                            dispatch('chatIdChanged', {oldId: chatId, newId: newChatId});
                            chatId = newChatId;
                        }

                        // Process messages
                        if (Array.isArray(data.messages)) {
                            data.messages.forEach((msg: any) => {
                                // Only set default type if it's missing
                                if (!msg.type) {
                                    msg.type = isGroup ? MessageType.GROUP_CHAT : MessageType.CHAT;
                                }
                            });

                            // Store in historical messages
                            loadedHistoricalMessages = data.messages;
                            console.log(`Loaded ${data.messages.length} messages from chat ID ${chatId}`);
                        } else {
                            console.log('No messages found or empty array');
                            loadedHistoricalMessages = [];
                        }
                    }
                    // Handle backward compatibility - old format with just an array of messages
                    else if (Array.isArray(data)) {
                        data.forEach((msg: any) => {
                            msg.type = MessageType.CHAT;
                        });

                        // If we have messages and chatId in the first message, update local chatId
                        if (data.length > 0 && data[0].chatId) {
                            chatId = data[0].chatId;
                        }

                        loadedHistoricalMessages = data;
                        console.log(`Loaded ${data.length} messages`);
                    } else {
                        console.log('Unexpected response format', data);
                        loadedHistoricalMessages = [];
                    }

                    // Now combine with any websocket messages
                    const wsMessages = get(chatMessages);
                    const messageMap = new Map();

                    // First add historical messages
                    loadedHistoricalMessages.forEach(msg => {
                        if (msg.id) {
                            messageMap.set(msg.id, msg);
                        } else {
                            // Use a unique key for messages without ID
                            const tempKey = `${msg.createdAt}-${msg.content}`;
                            messageMap.set(tempKey, msg);
                        }
                    });

                    // Then add new messages (they will override historical ones with same ID)
                    wsMessages.forEach(msg => {
                        if (msg.id) {
                            messageMap.set(msg.id, msg);
                        } else {
                            // Use a unique key for messages without ID
                            const tempKey = `${msg.createdAt}-${msg.content}`;
                            messageMap.set(tempKey, msg);
                        }
                    });

                    // Convert back to array and sort by timestamp
                    messages = Array.from(messageMap.values())
                        .sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());

                } catch (parseError) {
                    console.error('Error parsing messages:', parseError);
                    loadedHistoricalMessages = [];
                }
            } else {
                console.log(`Server returned ${response.status}: ${response.statusText}`);
                loadedHistoricalMessages = [];
            }
        } catch (error) {
            console.error('Error loading messages:', error);
            loadedHistoricalMessages = [];
        } finally {
            loading = false;
            setTimeout(scrollToBottom, 100);
        }
    }

    // Format timestamp for display
    function formatTimestamp(timestamp: string): string {
        const date = new Date(timestamp);
        const now = new Date();
        const yesterday = new Date(now);
        yesterday.setDate(yesterday.getDate() - 1);

        // Same day
        if (date.toDateString() === now.toDateString()) {
            return date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'});
        }

        // Yesterday
        if (date.toDateString() === yesterday.toDateString()) {
            return `Yesterday, ${date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'})}`;
        }

        // This week (within 7 days)
        const diffTime = Math.abs(now.getTime() - date.getTime());
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

        if (diffDays <= 7) {
            return `${date.toLocaleDateString([], {weekday: 'short'})}, ${date.toLocaleTimeString([], {
                hour: '2-digit',
                minute: '2-digit'
            })}`;
        }

        // Older
        return date.toLocaleDateString([], {year: 'numeric', month: 'short', day: 'numeric'});
    }

    // Get message sender's avatar
    function getMessageAvatar(message: ChatMessage | GroupChatMessage): string {
        if (isGroupChatMessage(message)) {
            return message.userAvatar || defaultProfileImg;
        } else {
            return message.senderAvatar || defaultProfileImg;
        }
    }

    // Get message sender's name
    function getMessageSenderName(message: ChatMessage | GroupChatMessage): string {
        if (isGroupChatMessage(message)) {
            return message.userName || 'User';
        } else {
            return message.senderName || 'User';
        }
    }

    // Check if message is from current user
    function isOwnMessage(message: ChatMessage | GroupChatMessage): boolean {
        if (isGroupChatMessage(message)) {
            return message.userId === currentUserId;
        } else {
            return message.senderId === currentUserId;
        }
    }

    // Scroll to the bottom of the messages container
    function scrollToBottom() {
        if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
    }

    // Handle key press events
    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            sendChatMessage();
        }
    }

    // Reset unread messages count when chat is opened
    function resetUnread() {
        resetUnreadCount(chatId, isGroup);
    }

    onMount(async () => {
        if (browser) {
            try {
                // Load emoji picker dynamically
                const [emojiMartModule, emojiDataModule] = await Promise.all([
                    import('emoji-mart'),
                    import('@emoji-mart/data')
                ]);
                Picker = emojiMartModule.default;
                emojiData = emojiDataModule.default;
                emojiPickerLoaded = true;
            } catch (error) {
                console.error('Failed to load emoji picker:', error);
            }

            // Initialize WebSocket if not already connected
            initializeWebSocket();

            // Update user info
            currentUserId = getCurrentUserId();
            currentUserName = getCurrentUserName();
            currentChatId.set(chatId);
            // Load message history
            await loadMessages();

            // Reset unread count
            resetUnread();
        }
    });
    $: {
        if (chatId) {
            currentChatId.set(chatId);

        }
    }
    onDestroy(() => {
        // Clean up subscription
        unsubscribe();
        currentChatId.set(null);
    });
</script>
<Card class="w-full h-full flex flex-col overflow-hidden max-w-full">
    <!-- Chat header -->
    {#if isGroup}
        <div class="flex items-center justify-between p-4 border-b">
            <div class="flex items-center gap-3">
                <Avatar src={recipientAvatar || defaultProfileImg} alt={recipientName} class="w-10 h-10" rounded/>
                <div>
                    <h3 class="text-lg font-semibold">{recipientName}</h3>
                    {#if $connectionState === ConnectionState.OPEN}
                        <Badge color="green" class="text-xs">Online</Badge>
                    {:else if $connectionState === ConnectionState.CONNECTING}
                        <Badge color="yellow" class="text-xs">Connecting...</Badge>
                    {:else}
                        <Badge color="gray" class="text-xs">Offline</Badge>
                    {/if}
                </div>
            </div>
        </div>
    {:else}
        <div class="flex items-center justify-between p-4 border-b">
            <div class="flex items-center gap-3">
                <Avatar src={recipientAvatar || defaultProfileImg} alt={recipientName} class="w-10 h-10" rounded/>
                <div>
                    <h3 class="text-lg font-semibold">{recipientName}</h3>
                    {#if $connectionState === ConnectionState.OPEN}
                        <Badge color="green" class="text-xs">Online</Badge>
                    {:else if $connectionState === ConnectionState.CONNECTING}
                        <Badge color="yellow" class="text-xs">Connecting...</Badge>
                    {:else}
                        <Badge color="gray" class="text-xs">Offline</Badge>
                    {/if}
                </div>
            </div>
        </div>
    {/if}
    <div
            bind:this={messagesContainer}
            class="flex-1 p-4 overflow-y-auto custom-scrollbar space-y-4 w-full"
            style="min-height: 300px;"
    >
        {#if loading}
            <div class="flex justify-center items-center h-full">
                <Spinner size="6"/>
            </div>
        {:else if messages.length === 0}
            <div class="flex justify-center items-center h-full text-gray-500">
                <p>No messages yet. Start the conversation!</p>
            </div>
        {:else}
            {#each messages as message, i (message.id || `${message.createdAt}-${i}`)}
                {#if isGroup}
                    {#if isOwnMessage(message)}
                        <div class="flex justify-end" transition:scale={{duration: 150, start: 0.95}}>
                            <div class="flex flex-row-reverse items-end gap-2 max-w-[80%]">
                                <div class="flex flex-col items-end">
                                    <div class="px-4 py-2 rounded-2xl bg-primary-500 text-white rounded-tr-none">
                                        <p class="whitespace-pre-wrap break-words">{message.content}</p>
                                    </div>
                                    <span class="text-xs text-gray-500 mt-1 mx-1">
                    {formatTimestamp(message.createdAt)}
                  </span>
                                </div>
                            </div>
                        </div>
                    {:else}
                        <div class="flex justify-start" transition:scale={{duration: 150, start: 0.95}}>
                            <div class="flex flex-row items-end gap-2 max-w-[80%]">
                                <Avatar
                                        src={getMessageAvatar(message)}
                                        alt={getMessageSenderName(message)}
                                        class="w-8 h-8"
                                        rounded
                                />
                                <div class="flex flex-col items-start">
                  <span class="text-xs text-gray-500 ml-1 mb-1">
                    {getMessageSenderName(message)}
                  </span>
                                    <div class="px-4 py-2 rounded-2xl bg-gray-100 dark:bg-gray-700 rounded-tl-none">
                                        <p class="whitespace-pre-wrap break-words">{message.content}</p>
                                    </div>
                                    <span class="text-xs text-gray-500 mt-1 mx-1">
                    {formatTimestamp(message.createdAt)}
                  </span>
                                </div>
                            </div>
                        </div>
                    {/if}
                {:else}
                    {#if isOwnMessage(message)}
                        <div class="flex justify-end" transition:scale={{duration: 150, start: 0.95}}>
                            <div class="flex flex-row-reverse items-end gap-2 max-w-[80%]">
                                <div class="flex flex-col items-end">
                                    <div class="px-4 py-2 rounded-2xl bg-primary-500 text-white rounded-tr-none">
                                        <p class="whitespace-pre-wrap break-words">{message.content}</p>
                                    </div>
                                    <span class="text-xs text-gray-500 mt-1 mx-1">
                    {formatTimestamp(message.createdAt)}
                  </span>
                                </div>
                            </div>
                        </div>
                    {:else}
                        <div class="flex justify-start" transition:scale={{duration: 150, start: 0.95}}>
                            <div class="flex flex-row items-end gap-2 max-w-[80%]">
                                <Avatar
                                        src={getMessageAvatar(message)}
                                        alt={getMessageSenderName(message)}
                                        class="w-8 h-8"
                                        rounded
                                />
                                <div class="flex flex-col items-start">
                                    <div class="px-4 py-2 rounded-2xl bg-gray-100 dark:bg-gray-700 rounded-tl-none">
                                        <p class="whitespace-pre-wrap break-words">{message.content}</p>
                                    </div>
                                    <span class="text-xs text-gray-500 mt-1 mx-1">
                    {formatTimestamp(message.createdAt)}
                  </span>
                                </div>
                            </div>
                        </div>
                    {/if}
                {/if}
            {/each}

            <!-- Typing indicator -->
            {#if isTyping}
                <div class="flex justify-start" in:scale={{duration: 150, start: 0.95}}>
                    <div class="bg-gray-100 dark:bg-gray-700 px-4 py-2 rounded-2xl rounded-tl-none max-w-[80%] flex items-center">
            <span class="typing-indicator">
              <span></span>
              <span></span>
              <span></span>
            </span>
                    </div>
                </div>
            {/if}
        {/if}
    </div>

    <div class="p-4 border-t">
        <div class="relative">
            {#if isShowingEmojiPicker && emojiPickerLoaded && browser}
                <div
                        class="absolute bottom-full right-0 mb-2"
                        transition:scale={{duration: 150, start: 0.9, opacity: 0}}
                >
                    <svelte:component
                            this={Picker}
                            data={emojiData}
                            onEmojiSelect={handleEmojiSelect}
                            theme="light"
                            set="native"
                    />
                </div>
            {/if}

            <div class="flex items-center gap-2">
                <Button
                        class="rounded-full min-w-10 flex-shrink-0"
                        color="light"
                        on:click={toggleEmojiPicker}
                >
                    {@html FaceSmileOutline}
                </Button>

                <textarea
                        bind:value={messageText}
                        on:input={handleInput}
                        on:keydown={handleKeyDown}
                        placeholder="Type a message..."
                        class="input resize-none py-2 max-h-32 w-full"
                        rows="1"
                        style="min-height: 42px"
                ></textarea>

                <Button
                        class="rounded-full min-w-10 flex-shrink-0"
                        color="primary"
                        disabled={!messageText.trim()}
                        on:click={sendChatMessage}
                >
                    {@html PaperAirplaneOutline}
                </Button>
            </div>
        </div>
    </div>
</Card>