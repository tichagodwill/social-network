<script lang="ts">
    import {onDestroy, onMount, createEventDispatcher} from 'svelte';
    import {scale, fade} from 'svelte/transition';
    import {debounce} from 'lodash-es';
    import {Avatar, Badge, Button, Card, Spinner} from 'flowbite-svelte';
    import {browser} from '$app/environment';
    import {auth} from '$lib/stores/auth';
    import {get} from 'svelte/store';
    import defaultProfileImg from '$lib/assets/default-profile.jpg';
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
        processedChatIds,
        messages as globalMessages,
        cleanupWebSocketResources
    } from '$lib/stores/websocket';

    const dispatch = createEventDispatcher();

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
    let showEmojiPanel: boolean = false;
    let isTyping: boolean = false;
    let lastTypingSignalSent: number = 0;
    let activeEmojiCategory: string = 'Smileys';

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

    // SVG Icons
    const FaceSmileOutline = `
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.182 15.182a4.5 4.5 0 0 1-6.364 0M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0ZM9 9h.01M15 9h.01" />
        </svg>
    `;

    const PaperAirplaneOutline = `
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5" />
        </svg>
    `;

    // Custom emoji collection
    const emojiCategories = {
        "Smileys": ["😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "☺️", "😊", "😇", "🙂", "🙃", "😉", "😌", "😍", "🥰", "😘", "😗", "😙", "😚"],
        "Emotions": ["😋", "😛", "😝", "😜", "🤪", "🤨", "🧐", "🤓", "😎", "🤩", "🥳", "😏", "😒", "😞", "😔", "😟", "😕", "🙁", "☹️", "😣", "😖"],
        "Hearts": ["❤️", "🧡", "💛", "💚", "💙", "💜", "🖤", "💔", "❣️", "💕", "💞", "💓", "💗", "💖", "💘", "💝", "💟", "♥️"],
        "Hands": ["👍", "👎", "👊", "✊", "🤛", "🤜", "👏", "🙌", "👐", "🤲", "🤝", "🙏", "✌️", "🤞", "🤟", "🤘", "👌", "🤌", "🤏", "👈", "👉"],
        "Celebrations": ["🎉", "🎊", "🎈", "🎂", "🎁", "🎄", "🎃", "🎗️", "🎟️", "🎫", "🎖️", "🏆", "🥇", "🥈", "🥉"],
        "Animals": ["🐶", "🐱", "🐭", "🐹", "🐰", "🦊", "🐻", "🐼", "🐨", "🐯", "🦁", "🐮", "🐷", "🐸", "🐵", "🙈", "🙉", "🙊", "🐔", "🐧", "🐦"],
        "Food": ["🍎", "🍐", "🍊", "🍋", "🍌", "🍉", "🍇", "🍓", "🍈", "🍒", "🍑", "🥭", "🍍", "🥥", "🥝", "🍅", "🥑", "🍆", "🌮", "🍕", "🍔"]
    };

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
        const wsMessages = msgs as (ChatMessage | GroupChatMessage)[];

        const messageMap = new Map();
        loadedHistoricalMessages.forEach(msg => {
            const key = msg.id || `${msg.createdAt}-${msg.content}`;
            messageMap.set(key, msg);
        });
        wsMessages.forEach(msg => {
            const key = msg.id || `${msg.createdAt}-${msg.content}`;
            messageMap.set(key, msg);
        });

        messages = Array.from(messageMap.values()).sort(
            (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
        );

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
        setTimeout(() => {
            isTyping = false;
            sendTypingSignal(false);
        }, 3000);
    }, 300);

    // Function to send typing signal
    function sendTypingSignal(isTyping: boolean) {
        const now = Date.now();
        if (now - lastTypingSignalSent < 1000) return;

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
                chatId: chatId,
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

        messages = [...messages, messageToSend];
        setTimeout(scrollToBottom, 10);

        const success = sendMessage(messageToSend);
        if (success) {
            messageText = '';
            isTyping = false;
        } else {
            console.error('Failed to send message via WebSocket');
        }
    }

    // Handle emoji selection
    function selectEmoji(emoji: string) {
        messageText += emoji;
        // Leave the emoji panel open to allow selecting multiple emojis
    }

    // Toggle emoji panel
    function toggleEmojiPanel() {
        showEmojiPanel = !showEmojiPanel;
    }

    // Change emoji category
    function changeEmojiCategory(category: string) {
        activeEmojiCategory = category;
    }

    // Load historical messages
    async function loadMessages() {
        loading = true;
        try {
            if (!currentUserId) {
                console.error('Missing current user ID for loading messages');
                return;
            }

            loadedHistoricalMessages = [];
            let response;
            if (isGroup) {
                response = await fetch(`http://localhost:8080/group-chats/${chatId}`, {credentials: 'include'});
            } else {
                if (!recipientId) {
                    console.error('Missing recipient ID for loading direct messages');
                    return;
                }
                response = await fetch(`http://localhost:8080/messages/${currentUserId}/${recipientId}`, {
                    credentials: 'include'
                });
            }

            if (response.ok) {
                const data = await response.json();
                console.log('[DEBUG] Loaded messages data:', data);

                globalMessages.update(existingMsgs =>
                    existingMsgs.filter(msg =>
                        isGroup
                            ? !(msg.type === MessageType.GROUP_CHAT && 'groupId' in msg && msg.groupId === chatId)
                            : !(msg.type === MessageType.CHAT && 'chatId' in msg && msg.chatId === chatId)
                    )
                );

                if (data && data.chatId) {
                    const newChatId = data.chatId;
                    if (newChatId !== chatId) {
                        console.log(`[DEBUG] Chat ID changed from ${chatId} to ${newChatId}`);
                        dispatch('chatIdChanged', {oldId: chatId, newId: newChatId});
                        chatId = newChatId;
                    }

                    if (Array.isArray(data.messages)) {
                        const processedMessages = data.messages.map((msg: any) => ({
                            ...msg,
                            type: msg.type || (isGroup ? MessageType.GROUP_CHAT : MessageType.CHAT)
                        }));
                        loadedHistoricalMessages = processedMessages;
                        globalMessages.update(msgs => [...msgs, ...processedMessages]);
                        console.log(`[DEBUG] Added ${processedMessages.length} historical messages to global store`);
                    }
                } else if (Array.isArray(data)) {
                    const processedMessages = data.map((msg: any) => ({
                        ...msg,
                        type: msg.type || MessageType.CHAT
                    }));
                    loadedHistoricalMessages = processedMessages;
                    globalMessages.update(msgs => [...msgs, ...processedMessages]);
                    console.log(`[DEBUG] Added ${processedMessages.length} historical messages to global store`);
                }
            }
        } catch (error) {
            console.error('[ERROR] Error loading messages:', error);
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

        if (date.toDateString() === now.toDateString()) {
            return date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'});
        }
        if (date.toDateString() === yesterday.toDateString()) {
            return `Yesterday, ${date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'})}`;
        }

        const diffDays = Math.ceil(Math.abs(now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));
        if (diffDays <= 7) {
            return `${date.toLocaleDateString([], {weekday: 'short'})}, ${date.toLocaleTimeString([], {
                hour: '2-digit',
                minute: '2-digit'
            })}`;
        }
        return date.toLocaleDateString([], {year: 'numeric', month: 'short', day: 'numeric'});
    }

    // Get message sender's avatar
    function getMessageAvatar(message: ChatMessage | GroupChatMessage): string {
        return isGroupChatMessage(message) ? message.userAvatar || defaultProfileImg : message.senderAvatar || defaultProfileImg;
    }

    // Get message sender's name
    function getMessageSenderName(message: ChatMessage | GroupChatMessage): string {
        return isGroupChatMessage(message) ? message.userName || 'User' : message.senderName || 'User';
    }

    // Check if message is from current user
    function isOwnMessage(message: ChatMessage | GroupChatMessage): boolean {
        return isGroupChatMessage(message) ? message.userId === currentUserId : message.senderId === currentUserId;
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
            initializeWebSocket();
            currentUserId = getCurrentUserId();
            currentUserName = getCurrentUserName();
            currentChatId.set(chatId);
            await loadMessages();
            resetUnread();
        }
        window.addEventListener('beforeunload', () => {
            if (document.visibilityState === 'hidden') {
                cleanupWebSocketResources();
            }
        });
    });

    $: if (chatId) currentChatId.set(chatId);

    onDestroy(() => {
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
                        <div class="flex justify-end" transition:scale={{ duration: 150, start: 0.95 }}>
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
                        <div class="flex justify-start" transition:scale={{ duration: 150, start: 0.95 }}>
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
                        <div class="flex justify-end" transition:scale={{ duration: 150, start: 0.95 }}>
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
                        <div class="flex justify-start" transition:scale={{ duration: 150, start: 0.95 }}>
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
                <div class="flex justify-start" in:scale={{ duration: 150, start: 0.95 }}>
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
        <div class="flex flex-col">
            <div class="flex items-center gap-2">
                <!-- Emoji Picker Button -->
                <Button
                        class="rounded-full min-w-10 flex-shrink-0"
                        color={showEmojiPanel ? "primary" : "light"}
                        on:click={toggleEmojiPanel}
                >
                    {@html FaceSmileOutline}
                </Button>

                <!-- Message Input -->
                <textarea
                        bind:value={messageText}
                        on:input={handleInput}
                        on:keydown={handleKeyDown}
                        placeholder="Type a message..."
                        class="input resize-none py-2 max-h-32 w-full rounded-lg border-gray-300 dark:border-gray-600"
                        rows="1"
                        style="min-height: 42px"
                ></textarea>

                <!-- Send Button -->
                <Button
                        class="rounded-full min-w-10 flex-shrink-0"
                        color="primary"
                        disabled={!messageText.trim()}
                        on:click={sendChatMessage}
                >
                    {@html PaperAirplaneOutline}
                </Button>
            </div>

            <!-- Emoji Panel -->
            {#if showEmojiPanel}
                <div class="emoji-panel mt-2 p-2 bg-white dark:bg-gray-800 rounded-lg border border-gray-300 dark:border-gray-700 shadow-lg"
                     transition:fade={{ duration: 150 }}>
                    <!-- Category tabs -->
                    <div class="emoji-categories mb-2 overflow-x-auto flex">
                        {#each Object.keys(emojiCategories) as category}
                            <button
                                    class="px-3 py-1 mr-1 text-sm rounded-full whitespace-nowrap {activeEmojiCategory === category ? 'bg-primary-100 dark:bg-primary-800 text-primary-700 dark:text-primary-300' : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}"
                                    on:click={() => changeEmojiCategory(category)}
                            >
                                {category}
                            </button>
                        {/each}
                    </div>

                    <!-- Emoji grid -->
                    <div class="emoji-grid grid grid-cols-8 gap-1 max-h-40 overflow-y-auto p-1">
                        {#each emojiCategories[activeEmojiCategory] as emoji}
                            <button
                                    class="emoji-btn w-9 h-9 flex items-center justify-center text-xl rounded hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer transition-colors"
                                    on:click={() => selectEmoji(emoji)}
                            >
                                {emoji}
                            </button>
                        {/each}
                    </div>
                </div>
            {/if}
        </div>
    </div>
</Card>

<style>
    /* Input styling */
    .input {
        background-color: #f9fafb; /* Light background */
        color: #111827; /* Dark text */
        transition: all 0.2s ease;
    }

    .input:focus {
        outline: none;
        border-color: #3b82f6; /* Blue focus ring */
    }

    .dark .input {
        background-color: #1f2937; /* Dark background */
        color: #f9fafb; /* Light text */
    }

    /* Typing indicator animation */
    .typing-indicator {
        display: flex;
        align-items: center;
    }

    .typing-indicator span {
        height: 8px;
        width: 8px;
        margin: 0 2px;
        background-color: #9ca3af;
        border-radius: 50%;
        display: inline-block;
        animation: typing 1.4s infinite ease-in-out both;
    }

    .typing-indicator span:nth-child(1) {
        animation-delay: 0s;
    }

    .typing-indicator span:nth-child(2) {
        animation-delay: 0.2s;
    }

    .typing-indicator span:nth-child(3) {
        animation-delay: 0.4s;
    }

    @keyframes typing {
        0%, 80%, 100% {
            transform: scale(0.7);
            opacity: 0.6;
        }
        40% {
            transform: scale(1);
            opacity: 1;
        }
    }

    /* Emoji panel styling */
    .emoji-panel {
        max-width: 100%;
    }

    .emoji-categories {
        scrollbar-width: thin;
    }

    .emoji-categories::-webkit-scrollbar {
        height: 4px;
    }

    .emoji-categories::-webkit-scrollbar-thumb {
        background-color: rgba(156, 163, 175, 0.5);
        border-radius: 4px;
    }

    .emoji-grid {
        scrollbar-width: thin;
    }

    .emoji-grid::-webkit-scrollbar {
        width: 4px;
    }

    .emoji-grid::-webkit-scrollbar-thumb {
        background-color: rgba(156, 163, 175, 0.5);
        border-radius: 4px;
    }

    .emoji-btn:hover {
        background-color: rgba(59, 130, 246, 0.1);
    }
</style>
