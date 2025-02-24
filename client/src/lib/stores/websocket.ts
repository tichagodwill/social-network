// src/lib/stores/websocket.ts
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
    FOLLOWER_REQUEST = 'followerRequest',
    // Add server notification types
    GROUP_INVITATION = 'group_invitation',
    JOIN_REQUEST = 'join_request',
    FOLLOW_REQUEST = 'follow_request'
}

// Base Message Interface
export interface BaseMessage {
    type: MessageType;
    id?: number;
    createdAt: string;
}

// Message Interfaces
export interface ChatMessage extends BaseMessage {
    type: MessageType.CHAT;
    senderId: number;
    recipientId: number;
    content: string;
    senderName?: string;
    senderAvatar?: string;
}

export interface GroupChatMessage extends BaseMessage {
    type: MessageType.GROUP_CHAT;
    groupId: number;
    userId: number;
    content: string;
    media?: string;
    userName?: string;
    userAvatar?: string;
}

export interface TypingIndicator extends BaseMessage {
    type: MessageType.TYPING;
    senderId: number;
    recipientId: number;
    isTyping: boolean;
}

export interface EventRSVPMessage extends BaseMessage {
    type: MessageType.EVENT_RSVP;
    groupId: number;
    eventId: number;
    status: 'going' | 'notGoing' | 'maybe';
    going: number;
    notGoing: number;
}

export interface NotificationMessage extends BaseMessage {
    type: MessageType;
    userId: number;
    content: string;
    link?: string;
    isRead: boolean;
    isProcessed?: boolean;
    // Additional fields for specific notification types
    groupId?: number;
    invitationId?: number;
    userRole?: string;
    requestId?: number;
    followerName?: string;
    followerAvatar?: string;
    followerId?: number;
    followedId?: number;
}

export interface FollowerRequestMessage extends BaseMessage {
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
        loadInitialNotifications().catch(error => {
            console.error('Failed to load initial notifications:', error);
        });
    };

    socket.onmessage = (event) => {
        try {
            let message;
            const data = JSON.parse(event.data);

            // Handle different server message formats
            if (data.type === 'notification' && data.data) {
                // Server sends notification in { type: 'notification', data: {...} } format
                message = normalizeNotification(data.data);
            } else {
                // Regular message format
                message = normalizeMessage(data);
            }

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
 * Normalize notification from server format to our internal format
 */
function normalizeNotification(notification: any): NotificationMessage {
    return {
        type: notification.type || MessageType.NOTIFICATION,
        id: notification.id,
        userId: notification.user_id || notification.userId,
        content: notification.content,
        createdAt: notification.created_at || notification.createdAt || new Date().toISOString(),
        isRead: notification.is_read || notification.isRead || false,
        isProcessed: notification.is_processed || notification.isProcessed || false,
        link: notification.link,
        groupId: notification.group_id || notification.groupId,
        invitationId: notification.invitation_id || notification.invitationId,
        userRole: notification.user_role || notification.userRole,
        requestId: notification.request_id || notification.requestId,
        followerName: notification.follower_name || notification.followerName,
        followerAvatar: notification.follower_avatar || notification.followerAvatar,
        followerId: notification.follower_id || notification.followerId,
        followedId: notification.followed_id || notification.followedId
    };
}

/**
 * Normalize message from server format to our internal format
 */
function normalizeMessage(message: any): WebSocketMessage {
    // Ensure we have a createdAt value for all messages
    const createdAt = message.created_at || message.createdAt || new Date().toISOString();

    // Handle different message types based on the type field
    switch (message.type) {
        case 'chat':
            return {
                type: MessageType.CHAT,
                id: message.id,
                senderId: message.sender_id || message.senderId,
                recipientId: message.recipient_id || message.recipientId,
                content: message.content,
                createdAt,
                senderName: message.sender_name || message.senderName,
                senderAvatar: message.sender_avatar || message.senderAvatar
            } as ChatMessage;

        case 'groupChat':
            return {
                type: MessageType.GROUP_CHAT,
                id: message.id,
                groupId: message.group_id || message.groupId,
                userId: message.user_id || message.userId,
                content: message.content,
                media: message.media,
                createdAt,
                userName: message.user_name || message.userName,
                userAvatar: message.user_avatar || message.userAvatar
            } as GroupChatMessage;

        case 'notification':
        case 'group_invitation':
        case 'join_request':
        case 'follow_request':
            return normalizeNotification(message);

        default:
            // Pass through as-is if no normalization needed
            // But ensure it has a createdAt value
            return {
                ...message,
                createdAt: message.createdAt || createdAt
            } as WebSocketMessage;
    }
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
                    if (msg.type === MessageType.CHAT && 'content' in msg) {
                        return `${msg.type}-${msg.createdAt}-${msg.content}` === messageId;
                    } else if (msg.type === MessageType.GROUP_CHAT && 'content' in msg) {
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

            updateActiveChat(message as ChatMessage | GroupChatMessage);
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
    console.log('Received message:', message);

    // Skip duplicate messages
    if (isMessageDuplicate(message)) {
        console.log('Skipping duplicate message');
        return;
    }

    switch (message.type) {
        case MessageType.CHAT:
            messages.update(msgs => [...msgs, message]);
            updateActiveChat(message as ChatMessage);
            break;
        case MessageType.GROUP_CHAT:
            messages.update(msgs => [...msgs, message]);
            updateActiveChat(message as GroupChatMessage);
            break;
        case MessageType.NOTIFICATION:
        case MessageType.GROUP_INVITATION:
        case MessageType.JOIN_REQUEST:
        case MessageType.FOLLOW_REQUEST:
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
 * Check if a message is a duplicate
 */
function isMessageDuplicate(message: WebSocketMessage): boolean {
    if ((message.type === MessageType.CHAT || message.type === MessageType.GROUP_CHAT) && message.id !== undefined) {
        const msgs = get(messages);
        return msgs.some(m =>
          m.type === message.type &&
          m.id === message.id &&
          m.id !== undefined
        );
    } else if ((
      message.type === MessageType.NOTIFICATION ||
      message.type === MessageType.GROUP_INVITATION ||
      message.type === MessageType.JOIN_REQUEST ||
      message.type === MessageType.FOLLOW_REQUEST
    ) && message.id !== undefined) {
        const notes = get(notifications);
        return notes.some(n => n.id === message.id && n.id !== undefined);
    }

    return false;
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
            type: MessageType.FOLLOW_REQUEST,
            userId: request.followedId,
            content: `${request.followerName} wants to follow you`,
            createdAt: new Date().toISOString(),
            link: `/profile/${request.followerId}`,
            isRead: false,
            followerId: request.followerId,
            followedId: request.followedId,
            followerName: request.followerName,
            followerAvatar: request.followerAvatar
        };

        notifications.update(notes => [notification, ...notes]);
    }
}

/**
 * Update active chats when new messages arrive
 */
function updateActiveChat(message: ChatMessage | GroupChatMessage): void {
    if (!('content' in message)) return;

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

export async function loadInitialNotifications(): Promise<void> {
    try {
        const response = await fetch('http://localhost:8080/notifications', {
            credentials: 'include'
        });

        if (response.ok) {
            const data = await response.json() || [];

            if (Array.isArray(data)) {
                // Convert server notifications to our format and update store
                notifications.set(data.map((n: any) => normalizeNotification(n)));
            } else {
                console.error('Expected array of notifications but got:', data);
                notifications.set([]); // Set empty array as fallback
            }
        } else {
            console.error('Failed to load initial notifications:', response.status);
        }
    } catch (error) {
        console.error('Error loading initial notifications:', error);
        notifications.set([]); // Set empty array on error
    }
}
/**
 * Mark all notifications as read
 */
export async function markAllNotificationsAsRead(): Promise<void> {
    try {
        const response = await fetch('http://localhost:8080/notifications/read-all', {
            method: 'POST',
            credentials: 'include'
        });

        if (response.ok) {
            notifications.update(notes =>
              notes.map(note => ({ ...note, isRead: true }))
            );
        } else {
            console.error('Failed to mark all notifications as read:', response.status);
        }
    } catch (error) {
        console.error('Error marking all notifications as read:', error);
        throw error;
    }
}

/**
 * Mark a specific notification as read
 */
export async function markNotificationAsRead(notificationId: number): Promise<void> {
    try {
        const response = await fetch(`http://localhost:8080/notifications/${notificationId}/read`, {
            method: 'GET',
            credentials: 'include'
        });

        if (response.ok) {
            notifications.update(notes =>
              notes.map(note =>
                note.id === notificationId ? { ...note, isRead: true } : note
              )
            );
        } else {
            throw new Error(`Failed to mark notification as read: ${response.status}`);
        }
    } catch (error) {
        console.error('Error marking notification as read:', error);
        throw error;
    }
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
                'groupId' in msg &&
                msg.groupId === chatId
            );
        } else {
            // For private chats, we need to check both directions
            const currentUserId = getCurrentUserId();
            if (!currentUserId) return [];

            return $messages.filter(
              msg => msg.type === MessageType.CHAT &&
                'senderId' in msg && 'recipientId' in msg &&
                (
                  (msg.senderId === currentUserId &&
                    msg.recipientId === getRecipientIdFromChatId(chatId, currentUserId)) ||
                  (msg.recipientId === currentUserId &&
                    msg.senderId === getRecipientIdFromChatId(chatId, currentUserId))
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