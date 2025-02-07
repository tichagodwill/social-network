import {writable, get} from 'svelte/store';
import type {Message, User} from '$lib/types';

interface ChatState {
    messages: Message[];
    activeChat: number | null;
    contacts: User[];
    socket: WebSocket | null;
    isConnecting: boolean;
    typingUsers: Set<number>;
    unreadMessages: Map<number, number>;
}

interface ChatContact extends User {
    id: number;
    username: string;
    firstName: string;
    lastName: string;
    avatar?: string;
}

function createChatStore() {
    const {subscribe, set, update} = writable<ChatState>({
        messages: [],
        activeChat: null,
        contacts: [],
        socket: null,
        isConnecting: false,
        typingUsers: new Set(),
        unreadMessages: new Map()
    });

    let socket: WebSocket | null = null;
    let reconnectAttempts = 0;
    let reconnectTimeout: number | undefined;
    const MAX_RECONNECT_ATTEMPTS = 5;
    const TYPING_TIMEOUT = 3000;
    let typingTimeouts = new Map<number, number>();

    async function initializeWebSocket() {
        try {
            if (socket) {
                socket.close();
                socket = null;
            }

            socket = new WebSocket('ws://localhost:8080/ws');
            let currentState = get({subscribe});

            socket.onmessage = (event) => {
                const message = JSON.parse(event.data);

                switch (message.type) {
                    case 'chat':
                        handleIncomingMessage(message);
                        break;
                    case 'typing':
                        handleTypingIndicator(message);
                        break;
                    case 'read':
                        handleReadReceipt(message);
                        break;
                }
            };

            socket.onopen = () => {
                reconnectAttempts = 0;
                update(state => ({
                    ...state,
                    socket,
                    isConnecting: false
                }));
            };

            socket.onclose = () => {
                if (reconnectTimeout) {
                    window.clearTimeout(reconnectTimeout);
                }

                if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
                    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 10000);
                    reconnectTimeout = window.setTimeout(() => {
                        reconnectAttempts++;
                        initializeWebSocket();
                    }, delay);
                }
            };

        } catch (error) {
            console.error('WebSocket connection failed:', error);
            update(state => ({...state, isConnecting: false}));
        }
    }

    function handleIncomingMessage(message: Message) {
        update(state => {
            if (message.recipientId !== state.activeChat) {
                const count = state.unreadMessages.get(message.senderId) || 0;
                state.unreadMessages.set(message.senderId, count + 1);
                return state;
            }

            // Mark message as delivered and notify sender
            if (socket?.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({
                    type: 'read',
                    messageIds: [message.id],
                    senderId: message.senderId
                }));
            }

            return {
                ...state,
                messages: [...state.messages, message]
            };
        });
    }

    function handleTypingIndicator(data: { senderId: number; isTyping: boolean }) {
        update(state => {
            if (data.isTyping) {
                state.typingUsers.add(data.senderId);
            } else {
                state.typingUsers.delete(data.senderId);
            }
            return state;
        });
    }

    function handleReadReceipt(data: { messageIds: number[] }) {
        update(state => ({
            ...state,
            messages: state.messages.map(msg =>
                data.messageIds.includes(msg.id)
                    ? {...msg, status: 'read'}
                    : msg
            )
        }));
    }

    return {
        subscribe,

        initialize: async () => {
            update(state => ({...state, isConnecting: true}));
            await initializeWebSocket();
        },
        loadContacts: async (userId: number | string) => {
            try {
                const response = await fetch(
                    `http://localhost:8080/contact/${userId}`,
                    {
                        credentials: 'include',
                        headers: {
                            'Accept': 'application/json'
                        }
                    }
                );
                if (!response.ok) {
                    if (response.status === 404) {
                        update(state => ({
                            ...state,
                            contacts: []
                        }));
                        return [];
                    }
                    throw new Error(`Failed to load contacts: ${response.status}`);
                }

                const contacts = await response.json();

                update(state => ({
                    ...state,
                    contacts
                }));

                return contacts;
            } catch (error) {
                console.error('Failed to load contacts:', error);
                update(state => ({...state, contacts: []}));
                throw error;
            }
        },

        getOrCreateDirectChat: async (userId: number) => {
            try {
                const response = await fetch('http://localhost:8080/chat/check-follow', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Accept': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({userId})
                });

                if (response.ok) {
                    const data = await response.json();
                    return {chatId: data.id};
                }

                if (response.status === 403) {
                    return {error: 'To chat, either you need to follow this user or they need to follow you'};
                }

                const errorData = await response.json().catch(() => ({message: 'Unknown error'}));
                return {error: errorData.message || 'Failed to create chat'};
            } catch (error) {
                console.error('Failed to create/get direct chat:', error);
                return {error: error instanceof Error ? error.message : 'Failed to create chat'};
            }
        },
        loadMessages: async (userId: number, contactId: number) => {
            try {
                update(state => ({
                    ...state,
                    activeChat: contactId,
                    messages: []
                }));

                const response = await fetch(
                    `http://localhost:8080/messages/${userId}/${contactId}`,
                    {
                        credentials: 'include',
                        headers: {'Accept': 'application/json'}
                    }
                );

                if (!response.ok && response.status !== 404) {
                    throw new Error(`Failed to load messages: ${response.status}`);
                }

                let messages: Message[] = [];

                if (response.ok) {
                    messages = await response.json();
                }

                update(state => {
                    // Mark unread messages as delivered
                    const unreadMessages = messages
                        .filter(m => m.senderId === contactId && m.status === 'sent')
                        .map(m => m.id);

                    if (unreadMessages.length > 0 && socket?.readyState === WebSocket.OPEN) {
                        socket.send(JSON.stringify({
                            type: 'read',
                            messageIds: unreadMessages,
                            senderId: contactId
                        }));
                    }

                    return {
                        ...state,
                        messages,
                        unreadMessages: new Map(state.unreadMessages.set(contactId, 0))
                    };
                });
            } catch (error) {
                console.error('Failed to load messages:', error);
                throw error;
            }
        },

        sendMessage: async (content: string, senderId: number, recipientId: number, file?: File) => {
            try {
                const messageData: Partial<Message> = {
                    senderId,
                    recipientId,
                    content,
                    status: 'sent',
                    messageType: file ? 'file' : 'text',
                    createdAt: new Date().toISOString()
                };

                if (file) {
                    messageData.fileData = await file.arrayBuffer();
                    messageData.fileName = file.name;
                    messageData.fileType = file.type;
                }

                if (socket?.readyState === WebSocket.OPEN) {
                    socket.send(JSON.stringify({
                        type: 'chat',
                        ...messageData
                    }));
                } else {
                    throw new Error('WebSocket not connected');
                }
            } catch (error) {
                console.error('Failed to send message:', error);
                throw error;
            }
        },

        setTyping: (userId: number, recipientId: number, isTyping: boolean) => {
            if (socket?.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({
                    type: 'typing',
                    recipientId,
                    isTyping
                }));
            }

            const existingTimeout = typingTimeouts.get(userId);
            if (existingTimeout) {
                window.clearTimeout(existingTimeout);
            }

            if (isTyping) {
                const timeoutId = window.setTimeout(() => {
                    if (socket?.readyState === WebSocket.OPEN) {
                        socket.send(JSON.stringify({
                            type: 'typing',
                            recipientId,
                            isTyping: false
                        }));
                    }
                }, TYPING_TIMEOUT);

                typingTimeouts.set(userId, timeoutId);
            }
        },

        cleanup: () => {
            if (socket) {
                socket.close();
                socket = null;
            }
            if (reconnectTimeout) {
                window.clearTimeout(reconnectTimeout);
            }
            typingTimeouts.forEach(id => window.clearTimeout(id));
            typingTimeouts.clear();
            set({
                messages: [],
                activeChat: null,
                contacts: [],
                socket: null,
                isConnecting: false,
                typingUsers: new Set(),
                unreadMessages: new Map()
            });
        }
    };
}

export const chat = createChatStore();