import { writable } from 'svelte/store';
import type { Message, User } from '$lib/types';
import { transformUser } from '$lib/utils/transformer';

interface ChatState {
    messages: Message[];
    activeChat: number | null;
    contacts: User[];
    socket: WebSocket | null;
    isConnecting: boolean;
}

interface SendMessageOptions {
    url?: string;
    fileName?: string;
    fileType?: string;
}

function createChatStore() {
    const { subscribe, set, update } = writable<ChatState>({
        messages: [],
        activeChat: null,
        contacts: [],
        socket: null,
        isConnecting: false
    });

    let socket: WebSocket | null = null;
    let reconnectAttempts = 0;
    const MAX_RECONNECT_ATTEMPTS = 5;

    function setupSocketListeners(currentState: ChatState) {
        if (!socket) return;

        socket.onmessage = (event) => {
            const message = JSON.parse(event.data);

            if (message.recipientId !== currentState.activeChat) {
                // TODO: Add notification handling
                return;
            }

            if (message.type === 'chat') {
                update(state => ({
                    ...state,
                    messages: [...state.messages, message]
                }));
            }
        };

        socket.onclose = () => {
            if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
                setTimeout(() => {
                    reconnectAttempts++;
                    initialize();
                }, 1000 * Math.pow(2, reconnectAttempts));
            }
        };
    }

    async function initialize() {
        update(state => ({ ...state, isConnecting: true }));

        try {
            socket = new WebSocket('ws://localhost:8080/ws');
            let currentState: ChatState = {
                messages: [],
                activeChat: null,
                contacts: [],
                socket: null,
                isConnecting: true
            };

            subscribe(state => {
                currentState = state;
            });

            setupSocketListeners(currentState);

            socket.onopen = () => {
                reconnectAttempts = 0;
                update(state => ({
                    ...state,
                    socket,
                    isConnecting: false
                }));
            };

        } catch (error) {
            console.error('WebSocket connection failed:', error);
            update(state => ({ ...state, isConnecting: false }));
        }
    }

    return {
        subscribe,
        initialize,
        loadMessages: async (userId: number, contactId: number) => {
            try {
                const response = await fetch(
                  `http://localhost:8080/messages/${userId}/${contactId}`,
                  { credentials: 'include' }
                );

                if (response.ok) {
                    const messages = await response.json() ?? [];
                    update(state => ({
                        ...state,
                        messages,
                        activeChat: contactId
                    }));
                } else if (response.status === 404) {
                    // No messages yet, but that's ok
                    update(state => ({
                        ...state,
                        messages: [],
                        activeChat: contactId
                    }));
                } else {
                    throw new Error(`Failed to load messages: ${response.status}`);
                }
            } catch (error) {
                console.error('Failed to load messages:', error);
                update(state => ({
                    ...state,
                    messages: [],
                    activeChat: contactId
                }));
            }
        },

        sendMessage: (content: string, senderId: number, recipientId: number, options?: SendMessageOptions) => {
            if (!socket || !recipientId) return;

            const message = {
                type: 'chat',
                content,
                senderId,
                recipientId,
                ...options,
                createdAt: new Date()
            };

            socket.send(JSON.stringify(message));
        },

        getOrCreateDirectChat: async (userId: number) => {
            try {
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
                    return { chatId: data.id };
                }
                
                if (response.status === 403) {
                    return { error: 'To chat, either you need to follow this user or they need to follow you' };
                }
                
                return { error: 'Failed to create chat' };
            } catch (error) {
                console.error('Failed to create/get direct chat:', error);
                return { error: 'Failed to create chat' };
            }
        },

        loadContacts: async (userId: string | number) => {
            try {
                const response = await fetch(
                  `http://localhost:8080/contact/${userId}`,
                  { credentials: 'include' }
                );

                if (response.ok) {
                    const contacts = (await response.json()).map((c: any) => transformUser(c));
                    update(state => ({ ...state, contacts }));
                } else {
                    throw new Error(`Failed to load contacts: ${response.status}`);
                }
            } catch (error) {
                console.error('Failed to load contacts:', error);
                update(state => ({ ...state, contacts: [] }));
            }
        },

        cleanup: () => {
            if (socket) {
                socket.close();
                socket = null;
            }
            set({
                messages: [],
                activeChat: null,
                contacts: [],
                socket: null,
                isConnecting: false
            });
        }
    };
}

export const chat = createChatStore();
