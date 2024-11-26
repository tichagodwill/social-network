import { writable } from 'svelte/store';
import { goto } from '$app/navigation';

interface User {
    id: number;
    username: string;
    email: string;
    firstName: string;
    lastName: string;
    avatar?: string;
    aboutMe?: string;
    isPrivate: boolean;
    dateOfBirth: string;
}

interface AuthState {
    user: User | null;
    isAuthenticated: boolean;
    loading: boolean;
    error: string | null;
}

function createAuthStore() {
    const { subscribe, set, update } = writable<AuthState>({
        user: null,
        isAuthenticated: false,
        loading: true,
        error: null
    });

    return {
        subscribe,
        login: async (email: string, password: string) => {
            try {
                const response = await fetch('http://localhost:8080/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({ email, password })
                });

                if (!response.ok) {
                    const errorData = await response.json();
                    throw new Error(errorData.error || 'Login failed');
                }

                const data = await response.json();
                set({ 
                    user: {
                        id: data.id,
                        username: data.username,
                        email: data.email,
                        firstName: data.firstName,
                        lastName: data.lastName,
                        avatar: data.avatar,
                        aboutMe: data.aboutMe,
                        isPrivate: data.isPrivate,
                        dateOfBirth: data.dateOfBirth
                    },
                    isAuthenticated: true,
                    loading: false,
                    error: null
                });
                goto('/');
            } catch (error) {
                console.error('Login failed:', error);
                throw error;
            }
        },
        logout: async () => {
            try {
                const response = await fetch('http://localhost:8080/logout', {
                    method: 'POST',
                    credentials: 'include'
                });

                if (!response.ok) {
                    throw new Error('Logout failed');
                }
            } catch (error) {
                console.error('Logout failed:', error);
            } finally {
                set({
                    user: null,
                    isAuthenticated: false,
                    loading: false,
                    error: null
                });
                goto('/login');
            }
        },
        initialize: async () => {
            try {
                const response = await fetch('http://localhost:8080/user/current', {
                    credentials: 'include'
                });
                
                if (response.ok) {
                    const data = await response.json();
                    set({ 
                        user: {
                            id: data.id,
                            username: data.username,
                            email: data.email,
                            firstName: data.firstName,
                            lastName: data.lastName,
                            avatar: data.avatar,
                            aboutMe: data.aboutMe,
                            isPrivate: data.isPrivate,
                            dateOfBirth: data.dateOfBirth
                        },
                        isAuthenticated: true,
                        loading: false,
                        error: null
                    });
                } else {
                    set({
                        user: null,
                        isAuthenticated: false,
                        loading: false,
                        error: null
                    });
                    if (window.location.pathname !== '/login' && 
                        window.location.pathname !== '/register' && 
                        window.location.pathname !== '/') {
                        goto('/login');
                    }
                }
            } catch (error) {
                console.error('Failed to initialize auth state:', error);
                set({
                    user: null,
                    isAuthenticated: false,
                    loading: false,
                    error: 'Failed to connect to server'
                });
            }
        }
    };
}

export const auth = createAuthStore(); 