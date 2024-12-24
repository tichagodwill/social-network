import { writable } from 'svelte/store';
import type { Notification } from '$lib/types';

interface NotificationState {
    notifications: Notification[];
    unreadCount: number;
    socket: WebSocket | null;
}

function createNotificationStore() {
    const { subscribe, set, update } = writable<NotificationState>({
        notifications: [],
        unreadCount: 0,
        socket: null
    });

    return {
        subscribe,
        initialize: (socket: WebSocket) => {
            update(state => ({ ...state, socket }));

            socket.addEventListener('message', (event) => {
                const data = JSON.parse(event.data);
                if (data.type === 'notification') {
                    update(state => ({
                        ...state,
                        notifications: [data.notification, ...state.notifications],
                        unreadCount: state.unreadCount + 1
                    }));
                }
            });
        },
        loadNotifications: async () => {
            try {
                const response = await fetch('http://localhost:8080/notifications', {
                    credentials: 'include'
                });
                if (response.ok) {
                    const notifications = await response.json();
                    update(state => ({
                        ...state, 
                        notifications,
                        unreadCount: notifications.filter((n: Notification) => !n.isRead).length
                    }));
                }
            } catch (error) {
                console.error('Failed to load notifications:', error);
            }
        },
        markAsRead: async (notificationId: number) => {
            try {
                const response = await fetch(`http://localhost:8080/notifications/${notificationId}/read`, {
                    method: 'GET',
                    credentials: 'include'
                });
                if (response.ok) {
                    update(state => ({
                        ...state,
                        notifications: state.notifications.map(n => 
                            n.id === notificationId ? { ...n, isRead: true } : n
                        ),
                        unreadCount: state.unreadCount - 1
                    }));
                }
            } catch (error) {
                console.error('Failed to mark notification as read:', error);
            }
        }
    };
}

export const notifications = createNotificationStore(); 