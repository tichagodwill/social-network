import { writable } from 'svelte/store';
import type { Group } from '$lib/types';

function createGroupsStore() {
    const { subscribe, set, update } = writable<Group[]>([]);

    return {
        subscribe,
        loadGroups: async () => {
            const response = await fetch('http://localhost:8080/groups', {
                credentials: 'include'
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to load groups');
            }
            
            const groups = await response.json();
            set(groups);
            return groups;
        },

        getGroup: async (id: number) => {
            const response = await fetch(`http://localhost:8080/groups/${id}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to load group');
            }
            
            return await response.json();
        },

        updateGroup: async (id: number, data: { title: string; description: string }) => {
            const response = await fetch(`http://localhost:8080/groups/${id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify(data)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to update group');
            }

            return await response.json();
        },

        deleteGroup: async (id: number) => {
            const response = await fetch(`http://localhost:8080/groups/${id}`, {
                method: 'DELETE',
                credentials: 'include'
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to delete group');
            }

            return await response.json();
        },

        createPost: async (groupId: number, postData: { title: string; content: string }) => {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/posts`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify(postData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to create post');
            }

            return await response.json();
        },

        createEvent: async (groupId: number, eventData: { 
            title: string; 
            description: string; 
            eventDate: string;
            creatorId: number;
        }) => {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/events`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify(eventData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to create event');
            }

            return await response.json();
        },

        inviteMember: async (groupId: number, data: { 
            identifier: string; 
            identifierType: 'email' | 'username' 
        }) => {
            const response = await fetch(`http://localhost:8080/groups/invitation`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    groupId,
                    identifier: data.identifier,
                    identifierType: data.identifierType
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to send invitation');
            }

            return await response.json();
        }
    };
}

export const groups = createGroupsStore();

export async function getInvitationStatus(groupId: number) {
    try {
        const response = await fetch(`http://localhost:8080/groups/${groupId}/invitations/status`, {
            credentials: 'include'
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to get invitation status');
        }

        return await response.json();
    } catch (error) {
        console.error('Error getting invitation status:', error);
        throw error;
    }
} 