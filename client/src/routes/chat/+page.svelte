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
    import defualtProfileImg from '$lib/assets/defualt-profile.jpg';

    export let loadContact: number | null = null;
    let newMessage = '';
    let chatInput: ChatInput;
    let dragDropActive = false;
    let contact = null;
    const userId = $auth.user!.id;

    onMount(() => {
        chat.initialize();
        chat.loadContacts(userId);
        if (loadContact) {
            chat.loadMessages(userId, loadContact);
            contact = $chat.contacts.find(c => c.id === loadContact);
        }
    });

    onDestroy(() => chat.cleanup());

    function handleSend() {
        if (!newMessage.trim() || !contact) return;
        chat.sendMessage(newMessage, userId, contact.id);
        newMessage = '';
    }

    const handleFileUpload = (event: CustomEvent<FileUploadResponse>) => {
        const { url, fileName, fileType } = event.detail;
        if (contact) chat.sendMessage('', userId, contact.id, { url, fileName, fileType });
    };

    const handleEmojiSelect = (event: EmojiPickerEvent) => {
        const emoji = event.detail.emoji.native;
        const pos = chatInput.getCursorPosition();
        newMessage = newMessage.slice(0, pos) + emoji + newMessage.slice(pos);
        chatInput.focus();
    };

    const selectContact = (selectedContact) => {
        contact = selectedContact;
        chat.loadMessages(userId, selectedContact.id);
    };
</script>

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
                                <Avatar src={c.avatar || defualtProfileImg} size="md" class="ring-2 ring-gray-100" />
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
                </div>
            </div>

            <!-- Chat Area -->
            <div class="col-span-9 bg-white dark:bg-gray-800 rounded-xl shadow-lg flex flex-col overflow-hidden">
                {#if contact}
                    <!-- Chat Header -->
                    <div class="sticky top-0 z-20 px-6 py-4 bg-white dark:bg-gray-800 border-b dark:border-gray-700">
                        <div class="flex items-center space-x-4">
                            <Avatar src={contact.avatar || defualtProfileImg} size="md" />
                            <div>
                                <h3 class="text-lg font-semibold">{contact.username}</h3>
                                <p class="text-sm text-gray-500">
                                    {contact.firstName} {contact.lastName}
                                </p>
                            </div>
                        </div>
                    </div>

                    <!-- Messages -->
                    <div class="flex-1 overflow-y-auto p-6" id="messages">
                        {#each $chat.messages as message}
                            <div class="mb-6 flex" class:justify-end={message.senderId === userId}>
                                <div class="flex {message.senderId === userId ? 'flex-row-reverse' : 'flex-row'} items-end space-x-2">
                                    {#if message.senderId !== userId}
                                        <Avatar src={message.senderAvatar || defualtProfileImg} size="sm" />
                                    {/if}
                                    <div class="max-w-[70%] {message.senderId === userId ? 'bg-primary-500 text-white' : 'bg-gray-100 dark:bg-gray-700'} rounded-2xl p-4 shadow-sm">
                                        <MessageContent {message} />
                                        <p class="text-xs {message.senderId === userId ? 'text-primary-100' : 'text-gray-500'} mt-2">
                                            {getLastDate(new Date(message.createdAt))}
                                        </p>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>

                    <!-- Input Area -->
                    <div class="p-4 bg-gray-50 dark:bg-gray-900 border-t dark:border-gray-700">
                        <div class="flex space-x-2">
                            <ChatInput
                              bind:this={chatInput}
                              bind:value={newMessage}
                              placeholder="Type your message..."
                              class="flex-1 rounded-xl border-gray-200 focus:ring-primary-500 focus:border-primary-500"
                              on:keypress={(e) => e.key === 'Enter' && !e.shiftKey && handleSend()}
                            />
                            <div class="flex space-x-2">
                                <FileUpload on:upload={handleFileUpload} />
                                <EmojiPicker on:emoji-select={handleEmojiSelect} />
                                <Button gradient color="primary" size="lg" on:click={handleSend}>
                                    Send
                                </Button>
                            </div>
                        </div>
                    </div>
                {:else}
                    <div class="flex-1 flex flex-col items-center justify-center p-6 text-center">
                        <div class="w-16 h-16 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center mb-4">
                            <svg class="w-8 h-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                            </svg>
                        </div>
                        <h3 class="text-xl font-semibold mb-2">Start a Conversation</h3>
                        <p class="text-gray-500">Select a contact to begin messaging</p>
                    </div>
                {/if}
            </div>
        </div>
    </div>
</DragDropZone>