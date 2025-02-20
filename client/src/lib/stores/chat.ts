// import { writable } from 'svelte/store';
// import type { Message, User } from '$lib/types';
//
// interface ChatState {
//     messages: Message[];
//     activeChat: number | null;
//     contacts: User[];
//     socket: WebSocket | null;
//     isConnecting: boolean;
// }
//
// interface SendMessageOptions {
//     url?: string;
//     fileName?: string;
//     fileType?: string;
// }
//
// function createChatStore() {
//     const { subscribe, set, update } = writable<ChatState>({
//         messages: [],
//         activeChat: null,
//         contacts: [],
//         socket: null,
//         isConnecting: false
//     });
//
//     let socket: WebSocket | null = null;
//     let reconnectAttempts = 0;
//     const MAX_RECONNECT_ATTEMPTS = 5;
//
//     function setupSocketListeners(currentState: ChatState) {
//         if (!socket) return;
//
//         socket.onmessage = (event) => {
//             const message = JSON.parse(event.data);
//
//             if (message.recipientId !== currentState.activeChat) {
//                 // TODO: Add notification handling
//                 return;
//             }
//
//             if (message.type === 'chat') {
//                 update(state => ({
//                     ...state,
//                     messages: [...state.messages, message]
//                 }));
//             }
//         };
//
//         socket.onclose = () => {
//             if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
//                 setTimeout(() => {
//                     reconnectAttempts++;
//                     initializeWebSocket();
//                 }, 1000 * Math.pow(2, reconnectAttempts));
//             }
//         };
//     }
//
//     async function initializeWebSocket() {
//         try {
//             if (socket) {
//                 socket.close();
//                 socket = null;
//             }
//
//             socket = new WebSocket('ws://localhost:8080/ws');
//             let currentState: ChatState = {
//                 messages: [],
//                 activeChat: null,
//                 contacts: [],
//                 socket: null,
//                 isConnecting: true
//             };
//
//             subscribe(state => {
//                 currentState = state;
//             });
//
//             setupSocketListeners(currentState);
//
//             socket.onopen = () => {
//                 reconnectAttempts = 0;
//                 update(state => ({
//                     ...state,
//                     socket,
//                     isConnecting: false
//                 }));
//             };
//
//             socket.onerror = () => {
//                 console.warn('WebSocket connection error - will retry');
//             };
//
//         } catch (error) {
//             console.error('WebSocket connection failed:', error);
//             update(state => ({ ...state, isConnecting: false }));
//         }
//     }
//
//     return {
//         subscribe,
//         initialize: async () => {
//             update(state => ({ ...state, isConnecting: true }));
//             // Start WebSocket connection in the background
//             initializeWebSocket();
//             // Don't wait for it to complete
//             update(state => ({ ...state, isConnecting: false }));
//         },
//
//         loadMessages: async (userId: number, contactId: number) => {
//             try {
//                 update(state => ({
//                     ...state,
//                     activeChat: contactId,
//                     messages: [] // Clear existing messages while loading
//                 }));
//
//                 const response = await fetch(
//                   `http://localhost:8080/messages/${userId}/${contactId}`,
//                   {
//                       credentials: 'include',
//                       headers: {
//                           'Accept': 'application/json'
//                       }
//                   }
//                 );
//
//                 if (!response.ok && response.status !== 404) {
//                     throw new Error(`Failed to load messages: ${response.status}`);
//                 }
//
//                 let messages: Message[] = [];
//
//                 if (response.ok) {
//                     const data = await response.json();
//                     console.log('Raw message data:', data);
//
//                     // Handle different response formats
//                     if (Array.isArray(data)) {
//                         messages = data;
//                     } else if (data && typeof data === 'object' && 'messages' in data) {
//                         messages = data.messages;
//                     } else if (data && typeof data === 'object') {
//                         // Convert single message to array if needed
//                         messages = [data];
//                     } else {
//                         console.warn('Unexpected message data format:', data);
//                     }
//                 }
//
//                 // Ensure all messages have required fields
//                 messages = messages.filter(msg => {
//                     if (!msg || typeof msg !== 'object') return false;
//                     if (!msg.createdAt) msg.createdAt = new Date().toISOString();
//                     return true;
//                 });
//
//                 // Sort messages by creation time
//                 messages.sort((a, b) => {
//                     const timeA = new Date(a.createdAt).getTime();
//                     const timeB = new Date(b.createdAt).getTime();
//                     return timeA - timeB;
//                 });
//
//                 update(state => ({
//                     ...state,
//                     messages
//                 }));
//             } catch (error) {
//                 console.error('Failed to load messages:', error);
//                 update(state => ({
//                     ...state,
//                     messages: []
//                 }));
//                 throw error;
//             }
//         },
//
//         sendMessage: (content: string, senderId: number, recipientId: number, options?: SendMessageOptions) => {
//             const message = {
//                 type: 'chat',
//                 content,
//                 senderId,
//                 recipientId,
//                 ...options,
//                 createdAt: new Date().toISOString()
//             };
//
//             // Optimistically add message to state
//             update(state => ({
//                 ...state,
//                 messages: [...state.messages, message].sort((a, b) => {
//                     const timeA = new Date(a.createdAt).getTime();
//                     const timeB = new Date(b.createdAt).getTime();
//                     return timeA - timeB;
//                 })
//             }));
//
//             // Send via WebSocket
//             if (socket && socket.readyState === WebSocket.OPEN) {
//                 socket.send(JSON.stringify(message));
//             } else {
//                 console.warn('WebSocket not connected - message may not be delivered');
//                 // Queue message to be sent when WebSocket reconnects
//                 socket?.addEventListener('open', () => {
//                     socket?.send(JSON.stringify(message));
//                 }, { once: true });
//             }
//         },
//
//         loadContacts: async (userId: string | number) => {
//             try {
//                 console.log('Loading contacts for user:', userId);
//                 const response = await fetch(
//                   `http://localhost:8080/contact/${userId}`,
//                   {
//                       credentials: 'include',
//                       headers: {
//                           'Accept': 'application/json'
//                       }
//                   }
//                 );
//
//                 if (!response.ok) {
//                     console.error('Failed to load contacts:', response.status, response.statusText);
//                     const text = await response.text();
//                     console.error('Error response:', text);
//                     throw new Error(`Failed to load contacts: ${response.status}`);
//                 }
//
//                 const rawContacts = await response.json();
//                 console.log('Raw contacts:', rawContacts);
//
//                 if (!Array.isArray(rawContacts)) {
//                     console.error('Expected array of contacts but got:', typeof rawContacts);
//                     throw new Error('Invalid contacts data');
//                 }
//
//                 const contacts = rawContacts.map(c => {
//                     console.log('Processing contact:', c);
//                     return {
//                         id: c.ID || c.id,
//                         username: c.Username || c.username,
//                         firstName: c.FirstName || c.first_name || '',
//                         lastName: c.LastName || c.last_name || '',
//                         avatar: c.Avatar || c.avatar || null,
//                         email: c.Email || c.email || '',
//                         bio: c.AboutMe || c.about_me || c.bio || ''
//                     };
//                 });
//
//                 console.log('Processed contacts:', contacts);
//
//                 update(state => ({
//                     ...state,
//                     contacts
//                 }));
//
//                 return contacts;
//             } catch (error) {
//                 console.error('Failed to load contacts:', error);
//                 update(state => ({ ...state, contacts: [] }));
//                 throw error;
//             }
//         },
//
//         getOrCreateDirectChat: async (userId: number) => {
//             try {
//                 const response = await fetch('http://localhost:8080/chat/direct', {
//                     method: 'POST',
//                     headers: {
//                         'Content-Type': 'application/json',
//                         'Accept': 'application/json'
//                     },
//                     credentials: 'include',
//                     body: JSON.stringify({ userId })
//                 });
//
//                 if (response.ok) {
//                     const data = await response.json();
//                     return { chatId: data.id };
//                 }
//
//                 if (response.status === 403) {
//                     return { error: 'To chat, either you need to follow this user or they need to follow you' };
//                 }
//
//                 const errorData = await response.json().catch(() => ({ message: 'Unknown error' }));
//                 return { error: errorData.message || 'Failed to create chat' };
//             } catch (error) {
//                 console.error('Failed to create/get direct chat:', error);
//                 return { error: error instanceof Error ? error.message : 'Failed to create chat' };
//             }
//         },
//
//         cleanup: () => {
//             if (socket) {
//                 socket.close();
//                 socket = null;
//             }
//             set({
//                 messages: [],
//                 activeChat: null,
//                 contacts: [],
//                 socket: null,
//                 isConnecting: false
//             });
//         }
//     };
// }
//
// export const chat = createChatStore();
