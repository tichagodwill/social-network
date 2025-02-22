import { writable } from 'svelte/store';

type ToastType = 'success' | 'error' | 'info' | 'warning';

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

    return {
        subscribe,
        show: (message: string, type: ToastType) => {
            set({ message, type, visible: true });
            setTimeout(() => {
                set({ message: '', type: 'info', visible: false });
            }, 3000);
        },
        hide: () => set({ message: '', type: 'info', visible: false })
    };
}

export const toast = createToastStore(); 