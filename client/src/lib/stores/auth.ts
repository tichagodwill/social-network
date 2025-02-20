import { writable, derived, get } from 'svelte/store';
import { goto } from '$app/navigation';
import { browser } from '$app/environment';
import { initializeWebSocket, closeConnection } from '$lib/stores/websocket';

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

    const store = {
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

                // Initialize WebSocket connection after successful login
                if (browser) {
                    initializeWebSocket();
                }

                goto('/');
            } catch (error) {
                console.error('Login failed:', error);
                throw error;
            }
        },
        logout: async () => {
            try {
                // Close WebSocket connection before logging out
                if (browser) {
                    closeConnection();
                }

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
            if (!browser) {
                set({
                    user: null,
                    isAuthenticated: false,
                    loading: false,
                    error: null
                });
                return;
            }

            try {
                const response = await fetch('http://localhost:8080/user/current', {
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json'
                    }
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

                    // Initialize WebSocket connection after authentication check
                    initializeWebSocket();

                } else {
                    set({
                        user: null,
                        isAuthenticated: false,
                        loading: false,
                        error: null
                    });

                    const currentPath = window.location.pathname;
                    const publicRoutes = ['/login', '/register', '/'];
                    if (!publicRoutes.includes(currentPath)) {
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
        },
        register: async (userData: {
            email: string;
            password: string;
            username: string;
            firstName: string;
            lastName: string;
            dateOfBirth: string;
            avatar?: string;
            aboutMe?: string;
        }) => {
            try {
                // Parse and format the date
                const dateOfBirth = new Date(userData.dateOfBirth);
                const formattedDate = dateOfBirth.toISOString();

                const requestData = {
                    email: userData.email,
                    password: userData.password,
                    username: userData.username,
                    first_name: userData.firstName,
                    last_name: userData.lastName,
                    date_of_birth: formattedDate,
                    avatar: userData.avatar || "",
                    about_me: userData.aboutMe || ""
                };

                console.log('Sending registration data:', requestData);

                const response = await fetch('http://localhost:8080/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify(requestData)
                });

                const responseText = await response.text();
                console.log('Raw response:', responseText);

                if (!response.ok) {
                    let errorMessage: string;
                    try {
                        const errorData = JSON.parse(responseText);
                        errorMessage = errorData.error || 'Registration failed';
                    } catch (e) {
                        errorMessage = responseText || 'Registration failed';
                    }
                    throw new Error(errorMessage);
                }

                const data = JSON.parse(responseText);

                set({
                    user: {
                        id: data.id,
                        username: data.username,
                        email: userData.email,
                        firstName: userData.firstName,
                        lastName: userData.lastName,
                        avatar: userData.avatar,
                        aboutMe: userData.aboutMe,
                        isPrivate: false,
                        dateOfBirth: userData.dateOfBirth
                    },
                    isAuthenticated: true,
                    loading: false,
                    error: null
                });

                // Initialize WebSocket connection after successful registration
                if (browser) {
                    initializeWebSocket();
                    goto('/');
                }
            } catch (error) {
                console.error('Registration failed:', error);
                throw error;
            }
        },
        // Add utility functions for WebSocket service
        getCurrentUserId: () => {
            const state = get({ subscribe });
            return state.user?.id || null;
        },
        isLoggedIn: () => {
            const state = get({ subscribe });
            return state.isAuthenticated;
        }
    };

    if (browser) {
        store.initialize();
    }

    return store;
}

export const auth = createAuthStore();

// Derived stores for WebSocket service
export const currentUserId = derived(
    auth,
    $auth => $auth.user?.id
);

export const isAuthenticated = derived(
    auth,
    $auth => $auth.isAuthenticated
);