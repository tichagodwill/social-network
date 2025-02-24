// src/lib/stores/toast.ts
import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

interface ToastState {
    message: string;
    type: ToastType;
    visible: boolean;
}

function createToastStore() {
    const { subscribe, set, update } = writable<ToastState>({
        message: '',
        type: 'info',
        visible: false
    });

    let timeoutId: ReturnType<typeof setTimeout> | null = null;

    // Create the store object
    const store = {
        subscribe,
        show: (message: string, type: ToastType = 'info', duration: number = 3000) => {
            // Clear any existing timeout
            if (timeoutId) {
                clearTimeout(timeoutId);
            }

            // Show the toast
            set({ message, type, visible: true });

            // Set timeout to hide toast
            timeoutId = setTimeout(() => {
                set({ message: '', type: 'info', visible: false });
                timeoutId = null;
            }, duration);
        },
        hide: () => {
            if (timeoutId) {
                clearTimeout(timeoutId);
                timeoutId = null;
            }
            set({ message: '', type: 'info', visible: false });
        }
    };

    // Add helper methods to the store object
    return {
        ...store,
        success: (message: string, duration = 3000) => {
            store.show(message, 'success', duration);
        },
        error: (message: string, duration = 4000) => {
            store.show(message, 'error', duration);
        },
        info: (message: string, duration = 3000) => {
            store.show(message, 'info', duration);
        },
        warning: (message: string, duration = 3000) => {
            store.show(message, 'warning', duration);
        }
    };
}

export const toast = createToastStore();