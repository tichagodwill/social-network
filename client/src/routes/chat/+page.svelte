<script lang="ts">
    import {onMount, onDestroy} from 'svelte';
    import {chat} from '$lib/stores/chat';
    import {auth} from '$lib/stores/auth';
    import {Button, Avatar} from 'flowbite-svelte';
    import {getLastDate} from '$lib/dateFormater';
    import {MessageSquare, Send, Check, CheckCheck, Clock, Download} from 'lucide-svelte';
    import EmojiPicker from '$lib/components/EmojiPicker.svelte';
    import type {EmojiPickerEvent, Message, User} from '$lib/types';
    import ChatInput from '$lib/components/ChatInput.svelte';
    import FileUpload from '$lib/components/FileUpload.svelte';
    import DragDropZone from '$lib/components/DragDropZone.svelte';
    import type {FileUploadResponse} from '$lib/types';
    import defaultProfileImg from '$lib/assets/default-profile.jpg';

    export let loadContact: number | null = null;
    let newMessage = '';
    let chatInput: ChatInput;
    let contact: User | null = null;
    let messagesContainer: HTMLElement;
    let dragDropActive = false;
    const userId = $auth.user?.id;

    // Auto-scroll to bottom when new messages arrive
    $: if ($chat.messages && messagesContainer) {
        setTimeout(() => {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }, 0);
    }

    // Watch for changes in loadContact prop and contacts
    $: if (loadContact && userId && $chat.contacts) {  // Remove .length check
        const selectedContact = $chat.contacts.find(c => c.id === loadContact);
        if (selectedContact && (!contact || contact.id !== selectedContact.id)) {
            selectContact(selectedContact);
        }
    }

    onMount(async () => {
        if (!userId) return;

        try {
            await chat.initialize();
            const contacts = await chat.loadContacts(userId);

            if (loadContact && contacts && contacts.length > 0) {
                const selectedContact = contacts.find(c => c.id === loadContact);
                if (selectedContact) {
                    await selectContact(selectedContact);
                }
            }
        } catch (error) {
            console.error('Error initializing chat:', error);
        }
    });

    onDestroy(() => chat.cleanup());

    function handleSend() {
        if (!newMessage.trim() && !fileToUpload) return;
        if (!contact || !userId) return;

        if (fileToUpload) {
            chat.sendMessage('', userId, contact.id, fileToUpload);
            fileToUpload = null;
        } else {
            chat.sendMessage(newMessage, userId, contact.id);
        }

        newMessage = '';
        chatInput.focus();
    }

    let fileToUpload: File | null = null;

    const handleFileUpload = (event: CustomEvent<FileUploadResponse>) => {
        fileToUpload = event.detail.file;
    };

    const handleEmojiSelect = (event: EmojiPickerEvent) => {
        const emoji = event.detail.emoji.native;
        const pos = chatInput.getCursorPosition();
        newMessage = newMessage.slice(0, pos) + emoji + newMessage.slice(pos);
        chatInput.focus();
    };

    const handleKeydown = (e: KeyboardEvent) => {
        if (!contact || !userId) return;

        chat.setTyping(userId, contact.id, true);
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    const selectContact = async (selectedContact: User) => {
        if (!userId) return;

        try {
            const result = await chat.getOrCreateDirectChat(selectedContact.id);
            if (result.error) {
                console.error('Failed to create chat:', result.error);
                return;
            }

            contact = selectedContact;
            await chat.loadMessages(userId, selectedContact.id);

            // Update the URL without reloading the page
            const url = new URL(window.location.href);
            url.pathname = `/chat/${selectedContact.id}`;
            window.history.pushState({}, '', url.toString());
        } catch (error) {
            console.error('Failed to load messages:', error);
        }
    };

    function formatFileSize(bytes: number): string {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
    }

    async function downloadFile(message: Message) {
        if (!message.fileData) return;

        const blob = new Blob([message.fileData], {type: message.fileType});
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = message.fileName || 'download';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }
</script>

<style lang="postcss">
    .message-container {
        @apply flex flex-col space-y-2 p-4;
    }

    .message {
        @apply max-w-[75%] rounded-lg p-3;
    }

    .sent {
        @apply ml-auto bg-primary-500 text-white;
    }

    .received {
        @apply bg-gray-100 dark:bg-gray-700;
    }

    .typing-indicator {
        @apply text-sm text-gray-500 italic;
    }
</style>

<DragDropZone bind:active={dragDropActive} on:upload={handleFileUpload}>
    <div class="container mx-auto px-4 py-8">
        <div class="grid grid-cols-12 gap-6 h-[calc(100vh-10rem)]">
            <!-- Contact List -->
            <div class="col-span-3 bg-white dark:bg-gray-800 rounded-xl shadow-lg overflow-hidden">
                <div class="sticky top-0 bg-white dark:bg-gray-800 z-10 p-4 border-b dark:border-gray-700">
                    <h2 class="text-xl font-semibold">Messages</h2>
                    <p class="text-sm text-gray-500 mt-1">{$chat.contacts.length} contacts</p>
                </div>

                <div class="p-4 overflow-y-auto h-[calc(100vh-16rem)]">
                    {#each $chat.contacts as c}
                        <button
                                class="w-full text-left p-3 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-xl mb-3 flex items-center space-x-3 transition-all
                            {contact?.id === c.id ? 'bg-primary-50 dark:bg-primary-900/20 border-l-4 border-primary-500' : ''}"
                                on:click={() => selectContact(c)}
                        >
                            <div class="relative">
                                <Avatar src={c.avatar || defaultProfileImg} size="md" class="ring-2 ring-gray-100"/>
                                <div class="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-white"></div>
                            </div>
                            <div class="flex-1 min-w-0">
                                <p class="font-medium truncate">{c.username}</p>
                                <p class="text-sm text-gray-500 truncate">
                                    {c.firstName} {c.lastName}
                                </p>
                            </div>
                        </button>
                    {/each}
                    <!-- no contacts-->
                    {#if $chat.contacts.length === 0}
                        <div class="flex-1 flex items-center justify-center p-6 text-center">
                            <p class="text-gray-500">Follow some users to start chatting</p>
                        </div>
                    {/if}
                </div>

            </div>

            <!-- Chat Area -->
            <div class="col-span-9 bg-white dark:bg-gray-800 rounded-xl shadow-lg flex flex-col overflow-hidden">
                {#if contact}
                    <!-- Chat Header -->
                    <div class="sticky top-0 z-20 px-6 py-4 bg-white dark:bg-gray-800 border-b dark:border-gray-700">
                        <div class="flex items-center space-x-4">
                            <Avatar src={contact.avatar || defaultProfileImg} size="md"/>
                            <div>
                                <h3 class="text-lg font-semibold">{contact.username}</h3>
                                <p class="text-sm text-gray-500">
                                    {#if $chat.typingUsers.has(contact.id)}
                                        typing...
                                    {:else}
                                        {contact.firstName} {contact.lastName}
                                    {/if}
                                </p>
                            </div>
                        </div>
                    </div>

                    <!-- Messages -->
                    <div
                            class="flex-1 overflow-y-auto px-6 py-4"
                            id="messages"
                            bind:this={messagesContainer}
                    >
                        {#each $chat.messages as message, i}
                            {@const isFirstInGroup = i === 0 || $chat.messages[i - 1].senderId !== message.senderId}
                            {@const isLastInGroup = i === $chat.messages.length - 1 || $chat.messages[i + 1].senderId !== message.senderId}

                            <div class="mb-2 last:mb-0 flex" class:justify-end={message.senderId === userId}>
                                <div class="flex {message.senderId === userId ? 'flex-row-reverse' : 'flex-row'} items-end max-w-[75%]">
                                    {#if message.senderId !== userId}
                                        {#if isFirstInGroup}
                                            <Avatar src={message.senderAvatar || defaultProfileImg} size="sm"
                                                    class="mb-1 mr-2"/>
                                        {:else}
                                            <div class="w-8 mr-2"></div>
                                        {/if}
                                    {/if}

                                    <div class="
                                        {message.senderId === userId ? 'bg-primary-500 text-white' : 'bg-gray-100 dark:bg-gray-700'}
                                        p-3 shadow-sm
                                        {message.senderId === userId ? 'rounded-l-2xl' : 'rounded-r-2xl'}
                                        {isFirstInGroup ? (message.senderId === userId ? 'rounded-tr-2xl' : 'rounded-tl-2xl') : ''}
                                        {isLastInGroup ? (message.senderId === userId ? 'rounded-br-2xl' : 'rounded-bl-2xl') : ''}
                                        relative w-full
                                    ">
                                        {#if message.messageType === 'file'}
                                            <div class="flex items-center space-x-2">
                                                <div class="flex-1">
                                                    <p class="font-medium">{message.fileName}</p>
                                                    <p class="text-sm opacity-75">{formatFileSize(message.fileData?.byteLength || 0)}</p>
                                                </div>
                                                <button
                                                        class="p-2 hover:bg-black/10 rounded-full transition-colors"
                                                        on:click={() => downloadFile(message)}
                                                >
                                                    <Download size={20}/>
                                                </button>
                                            </div>
                                        {:else}
                                            <p class="whitespace-pre-wrap">{message.content}</p>
                                        {/if}

                                        <div class="flex items-center space-x-1 text-[10px] {message.senderId === userId ? 'text-primary-100' : 'text-gray-400'} mt-1">
                                            <span>{getLastDate(new Date(message.createdAt))}</span>
                                            {#if message.senderId === userId}
                                                <span>
                                                    {#if message.status === 'read'}
                                                        <CheckCheck size={12}/>
                                                    {:else if message.status === 'delivered'}
                                                        <Check size={12}/>
                                                    {:else}
                                                        <Clock size={12}/>
                                                    {/if}
                                                </span>
                                            {/if}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>

                    <!-- Input Area -->
                    <div class="p-4 bg-gray-50 dark:bg-gray-900 border-t dark:border-gray-700">
                        <div class="flex items-end space-x-2">
                            <div class="flex-1 bg-white dark:bg-gray-800 rounded-xl shadow-sm">
                                <ChatInput
                                        bind:this={chatInput}
                                        bind:value={newMessage}
                                        placeholder="Type your message..."
                                        class="w-full border-0 focus:ring-0 rounded-xl bg-transparent"
                                        on:keydown={(e) => {
                                        // Handle typing indicator
                                        chat.setTyping(userId, contact.id, true);
                                        if (e.key === 'Enter' && !e.shiftKey) {
                                            e.preventDefault();
                                            handleSend();
                                        }
                                    }}
                                />
                            </div>
                            <div class="flex space-x-2">
                                <FileUpload on:upload={handleFileUpload}/>
                                <EmojiPicker on:emoji-select={handleEmojiSelect}/>
                                <Button gradient color="primary" size="lg" on:click={handleSend}>
                                    <Send size={20}/>
                                </Button>
                            </div>
                        </div>
                    </div>
                {:else}
                    <!-- Empty State -->
                    <div class="flex-1 flex flex-col items-center justify-center p-6 text-center">
                        <div class="w-16 h-16 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center mb-4">
                            <MessageSquare size={32} class="text-gray-400"/>
                        </div>
                        {#if $chat.contacts.length === 0}
                            <h3 class="text-xl font-semibold mb-2">No Contacts</h3>
                            <p class="text-gray-500">Follow some users to start chatting</p>
                        {:else}
                            <h3 class="text-xl font-semibold mb-2">Start a Conversation</h3>
                            <p class="text-gray-500">Select a contact to begin messaging</p>
                        {/if}
                    </div>
                {/if}
            </div>
        </div>
    </div>
</DragDropZone>