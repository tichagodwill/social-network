// src/lib/stores/websocket.ts
import {writable, derived, get} from 'svelte/store';
import type {Writable} from 'svelte/store';
import {auth} from '$lib/stores/auth';

// Message Types
export enum MessageType {
    CHAT = 'chat',
    GROUP_CHAT = 'groupChat',
    EVENT_RSVP = 'eventRSVP',
    TYPING = 'typing',
    NOTIFICATION = 'notification',
    FOLLOWER_REQUEST = 'followerRequest',
    ERROR = 'error', // Add error message type
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
    chatId?: number;
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
    chatId?: number;
}

export interface FollowerRequestMessage extends BaseMessage {
    type: MessageType.FOLLOWER_REQUEST;
    followerId: number;
    followedId: number;
    status: 'pending' | 'accepted' | 'rejected';
    followerName?: string;
    followerAvatar?: string;
}

export interface ErrorMessage extends BaseMessage {
    type: MessageType.ERROR;
    message: string;
    code: string;
}

export type WebSocketMessage =
    | ChatMessage
    | GroupChatMessage
    | TypingIndicator
    | EventRSVPMessage
    | NotificationMessage
    | FollowerRequestMessage
    | ErrorMessage;

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

// Store for message tracking to prevent duplicates
export const processedMessageIds: Writable<Set<string>> = writable(new Set());

// Store for active chats
interface ActiveChat {
    id: number;
    name: string;
    avatar?: string;
    unreadCount: number;
    lastMessage?: string;
    lastMessageTime?: string;
    isGroup: boolean;
    recipientId?: number;
}

export const activeChats: Writable<ActiveChat[]> = writable([]);

// Store for notifications
export const notifications: Writable<NotificationMessage[]> = writable([]);

// Create a writable store for unread count that we'll keep in sync
export const unreadNotificationsCount: Writable<number> = writable(0);

// Create a derived store to automatically calculate unread count
const derivedUnreadCount = derived(
    notifications,
    $notifications => $notifications.filter(n => !n.isRead).length
);

// Subscribe to the derived store to keep the writable store in sync
derivedUnreadCount.subscribe((count) => {
    unreadNotificationsCount.set(count);
});

// WebSocket instance
let wsInstance: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let heartbeatInterval: ReturnType<typeof setInterval> | null = null;
const MAX_RECONNECT_DELAY = 30000; // 30 seconds max
const INITIAL_RECONNECT_DELAY = 1000;
let reconnectDelay = INITIAL_RECONNECT_DELAY;
let isReconnecting = false;

// Constants for connection management
const CONNECTION_TIMEOUT = 5000;
const RECONNECT_DELAY = 2000;
const MAX_RECONNECT_ATTEMPTS = 5;
let reconnectAttempts = 0;

/**
 * Initialize WebSocket connection
 */
export function initializeWebSocket(): void {
    let currentState: ConnectionState;
    const unsubscribe = connectionState.subscribe(state => {
        currentState = state;
    });
    unsubscribe();

    if (currentState === ConnectionState.CONNECTING) {
        console.log('Already attempting to connect...');
        return;
    }

    // Reset reconnect attempts if this is a fresh connection
    if (currentState === ConnectionState.CLOSED) {
        reconnectAttempts = 0;
    }

    if (wsInstance?.readyState === WebSocket.OPEN) {
        console.log('WebSocket connection already exists and is open');
        return;
    }

    connectionState.set(ConnectionState.CONNECTING);
    cleanupConnection();

    const connectWithRetry = () => {
        try {
            wsInstance = new WebSocket('ws://localhost:8080/ws');
            wsInstance.binaryType = 'arraybuffer';

            let isConnected = false;
            let connectionTimeout: NodeJS.Timeout;

            // Set connection timeout
            connectionTimeout = setTimeout(() => {
                if (!isConnected) {
                    console.log('Connection attempt timed out');
                    wsInstance?.close();
                    connectionState.set(ConnectionState.ERROR);
                    scheduleReconnect();
                }
            }, CONNECTION_TIMEOUT);

            wsInstance.onopen = (event) => {
                console.log('WebSocket connection established');
                isConnected = true;
                connectionState.set(ConnectionState.OPEN);
                reconnectAttempts = 0; // Reset attempts on successful connection

                if (connectionTimeout) {
                    clearTimeout(connectionTimeout);
                }

                if (heartbeatInterval) {
                    clearInterval(heartbeatInterval);
                }

                // Start heartbeat
                startHeartbeat();

                // Fetch initial notifications
                loadInitialNotifications();

                // Fetch active chats - this ensures we get the most up-to-date chat list
                fetchActiveChats();
            };

            wsInstance.onmessage = (event) => {
                try {
                    let message;
                    const data = JSON.parse(event.data);
                    console.log("Received WebSocket message:", JSON.stringify(data, null, 2));

                    // Handle different server message formats
                    if (data.type === 'notification' && data.data) {
                        message = normalizeNotification(data.data);

                        // Update notifications store immediately
                        notifications.update(notes => {
                            return [message, ...notes];
                        });

                        // The unread count will be automatically updated via the derived store

                        // Show toast for event notifications
                        if (message.type === 'group_event') {
                            showToast(message.content, 'info');
                        }
                    } else {
                        message = normalizeMessage(data);
                    }

                    handleMessage(message);
                } catch (error) {
                    console.error('Error handling WebSocket message:', error);
                    console.error('Message data:', event.data);
                }
            };

            wsInstance.onclose = (event) => {
                const reason = event.reason || 'Unknown reason';
                console.log(`WebSocket connection closed: ${event.code} (${reason})`);
                isConnected = false;

                // Only change state if it wasn't intentionally closed
                if (event.code !== 1000) {
                    connectionState.set(ConnectionState.CLOSED);
                    scheduleReconnect();
                }

                if (heartbeatInterval) {
                    clearInterval(heartbeatInterval);
                    heartbeatInterval = null;
                }

                if (connectionTimeout) {
                    clearTimeout(connectionTimeout);
                }

                wsInstance = null;
                cleanupConnection();
            };

            wsInstance.onerror = (error) => {
                console.error('WebSocket error occurred:', {
                    error,
                    readyState: wsInstance?.readyState,
                    url: wsInstance?.url
                });
                connectionState.set(ConnectionState.ERROR);
                reconnectAttempts++;
                scheduleReconnect();
            };

        } catch (error) {
            console.error('Error initializing WebSocket:', error);
            connectionState.set(ConnectionState.ERROR);
            reconnectAttempts++;
            scheduleReconnect();
        }
    };

    connectWithRetry();
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
        followedId: notification.followed_id || notification.followedId,
        chatId: notification.chatId
    };
}

/**
 * Normalize message from server format to our internal format
 */
function normalizeMessage(message: any): WebSocketMessage {
    if (!message || !message.type) {
        console.error('Invalid message format:', message);
        return {
            type: MessageType.ERROR,
            message: 'Invalid message format',
            code: 'invalid_format',
            createdAt: new Date().toISOString()
        };
    }

    // Handle error messages
    if (message.type === 'error') {
        return {
            type: MessageType.ERROR,
            message: message.data?.message || 'An error occurred',
            code: message.data?.code || 'unknown_error',
            createdAt: new Date().toISOString()
        };
    }

    // Check if the actual message data is in the 'data' property
    const messageData = message.data || message;

    // Ensure we have a createdAt value for all messages
    const createdAt = messageData.created_at || messageData.createdAt || new Date().toISOString();

    // Handle chat messages
    if (message.type === 'chat' || message.type === MessageType.CHAT) {
        // Get the current user ID to handle missing recipients or senders
        const currentUserId = getCurrentUserId();

        // Extract fields, using either snake_case or camelCase versions, with fallbacks
        const id = messageData.id || message.id || Date.now();
        let chatId = messageData.chat_id || messageData.chatId || message.chat_id || message.chatId;
        const senderId = messageData.sender_id || messageData.senderId || message.sender_id || message.senderId || currentUserId;
        let recipientId = messageData.recipient_id || messageData.recipientId || message.recipient_id || message.recipientId;
        const content = messageData.content || message.content || '';
        const senderName = messageData.sender_name || messageData.senderName || message.sender_name || message.senderName;
        const senderAvatar = messageData.sender_avatar || messageData.senderAvatar || message.sender_avatar || message.senderAvatar;

        // If we still don't have chat ID but have both sender and recipient, check if we have a chat cache
        if (!chatId && senderId && recipientId) {
            console.log('Missing chatId, attempting to retrieve from active chats');
            // Look in active chats for a match between these two users
            const chats = get(activeChats);
            const foundChat = chats.find(chat => 
                !chat.isGroup && 
                (chat.recipientId === recipientId || chat.recipientId === senderId));
            
            if (foundChat) {
                chatId = foundChat.id;
                console.log('Retrieved chatId from active chats:', chatId);
            } else {
                // Fallback - create a temporary chat ID from the two user IDs
                // This should be replaced when the actual chat is created
                chatId = id; // Use message ID as temporary chat ID
                console.log('Created temporary chatId:', chatId);
            }
        }

        // If still missing recipient ID and we're the sender, try to get it from active chats
        if (!recipientId && senderId === currentUserId) {
            const chats = get(activeChats);
            const foundChat = chats.find(chat => chat.id === chatId && !chat.isGroup);
            if (foundChat && foundChat.recipientId) {
                recipientId = foundChat.recipientId;
            }
        }

        // Create the normalized chat message
        const chatMessage: ChatMessage = {
            type: MessageType.CHAT,
            id,
            chatId,
            senderId,
            recipientId,
            content,
            createdAt,
            senderName,
            senderAvatar
        };

        console.log('Normalized chat message:', chatMessage);
        return chatMessage;
    }

    // Handle different message types based on the type field
    switch (message.type) {
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
    if (!wsInstance || wsInstance.readyState !== WebSocket.OPEN) {
        console.error('Cannot send message: WebSocket not connected');
        return false;
    }

    // For chat messages, we MUST have a chatId from the server
    if (message.type === MessageType.CHAT && (!('chatId' in message) || !message.chatId)) {
        console.error('Cannot send message without a valid chatId:', message);
        return false;
    }

    try {
        // Wrap the message in the format expected by the server
        const serverMessage = {
            type: message.type,
            data: message
        };

        console.log("Sending message with chatId:", (message as any).chatId);

        // Send the message through WebSocket
        wsInstance.send(JSON.stringify(serverMessage));

        // We won't add the message to messages store here anymore
        // The server will echo back our own messages
        // This prevents duplicate messages showing up

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
    if (wsInstance) {
        wsInstance.close();
        wsInstance = null;
    }
}

/**
 * Handle incoming messages based on type
 */
function handleMessage(message: WebSocketMessage): void {
    if (!message || !message.type) {
        console.error('Invalid message format:', message);
        return;
    }

    // Check for duplicate messages
    if (isMessageDuplicate(message)) {
        return;
    }

    console.log('Processing message:', message);

    // Handle different message types
    switch (message.type) {
        case MessageType.CHAT:
            handleChatMessage(message as ChatMessage);
            break;

        case MessageType.GROUP_CHAT:
            handleGroupChatMessage(message as GroupChatMessage);
            break;

        case MessageType.NOTIFICATION:
        case MessageType.GROUP_INVITATION:
        case MessageType.JOIN_REQUEST:
        case MessageType.FOLLOW_REQUEST:
            handleNotification(message as NotificationMessage);
            break;

        case MessageType.FOLLOWER_REQUEST:
            handleFollowerRequest(message as FollowerRequestMessage);
            break;
            
        case MessageType.ERROR:
            handleErrorMessage(message as ErrorMessage);
            break;

        default:
            console.log('Unhandled message type:', message.type);
    }
}

/**
 * Handle chat messages
 */
function handleChatMessage(message: ChatMessage): void {
    console.log('Processing chat message:', message);
    
    // Update the read status of the message if it's from the current user
    const currentUserId = getCurrentUserId();
    const isFromCurrentUser = message.senderId === currentUserId;
    
    // Add to messages
    messages.update(msgs => [...msgs, message]);
    
    // Update active chats
    updateActiveChat(message);
    
    // If the message is from someone else, and we're not currently viewing this chat,
    // add a notification for it
    if (!isFromCurrentUser) {
        const activeChatId = get(currentChatId);
        // Only create notification if not currently viewing this chat
        if (activeChatId !== message.chatId) {
            // Format notification data
            const notification: NotificationMessage = {
                id: message.id,
                type: MessageType.CHAT,
                userId: message.recipientId,
                fromUserId: message.senderId,
                content: `${message.senderName || 'Someone'} sent you a message: ${message.content.substring(0, 30)}${message.content.length > 30 ? '...' : ''}`,
                createdAt: message.createdAt,
                link: `/chat/${message.chatId}`,
                isRead: false,
                chatId: message.chatId
            };
            
            // Add to notifications
            addNotification(notification);
        } else {
            // Mark chat messages as read if currently viewing
            markChatAsRead(message.chatId);
        }
    }
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
 * Handle error messages
 */
function handleErrorMessage(message: ErrorMessage): void {
    console.error('Error message received:', message);
    // Show toast notification for the error
    showToast(message.message, 'error');
}

/**
 * Helper functions for message types
 */
export function isGroupChatMessage(message: any): message is GroupChatMessage {
    return message.type === MessageType.GROUP_CHAT;
}

/**
 * Update active chats when new messages arrive
 */
function updateActiveChat(message: ChatMessage | GroupChatMessage): void {
    if (!('content' in message)) return;

    // Get current user ID to check if the message is from the current user
    const currentUserId = getCurrentUserId();
    const isFromCurrentUser = isGroupChatMessage(message)
        ? message.userId === currentUserId
        : message.senderId === currentUserId;

    activeChats.update(chats => {
        const isGroupMessage = message.type === MessageType.GROUP_CHAT;
        const chatId = isGroupMessage
            ? (message as GroupChatMessage).groupId
            : (message as ChatMessage).chatId;

        if (!chatId) {
            console.error('Message is missing required chatId or groupId:', message);
            return chats;
        }

        // First check if we already have a chat with this ID
        let existingChatIndex = chats.findIndex(c =>
            c.id === chatId && c.isGroup === isGroupMessage
        );

        // If found, update the existing chat
        if (existingChatIndex >= 0) {
            const updatedChats = [...chats];

            // Only increment unread count if message is NOT from current user
            const newUnreadCount = isFromCurrentUser
                ? updatedChats[existingChatIndex].unreadCount
                : updatedChats[existingChatIndex].unreadCount + 1;

            updatedChats[existingChatIndex] = {
                ...updatedChats[existingChatIndex],
                lastMessage: message.content,
                lastMessageTime: message.createdAt,
                unreadCount: newUnreadCount
            };
            return updatedChats;
        }

        // For direct chats, try to match by participant IDs if chatId doesn't match
        if (!isGroupMessage) {
            const chatMessage = message as ChatMessage;
            const otherUserId = chatMessage.senderId === currentUserId 
                ? chatMessage.recipientId 
                : chatMessage.senderId;
            
            // Find chats with this user as a participant
            existingChatIndex = chats.findIndex(c => 
                !c.isGroup && 
                (c.recipientId === otherUserId));
            
            if (existingChatIndex >= 0) {
                // Update chat with the correct chatId
                const updatedChats = [...chats];
                const newUnreadCount = isFromCurrentUser
                    ? updatedChats[existingChatIndex].unreadCount
                    : updatedChats[existingChatIndex].unreadCount + 1;
                
                updatedChats[existingChatIndex] = {
                    ...updatedChats[existingChatIndex],
                    id: chatId, // Update with the correct chatId
                    lastMessage: message.content,
                    lastMessageTime: message.createdAt,
                    unreadCount: newUnreadCount
                };
                return updatedChats;
            }
        }

        // If not found, fetch info for new chat
        if (isGroupMessage) {
            fetchGroupInfo((message as GroupChatMessage).groupId);
        } else {
            const otherUserId = getOtherUserId(message as ChatMessage);
            if (otherUserId > 0) {
                // We need to make an API call to get the chat ID
                fetchChatInfo(otherUserId);
            }
        }

        return chats;
    });
}

/**
 * Get the other user's ID from a chat message
 */
function getOtherUserId(message: ChatMessage): number {
    const currentUserId = getCurrentUserId();
    if (!currentUserId) return 0;

    // Safety check to ensure we have valid user IDs
    if (message.senderId === undefined || message.recipientId === undefined) {
        console.error('Invalid message: missing sender or recipient ID', message);
        return 0;
    }

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
            activeChats.update(chats => {
                // Check if group already exists
                const existingChat = chats.some(chat =>
                    chat.id === groupId && chat.isGroup === true
                );
                
                // Only add if doesn't exist
                if (!existingChat) {
                    return [...chats, {
                        id: groupId,
                        name: groupData.name,
                        avatar: groupData.avatar,
                        unreadCount: 1,
                        isGroup: true
                    }];
                }
                return chats;
            });
        }
    } catch (error) {
        console.error('Error fetching group info:', error);
    }
}

/**
 * Fetch chat information for a user
 */
async function fetchChatInfo(userId: number): Promise<void> {
    try {
        const currentUserId = getCurrentUserId();
        if (!currentUserId) return;

        // Get or create a chat with this user
        const response = await fetch(`http://localhost:8080/chats/user/${userId}`, {
            method: 'GET',
            credentials: 'include'
        });

        if (response.ok) {
            const chatData = await response.json();

            // Now fetch user details to get name and avatar
            const userResponse = await fetch(`http://localhost:8080/user/${userId}`, {
                credentials: 'include'
            });

            if (userResponse.ok) {
                const userData = await userResponse.json();

                // Get user name and avatar
                const firstName = userData.user?.first_name || userData.first_name || "Unknown";
                const lastName = userData.user?.last_name || userData.last_name || "User";
                const avatar = userData.user?.avatar || userData.avatar;

                // Update active chats
                activeChats.update(chats => {
                    // Check if chat already exists
                    const existingChat = chats.some(chat =>
                        chat.id === chatData.id && !chat.isGroup
                    );

                    // Only add if doesn't exist
                    if (!existingChat) {
                        return [...chats, {
                            id: chatData.id,
                            name: `${firstName} ${lastName}`,
                            avatar: avatar,
                            unreadCount: 1,
                            isGroup: false,
                            recipientId: userId,
                            lastMessage: "",
                            lastMessageTime: new Date().toISOString()
                        }];
                    }
                    return chats;
                });
            }
        }
    } catch (error) {
        console.error('Error fetching chat info:', error);
    }
}

/**
 * Send a ping message to keep the connection alive
 */
function startHeartbeat(): void {
    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
    }

    heartbeatInterval = setInterval(() => {
        if (wsInstance?.readyState === WebSocket.OPEN) {
            try {
                // Add timestamp to help track latency
                wsInstance.send(JSON.stringify({
                    type: 'ping',
                    timestamp: Date.now()
                }));
            } catch (error) {
                console.error('Error sending heartbeat:', error);
                cleanupConnection();
                scheduleReconnect();
            }
        }
    }, 20000); // 20 seconds interval
}

/**
 * Clean up resources
 */
function cleanupConnection(): void {
    if (wsInstance) {
        // Only close if not already closing/closed
        if (wsInstance.readyState === WebSocket.OPEN) {
            wsInstance.close(1000, 'Normal closure');
        }
        wsInstance = null;
    }

    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
    }

    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
    }
}

/**
 * Schedule reconnection with exponential backoff
 */
function scheduleReconnect(): void {
    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
    }

    // Stop reconnecting after max attempts
    if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        console.log('Max reconnection attempts reached');
        connectionState.set(ConnectionState.ERROR);
        return;
    }

    reconnectTimer = setTimeout(() => {
        console.log(`Reconnect attempt ${reconnectAttempts + 1} of ${MAX_RECONNECT_ATTEMPTS}`);
        initializeWebSocket();
    }, RECONNECT_DELAY);
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
 * Fetch active chats
 */
export async function fetchActiveChats(): Promise<void> {
    try {
        const response = await fetch('http://localhost:8080/chats', {
            credentials: 'include'
        });
        
        if (response.ok) {
            const chats = await response.json();
            
            // Process chats and update the store
            // Create a Set to track unique user IDs for deduplication
            const processedUserIds = new Set<number>();
            const processedChats = [];
            
            // First process all existing chats (non-potential)
            for (const chat of chats) {
                if (!chat.potential) {
                    processedChats.push({
                        id: chat.id,
                        name: chat.name || `${chat.first_name} ${chat.last_name}`,
                        avatar: chat.avatar,
                        unreadCount: chat.unread_count || 0,
                        isGroup: chat.type === 'group',
                        lastMessage: chat.last_message,
                        lastMessageTime: chat.last_message_time,
                        recipientId: chat.participant_id,
                        potential: false
                    });
                    
                    // Track user IDs for direct chats to avoid duplicates
                    if (chat.type !== 'group' && chat.participant_id) {
                        processedUserIds.add(chat.participant_id);
                    }
                }
            }
            
            // Then process potential chats, but only if we don't already have a chat with that user
            for (const chat of chats) {
                if (chat.potential && !processedUserIds.has(chat.participant_id)) {
                    processedChats.push({
                        id: chat.id,
                        name: chat.name || `${chat.first_name} ${chat.last_name}`,
                        avatar: chat.avatar,
                        unreadCount: chat.unread_count || 0,
                        isGroup: chat.type === 'group',
                        lastMessage: chat.last_message,
                        lastMessageTime: chat.last_message_time,
                        recipientId: chat.participant_id,
                        potential: true
                    });
                    processedUserIds.add(chat.participant_id);
                }
            }
            
            activeChats.set(processedChats);
        }
    } catch (error) {
        console.error('Error fetching active chats:', error);
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
                notes.map(note => ({...note, isRead: true}))
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
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        let errorMessage = 'Failed to mark notification as read';

        if (!response.ok) {
            if (response.headers.get('content-type')?.includes('application/json')) {
                const errorData = await response.json();
                errorMessage = errorData.error || `Failed to mark notification as read: ${response.status}`;
            } else {
                errorMessage = `Failed to mark notification as read: ${response.statusText}`;
            }
            throw new Error(errorMessage);
        }

        // Update the notifications store with the new state
        notifications.update(notifications =>
            notifications.map(notification =>
                notification.id === notificationId
                    ? {...notification, isRead: true}
                    : notification
            )
        );

        // The unread count will be automatically updated via the derived store

    } catch (error) {
        console.error('Error marking notification as read:', error);
        throw error;
    }
}

/**
 * Mark a chat as read
 */
export async function markChatAsRead(chatId: number): Promise<void> {
    try {
        // Send request to mark chat as read
        const response = await fetch(`http://localhost:8080/chats/${chatId}/read`, {
            method: 'POST',
            credentials: 'include'
        });
        
        if (response.ok) {
            // Update local state
            activeChats.update(chats => {
                return chats.map(chat => {
                    if (chat.id === chatId) {
                        return { ...chat, unreadCount: 0 };
                    }
                    return chat;
                });
            });
            
            // Also mark any related notifications as read
            markNotificationsReadByChatId(chatId);
        }
    } catch (error) {
        console.error('Error marking chat as read:', error);
    }
}

/**
 * Reset unread count for a specific chat
 */
export function resetUnreadCount(chatId: number, isGroup: boolean): void {
    activeChats.update(chats =>
        chats.map(chat =>
            chat.id === chatId && chat.isGroup === isGroup
                ? {...chat, unreadCount: 0}
                : chat
        )
    );
}

/**
 * Mark notifications for a specific chat as read
 */
export function markNotificationsReadByChatId(chatId: number): void {
    notifications.update(notificationsList => {
        const updatedNotifications = notificationsList.map(notification => {
            if (notification.chatId === chatId && !notification.isRead) {
                // If notification has an ID, also mark it as read on the server
                if (notification.id) {
                    markNotificationAsRead(notification.id);
                }
                return { ...notification, isRead: true };
            }
            return notification;
        });
        
        updateUnreadCount();
        return updatedNotifications;
    });
}

/**
 * Add a new notification to the store
 */
export function addNotification(notification: NotificationMessage): void {
    // Add to notifications store
    notifications.update(list => {
        // Check if this notification already exists to avoid duplicates
        const exists = list.some(n => 
            (n.id && n.id === notification.id) || 
            (n.content === notification.content && 
             n.createdAt === notification.createdAt &&
             n.userId === notification.userId)
        );
        
        if (!exists) {
            return [notification, ...list];
        }
        return list;
    });
    
    // Update the unread count
    updateUnreadCount();
    
    // Show browser notification if enabled
    if (!notification.isRead) {
        showBrowserNotification(notification);
    }
}

/**
 * Show browser notification
 */
export function showBrowserNotification(notification: NotificationMessage): void {
    if (Notification.permission === 'granted' && document.visibilityState !== 'visible') {
        const notif = new Notification('Social Network', {
            body: notification.content,
            icon: '/favicon.ico'
        });
        
        notif.onclick = () => {
            window.focus();
            if (notification.link) {
                window.location.href = notification.link;
            }
            notif.close();
        };
    }
}

/**
 * Helper function to update unread count
 */
function updateUnreadCount(): void {
    const unreadCount = get(notifications).filter(n => !n.isRead).length;
    unreadNotificationsCount.set(unreadCount);
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
            // Only match on chatId for chat messages
            return $messages.filter(
                msg => msg.type === MessageType.CHAT &&
                    'chatId' in msg &&
                    msg.chatId === chatId
            );
        }
    });
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

/**
 * Check if a message is a duplicate by ID or content
 */
function isMessageDuplicate(message: WebSocketMessage): boolean {
    // Skip duplicate check for certain message types
    if (message.type === 'pong' || message.type === 'ping') {
        return false;
    }
    
    // If it's a message without an ID, use content hash
    let messageKey = '';
    
    if (message.id) {
        messageKey = `${message.type}_${message.id}`;
    } else if ('content' in message) {
        // For messages without an ID but with content
        const content = (message as any).content;
        const sender = (message as any).senderId || (message as any).userId || '';
        const timestamp = message.createdAt || new Date().toISOString();
        messageKey = `${message.type}_${sender}_${content}_${timestamp}`;
    } else {
        // For messages without ID or content, just make a best effort
        messageKey = `${message.type}_${JSON.stringify(message)}_${new Date().toISOString()}`;
    }
    
    // Check if we've processed this message before
    const processed = get(processedMessageIds);
    if (processed.has(messageKey)) {
        console.log('Duplicate message detected, ignoring:', messageKey);
        return true;
    }
    
    // Add to processed set
    processed.add(messageKey);
    // Limit the size of the processed set to prevent memory leaks
    if (processed.size > 1000) {
        const iterator = processed.values();
        processed.delete(iterator.next().value);
    }
    processedMessageIds.set(processed);
    
    return false;
}

/**
 * Helper function to show toast notifications
 */
export function showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
    const event = new CustomEvent('toast', {
        detail: {message, type}
    });
    window.dispatchEvent(event);
}