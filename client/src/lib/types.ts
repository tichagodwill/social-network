export interface Post {
    id: number;
    title: string;
    content: string;
    media?: string;
    privacy: number;
    author: number;
    created_at?: string;
    groupId?: number;
}

export interface User {
    id: number;
    email: string;
    username: string;
    firstName: string;
    lastName: string;
    dateOfBirth: string;
    avatar?: string;
    aboutMe?: string;
    isPrivate: boolean;
    createdAt?: string;
}

export interface FollowRequest {
    id: number;
    followerId: number;
    followedId: number;
    status: 'pending' | 'accepted' | 'rejected';
    createdAt: string;
    follower?: {
        id: number;
        username: string;
        avatar?: string;
    };
    followed?: {
        id: number;
        username: string;
        avatar?: string;
    };
}

export interface Notification {
    id: number;
    toUserId: number;
    fromUserId: number;
    content: string;
    type: 'follow_request' | 'group_invite' | 'group_event' | 'post_like' | 'comment' | 'chat';
    isRead: boolean;
    createdAt: string;
    groupId?: number;
    eventId?: number;
    postId?: number;
}

export interface FileEvent extends Event {
    target: HTMLInputElement & EventTarget;
}

export interface Message {
    id: number;
    senderId: number;
    recipientId?: number;
    groupId?: number;
    content: string;
    fileUrl?: string;
    fileName?: string;
    fileType?: string;
    createdAt: string;
    senderName?: string;
    senderAvatar?: string;
}

export interface EmojiPickerEvent extends CustomEvent {
    detail: {
        emoji: {
            native: string;
        };
    };
}

export interface RegisterData {
    email: string;
    password: string;
    confirmPassword: string;
    firstName: string;
    lastName: string;
    dateOfBirth: string;
    avatar: string;
    username: string;
    aboutMe: string;
}

export interface FileUploadResponse {
    url: string;
    fileName: string;
    fileType: string;
}

export interface FilePreview {
    file: File;
    preview: string;
    type: 'image' | 'video' | 'audio' | 'document';
}

export interface CustomEvent<T = any> extends Event {
    detail: T;
}

export interface CreateGroupRequest {
    title: string;
    description: string;
    creator_id: number;
}

export interface Group {
    id: number;
    title: string;
    description: string;
    creator_id: number;
    created_at?: string;
} 