import {writable} from 'svelte/store';
import type {User} from '$lib/types';

interface FollowRequest {
    id: number;
    followerId: number;
    followedId: number;
    status: 'pending' | 'accepted' | 'rejected';
    createdAt: string;
}

function createFollowersStore() {
    const {subscribe, set, update} = writable<{
        followers: User[];
        following: User[];
        requests: FollowRequest[];
    }>({
        followers: [],
        following: [],
        requests: []
    });

    return {
        subscribe,
        followUser: async (userId: number) => {
            try {
                const response = await fetch('http://localhost:8080/follow', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        userToFollowId: userId,       // User to follow
                    }),
                });
                if (response.ok) {
                    // Refresh followers list
                    const data = await response.json();
                    return data;
                }
            } catch (error) {
                console.error('Failed to follow user:', error);
            }
        },
        handleRequest: async (requestId: number, accept: boolean) => {
            try {
                const response = await fetch(`http://localhost:8080/follow/${requestId}`, {
                    method: 'PATCH',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({accept})
                });
                if (response.ok) {
                    // Remove request from list
                    update(state => ({
                        ...state,
                        requests: state.requests.filter(r => r.id !== requestId)
                    }));
                }
            } catch (error) {
                console.error('Failed to handle follow request:', error);
            }
        }
    };
}

export const followers = createFollowersStore(); 