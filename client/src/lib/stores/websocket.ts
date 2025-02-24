// src/lib/services/websocket.ts
import { writable, derived, get } from 'svelte/store';
import type { Writable } from 'svelte/store';
import { auth } from '$lib/stores/auth';

// Message Types
export enum MessageType {
    CHAT = 'chat',
    GROUP_CHAT = 'groupChat',
    EVENT_RSVP = 'eventRSVP',
    TYPING = 'typing',
    NOTIFICATION = 'notification',
    FOLLOWER_REQUEST = 'followerRequest'
}

// Message Interfaces
export interface ChatMessage {
    type: MessageType.CHAT;
    id?: number;
    senderId: number;
    recipientId: number;
    content: string;
    createdAt: string;
    senderName?: string;
    senderAvatar?: string;
}

export interface GroupChatMessage {
    type: MessageType.GROUP_CHAT;
    id?: number;
    groupId: number;
    userId: number;
    content: string;
    media?: string;
    createdAt: string;
    userName?: string;
    userAvatar?: string;
}

export interface TypingIndicator {
    type: MessageType.TYPING;
    senderId: number;
    recipientId: number;
    isTyping: boolean;
}

export interface EventRSVPMessage {
    type: MessageType.EVENT_RSVP;
    groupId: number;
    eventId: number;
    status: 'going' | 'notGoing' | 'maybe';
    going: number;
    notGoing: number;
}

export interface NotificationMessage {
    type: MessageType.NOTIFICATION;
    id?: number;
    userId: number;
    content: string;
    createdAt: string;
    link?: string;
    isRead: boolean;
}

export interface FollowerRequestMessage {
    type: MessageType.FOLLOWER_REQUEST;
    followerId: number;
    followedId: number;
    status: 'pending' | 'accepted' | 'rejected';
    followerName?: string;
    followerAvatar?: string;
}

export type WebSocketMessage =
    | ChatMessage
    | GroupChatMessage
    | TypingIndicator
    | EventRSVPMessage
    | NotificationMessage
    | FollowerRequestMessage;

// WebSocket connection state
export enum ConnectionState {
    CONNECTING = 'connecting',
    OPEN = 'open',
    CLOSED = 'closed',
    ERROR = 'error'
}

// Store for connection state
export const connectionState: Writable<ConnectionState> = writable(ConnectionState.CLOSED);

// Store for all messages
export const messages: Writable<WebSocketMessage[]> = writable([]);

// Store for active chats
interface ActiveChat {
    id: number;
    name: string;
    avatar?: string;
    unreadCount: number;
    lastMessage?: string;
    lastMessageTime?: string;
    isGroup: boolean;
}

export const activeChats: Writable<ActiveChat[]> = writable([]);

// Store for notifications
export const notifications: Writable<NotificationMessage[]> = writable([]);
export const unreadNotificationsCount = derived(
    notifications,
    $notifications => $notifications.filter(n => !n.isRead).length
);

// WebSocket instance
let socket: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let heartbeatInterval: ReturnType<typeof setInterval> | null = null;
const MAX_RECONNECT_DELAY = 30000; // 30 seconds max
let reconnectDelay = 1000; // Start with 1 second

/**
 * Initialize WebSocket connection
 */
export function initializeWebSocket(): void {
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
        console.log('WebSocket connection already exists');
        return;
    }

    connectionState.set(ConnectionState.CONNECTING);

    const currentUserId = getCurrentUserId();
    if (!currentUserId) {
        console.error('Cannot initialize WebSocket: No user ID available');
        connectionState.set(ConnectionState.ERROR);
        return;
    }

    // Connect to your Go backend's WebSocket endpoint
    socket = new WebSocket('ws://localhost:8080/ws');

    socket.onopen = () => {
        console.log('WebSocket connection established');
        connectionState.set(ConnectionState.OPEN);
        reconnectDelay = 1000;
        startHeartbeat();
    };

    socket.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data) as WebSocketMessage;
            handleMessage(message);
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };

    socket.onclose = (event) => {
        console.log(`WebSocket connection closed: ${event.code} ${event.reason}`);
        connectionState.set(ConnectionState.CLOSED);
        cleanupConnection();
        scheduleReconnect();
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        connectionState.set(ConnectionState.ERROR);
    };
}
/**
 * Send a message through the WebSocket
 */
export function sendMessage(message: WebSocketMessage): boolean {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.error('Cannot send message: WebSocket not connected');
        return false;
    }

    try {
        socket.send(JSON.stringify(message));

        // If it's a chat message, add it to the messages store if it's not already present
        if (message.type === MessageType.CHAT || message.type === MessageType.GROUP_CHAT) {
            messages.update(msgs => {
                // Create a unique identifier for the message
                const messageId = `${message.type}-${message.createdAt}-${message.content}`;

                // Check if the message already exists in the store
                const existingMessageIndex = msgs.findIndex(msg => {
                    if (msg.type === MessageType.CHAT) {
                        return `${msg.type}-${msg.createdAt}-${msg.content}` === messageId;
                    } else if (msg.type === MessageType.GROUP_CHAT) {
                        return `${msg.type}-${msg.createdAt}-${msg.content}` === messageId;
                    }
                    return false;
                });

                if (existingMessageIndex === -1) {
                    return [...msgs, message];
                } else {
                    return msgs;
                }
            });

            updateActiveChat(message);
        }

        return true;
    } catch (error) {
        console.error('Error sending message:', error);
        return false;
    }
}

/**
 * Close the WebSocket connection
 */
export function closeConnection(): void {
    if (socket) {
        socket.close(1000, 'User initiated close');
        cleanupConnection();
    }
}

/**
 * Handle incoming messages based on type
 */
function handleMessage(message: WebSocketMessage): void {
    console.log('Current messages:', get(messages));
    messages.update(msgs => {
        // Create a unique identifier for the message
        let messageId: string;
        if (message.type === MessageType.CHAT || message.type === MessageType.GROUP_CHAT) {
            messageId = `${message.type}-${message.createdAt}-${message.content}`;
        } else {
            messageId = `${message.type}`;
        }

        // Check if the message already exists in the store
        const existingMessageIndex = msgs.findIndex(msg => {
            if (msg.type === MessageType.CHAT) {
                return `${msg.type}-${msg.createdAt}-${msg.content}` === messageId;
            } else if (msg.type === MessageType.GROUP_CHAT) {
                return `${msg.type}-${msg.createdAt}-${msg.content}` === messageId;
            } else if (msg.type === MessageType.TYPING) {
                return `${msg.type}` === messageId;
            }
            return false;
        });

        if (existingMessageIndex === -1) {
            return [...msgs, message];
        } else {
            return msgs;
        }
    });
    console.log('Updated messages:', get(messages));

    switch (message.type) {
        case MessageType.CHAT:
            handleChatMessage(message as ChatMessage);
            break;
        case MessageType.GROUP_CHAT:
            handleGroupChatMessage(message as GroupChatMessage);
            break;
        case MessageType.NOTIFICATION:
            handleNotification(message as NotificationMessage);
            break;
        case MessageType.EVENT_RSVP:
            // Handle RSVP updates
            break;
        case MessageType.FOLLOWER_REQUEST:
            handleFollowerRequest(message as FollowerRequestMessage);
            break;
        case MessageType.TYPING:
            // Handle typing indicators
            break;
    }
}
/**
 * Handle chat messages
 */
function handleChatMessage(message: ChatMessage): void {
    updateActiveChat(message);
}

/**
 * Handle group chat messages
 */
function handleGroupChatMessage(message: GroupChatMessage): void {
    updateActiveChat(message);
}

/**
 * Handle notifications
 */
function handleNotification(notification: NotificationMessage): void {
    notifications.update(notes => [notification, ...notes]);

    // Show browser notification if supported and user gave permission
    if (
        'Notification' in window &&
        Notification.permission === 'granted' &&
        document.visibilityState !== 'visible'
    ) {
        new Notification('Social Network', {
            body: notification.content,
            icon: '/favicon.png'
        });
    }
}

/**
 * Handle follower requests
 */
function handleFollowerRequest(request: FollowerRequestMessage): void {
    if (request.status === 'pending') {
        const notification: NotificationMessage = {
            type: MessageType.NOTIFICATION,
            userId: request.followedId,
            content: `${request.followerName} wants to follow you`,
            createdAt: new Date().toISOString(),
            link: `/profile/${request.followerId}`,
            isRead: false
        };

        notifications.update(notes => [notification, ...notes]);
    }
}

/**
 * Update active chats when new messages arrive
 */
function updateActiveChat(message: ChatMessage | GroupChatMessage): void {
    activeChats.update(chats => {
        const isGroupMessage = message.type === MessageType.GROUP_CHAT;
        const chatId = isGroupMessage
          ? (message as GroupChatMessage).groupId
          : getPrivateChatId(message as ChatMessage);

        // Find if chat already exists
        const existingChatIndex = chats.findIndex(c =>
          c.id === chatId && c.isGroup === isGroupMessage
        );

        if (existingChatIndex >= 0) {
            // Update existing chat
            const updatedChats = [...chats];
            updatedChats[existingChatIndex] = {
                ...updatedChats[existingChatIndex],
                lastMessage: message.content,
                lastMessageTime: message.createdAt,
                unreadCount: updatedChats[existingChatIndex].unreadCount + 1
            };
            return updatedChats;
        }

        // We need to fetch additional info for new chats
        if (isGroupMessage) {
            fetchGroupInfo((message as GroupChatMessage).groupId);
        } else {
            fetchUserInfo(getOtherUserId(message as ChatMessage));
        }

        return chats;
    });
}

/**
 * Get private chat ID from a message
 */
function getPrivateChatId(message: ChatMessage): number {
    // For direct messages, we use a convention where the chat ID is min(senderId, recipientId)_max(senderId, recipientId)
    const currentUserId = getCurrentUserId();
    if (!currentUserId) return 0;

    const otherUserId = message.senderId === currentUserId ? message.recipientId : message.senderId;
    return Math.min(currentUserId, otherUserId) * 1000000 + Math.max(currentUserId, otherUserId);
}

/**
 * Get the other user's ID from a chat message
 */
function getOtherUserId(message: ChatMessage): number {
    const currentUserId = getCurrentUserId();
    if (!currentUserId) return 0;

    return message.senderId === currentUserId ? message.recipientId : message.senderId;
}

/**
 * Get current user ID from auth store
 */
function getCurrentUserId(): number | null {
    const authState = get(auth);
    return authState.user?.id || null;
}

/**
 * Fetch group information
 */
async function fetchGroupInfo(groupId: number): Promise<void> {
    try {
        const response = await fetch(`http://localhost:8080/groups/${groupId}`, {
            credentials: 'include'
        });
        if (response.ok) {
            const groupData = await response.json();
            activeChats.update(chats => [
                ...chats,
                {
                    id: groupId,
                    name: groupData.name,
                    avatar: groupData.avatar,
                    unreadCount: 1,
                    isGroup: true
                }
            ]);
        }
    } catch (error) {
        console.error('Error fetching group info:', error);
    }
}

/**
 * Fetch user information
 */
async function fetchUserInfo(userId: number): Promise<void> {
    try {
        const response = await fetch(`http://localhost:8080/user/${userId}`, {
            credentials: 'include'
        });
        if (response.ok) {
            const userData = await response.json();
            const currentUserId = getCurrentUserId();
            if (!currentUserId) return;

            const chatId = getPrivateChatIdFromUserIds(currentUserId, userId);

            activeChats.update(chats => [
                ...chats,
                {
                    id: chatId,
                    name: `${userData.firstName} ${userData.lastName}`,
                    avatar: userData.avatar,
                    unreadCount: 1,
                    isGroup: false
                }
            ]);
        }
    } catch (error) {
        console.error('Error fetching user info:', error);
    }
}
/**
 * Get private chat ID from user IDs
 */
function getPrivateChatIdFromUserIds(userId1: number, userId2: number): number {
    return Math.min(userId1, userId2) * 1000000 + Math.max(userId1, userId2);
}

/**
 * Send a ping message to keep the connection alive
 */
function startHeartbeat(): void {
    heartbeatInterval = setInterval(() => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            // Send a ping message
            socket.send(JSON.stringify({ type: 'ping' }));
        }
    }, 30000); // Send a ping every 30 seconds
}

/**
 * Clean up resources
 */
function cleanupConnection(): void {
    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
    }
}

/**
 * Schedule reconnection with exponential backoff
 */
function scheduleReconnect(): void {
    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
    }

    reconnectTimer = setTimeout(() => {
        console.log(`Attempting to reconnect in ${reconnectDelay}ms...`);
        initializeWebSocket();
        // Exponential backoff with jitter
        reconnectDelay = Math.min(
            MAX_RECONNECT_DELAY,
            reconnectDelay * 1.5 + Math.random() * 1000
        );
    }, reconnectDelay);
}

/**
 * Mark all notifications as read
 */
export function markAllNotificationsAsRead(): void {
    notifications.update(notes =>
        notes.map(note => ({ ...note, isRead: true }))
    );
}

/**
 * Mark a specific notification as read
 */
export function markNotificationAsRead(notificationId: number): void {
    notifications.update(notes =>
        notes.map(note =>
            note.id === notificationId ? { ...note, isRead: true } : note
        )
    );
}

/**
 * Reset unread count for a specific chat
 */
export function resetUnreadCount(chatId: number, isGroup: boolean): void {
    activeChats.update(chats =>
        chats.map(chat =>
            chat.id === chatId && chat.isGroup === isGroup
                ? { ...chat, unreadCount: 0 }
                : chat
        )
    );
}

/**
 * Get messages for a specific chat
 */
export function getChatMessages(chatId: number, isGroup: boolean) {
    return derived(messages, $messages => {
        if (isGroup) {
            return $messages.filter(
                msg => msg.type === MessageType.GROUP_CHAT &&
                    (msg as GroupChatMessage).groupId === chatId
            );
        } else {
            // For private chats, we need to check both directions
            const currentUserId = getCurrentUserId();
            if (!currentUserId) return [];

            return $messages.filter(
                msg => msg.type === MessageType.CHAT &&
                    (
                        ((msg as ChatMessage).senderId === currentUserId &&
                            (msg as ChatMessage).recipientId === getRecipientIdFromChatId(chatId, currentUserId)) ||
                        ((msg as ChatMessage).recipientId === currentUserId &&
                            (msg as ChatMessage).senderId === getRecipientIdFromChatId(chatId, currentUserId))
                    )
            );
        }
    });
}

/**
 * Get recipient ID from a chat ID
 */
function getRecipientIdFromChatId(chatId: number, currentUserId: number): number {
    const id1 = Math.floor(chatId / 1000000);
    const id2 = chatId % 1000000;

    return id1 === currentUserId ? id2 : id1;
}

/**
 * Request notification permission
 */
export async function requestNotificationPermission(): Promise<boolean> {
    if (!('Notification' in window)) {
        console.log('This browser does not support notifications');
        return false;
    }

    if (Notification.permission === 'granted') {
        return true;
    }

    if (Notification.permission !== 'denied') {
        const permission = await Notification.requestPermission();
        return permission === 'granted';
    }

    return false;
}