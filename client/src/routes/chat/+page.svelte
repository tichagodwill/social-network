<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { chat } from '$lib/stores/chat';
    import { auth } from '$lib/stores/auth';
    import { Button, Avatar } from 'flowbite-svelte';
    import { getLastDate } from '$lib/dateFormater';
    import EmojiPicker from '$lib/components/EmojiPicker.svelte';
    import type { EmojiPickerEvent } from '$lib/types';
    import ChatInput from '$lib/components/ChatInput.svelte';
    import FileUpload from '$lib/components/FileUpload.svelte';
    import MessageContent from '$lib/components/MessageContent.svelte';
    import type { FileUploadResponse } from '$lib/types';
    import DragDropZone from '$lib/components/DragDropZone.svelte';
    import defualtProfileImg from '$lib/assets/defualt-profile.jpg'

    let newMessage = '';
    let chatInput: ChatInput;
    let dragDropActive = false;
    const userId = $auth.user!.id

    onMount(() => {
        chat.initialize();
        chat.loadContacts(userId);
    });

    onDestroy(() => {
        chat.cleanup();
    });

    function handleSend() {
        if (!newMessage.trim() || !$chat.activeChat) return;
        
        chat.sendMessage(newMessage, userId, $chat.activeChat);
        newMessage = '';
    }

    function handleEmojiSelect(event: EmojiPickerEvent) {
        const emoji = event.detail.emoji.native;
        const cursorPosition = chatInput.getCursorPosition();
        newMessage = 
            newMessage.slice(0, cursorPosition) + 
            emoji + 
            newMessage.slice(cursorPosition);
        chatInput.focus();
    }

    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            handleSend();
        }
    }

    function handleFileUpload(event: CustomEvent<FileUploadResponse>) {
        const { url, fileName, fileType } = event.detail;
        chat.sendMessage('', userId, $chat.activeChat, { url, fileName, fileType });
    }

    function handleDragEnter() {
        dragDropActive = true;
    }

    function handleDragLeave() {
        dragDropActive = false;
    }
</script>

<DragDropZone 
    bind:active={dragDropActive}
    on:upload={handleFileUpload}
>
    <div class="container mx-auto px-4 py-8">
        <div 
            class="grid grid-cols-12 gap-4 h-[calc(100vh-12rem)]"
            role="presentation"
            on:dragenter={handleDragEnter}
            on:dragleave={handleDragLeave}
        >
            <!-- Contacts List -->
            <div class="col-span-3 bg-white dark:bg-gray-800 rounded-lg shadow overflow-y-auto">
                <div class="p-4">
                    <h2 class="text-lg font-semibold mb-4">Contacts</h2>
                    {#each $chat.contacts as contact}
                        <button
                            class="w-full text-left p-3 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg mb-2 flex items-center space-x-3"
                            class:bg-gray-100={$chat.activeChat === contact.id}
                            on:click={() => chat.loadMessages(userId, contact.id)}
                        >
                            <Avatar src={contact.avatar || defualtProfileImg} size="sm" />
                            <div>
                                <p class="font-medium">{contact.username}</p>
                                <p class="text-sm text-gray-500">{contact.firstName} {contact.lastName}</p>
                            </div>
                        </button>
                    {/each}
                </div>
            </div>

            <!-- Chat Area -->
            <div class="col-span-9 bg-white dark:bg-gray-800 rounded-lg shadow flex flex-col">
                {#if $chat.activeChat}
                    <div class="flex-1 overflow-y-auto p-4">
                        {#each $chat.messages as message}
                            <div class="mb-4 flex" class:justify-end={message.senderId === $auth.user?.id}>
                                <div class="max-w-[70%] bg-gray-100 dark:bg-gray-700 rounded-lg p-3">
                                    <MessageContent {message} />
                                    <p class="text-xs text-gray-500 mt-1">
                                        {getLastDate(new Date(message.createdAt))}
                                    </p>
                                </div>
                            </div>
                        {/each}
                    </div>
                    <div class="p-4 border-t dark:border-gray-700">
                        <div class="flex space-x-2">
                            <ChatInput
                                bind:this={chatInput}
                                bind:value={newMessage}
                                placeholder="Type a message..."
                                on:keypress={handleKeyPress}
                            />
                            <FileUpload on:upload={handleFileUpload} />
                            <EmojiPicker on:emoji-select={handleEmojiSelect} />
                            <Button on:click={handleSend}>Send</Button>
                        </div>
                    </div>
                {:else}
                    <div class="flex-1 flex items-center justify-center">
                        <p class="text-gray-500">Select a contact to start chatting</p>
                    </div>
                {/if}
            </div>
        </div>
    </div>
</DragDropZone>
