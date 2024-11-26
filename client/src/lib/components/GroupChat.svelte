<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { Card, Button, Input } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { getLastDate } from '$lib/dateFormater';
    import EmojiPicker from './EmojiPicker.svelte';
    import type { EmojiPickerEvent } from '$lib/types';

    export let groupId: number;
    let messages: any[] = [];
    let newMessage = '';
    let socket: WebSocket;
    let chatContainer: HTMLDivElement;

    onMount(() => {
        connectWebSocket();
        loadMessages();
    });

    onDestroy(() => {
        if (socket) {
            socket.close();
        }
    });

    function connectWebSocket() {
        socket = new WebSocket('ws://localhost:8080/ws');
        
        socket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'groupChat' && data.groupId === groupId) {
                messages = [...messages, data.message];
                scrollToBottom();
            }
        };
    }

    async function loadMessages() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/messages`, {
                credentials: 'include'
            });
            if (response.ok) {
                messages = await response.json();
                scrollToBottom();
            }
        } catch (error) {
            console.error('Failed to load messages:', error);
        }
    }

    function handleSend() {
        if (!newMessage.trim() || !socket) return;

        const message = {
            type: 'groupChat',
            groupId,
            content: newMessage,
            senderId: $auth.user?.id
        };

        socket.send(JSON.stringify(message));
        newMessage = '';
    }

    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            handleSend();
        }
    }

    function handleEmojiSelect(event: CustomEvent<EmojiPickerEvent>) {
        const emoji = event.detail.emoji.native;
        newMessage += emoji;
    }

    function scrollToBottom() {
        setTimeout(() => {
            if (chatContainer) {
                chatContainer.scrollTop = chatContainer.scrollHeight;
            }
        }, 0);
    }
</script>

<Card class="h-[calc(100vh-16rem)] flex flex-col">
    <div 
        class="flex-1 overflow-y-auto p-4 space-y-4"
        bind:this={chatContainer}
    >
        {#each messages as message}
            <div class="flex {message.senderId === $auth.user?.id ? 'justify-end' : 'justify-start'}">
                <div class="max-w-[70%] {message.senderId === $auth.user?.id ? 'bg-primary-100 dark:bg-primary-800' : 'bg-gray-100 dark:bg-gray-700'} rounded-lg p-3">
                    {#if message.senderId !== $auth.user?.id}
                        <p class="text-sm font-semibold mb-1">{message.senderName}</p>
                    {/if}
                    <p class="break-words">{message.content}</p>
                    <p class="text-xs text-gray-500 mt-1">
                        {getLastDate(message.createdAt)}
                    </p>
                </div>
            </div>
        {/each}
    </div>

    <div class="border-t dark:border-gray-700 p-4">
        <div class="flex space-x-2">
            <Input
                type="text"
                bind:value={newMessage}
                on:keypress={handleKeyPress}
                placeholder="Type a message..."
                class="flex-1"
            />
            <EmojiPicker on:emoji-select={handleEmojiSelect} />
            <Button on:click={handleSend}>Send</Button>
        </div>
    </div>
</Card> 