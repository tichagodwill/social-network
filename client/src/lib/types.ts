export interface Post {
    id: number;
    title: string;
    content: string;
    media?: string;
    privacy: number;
    author: number;
    authorName: string;
    created_at: string;
    groupId?: number;

    // Additional helpful fields
    authorAvatar?: string;
    isEdited?: boolean;
    lastEditedAt?: string;
    mediaType?: 'image' | 'gif';
    likes?: number;
    comments?: Comment[];
    privacyLabel?: 'Public' | 'Private' | 'Almost Private';
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

export interface Followers {
    id: number;
    userId: number;
    username: string;
    avatar?: string;
    firstName: string;
    lastName: string;
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
    type: string;
    content: string;
    group_id?: number;
    invitation_id?: number;
    user_id?: number;
    from_user_id?: number;
    is_read?: boolean;
    created_at?: string;
    user_role?: string;
    is_processed?: boolean;
    groupId?: number;
    invitationId?: number;
    userId?: number;
    fromUserId?: number;
    isRead?: boolean;
    createdAt?: string;
    userRole?: string;
    isProcessed?: boolean;
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

export interface EmojiPickerEvent {
    emoji: {
        native: string;
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
    creator_username?: string;
    created_at: string;
    is_member?: boolean;
}

export interface GroupMember {
    id: number;
    username: string;
    role: 'creator' | 'admin' | 'moderator' | 'member';
}

export interface GroupPost {
    id: number;
    title: string;
    content: string;
    author_id: number;
    author: string;
    created_at: string;
    comments?: GroupComment[];
}

export interface GroupComment {
    id: number;
    content: string;
    author_id: number;
    author: string;
    created_at: string;
}

export interface Comment {
    id: number;
    content: string;
    author_id: number;
    author: string;
    avatar:string;
    created_at: string;
    author_name:string;
}
