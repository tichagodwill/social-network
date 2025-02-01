import { auth } from '$lib/stores/auth';

export async function fetchGroups() {
    try {
        const response = await fetch('http://localhost:8080/groups', {
            method: 'GET',
            credentials: 'include',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            }
        });

        if (response.status === 401) {
            auth.set({ isAuthenticated: false, user: null });
            throw new Error('Unauthorized');
        }

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to fetch groups');
        }

        return response.json();
    } catch (error) {
        console.error('Error fetching groups:', error);
        throw error;
    }
}

export async function inviteToGroup(groupId: number, username: string) {
    try {
        const response = await fetch(`http://localhost:8080/groups/${groupId}/invitations`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({ username })
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to send invitation');
        }

        return response.json();
    } catch (error) {
        console.error('Error inviting member:', error);
        throw error;
    }
}

export async function handleJoinRequest(groupId: number, requestId: number, action: 'accept' | 'reject') {
    try {
        const response = await fetch(`http://localhost:8080/groups/${groupId}/join-requests/${action}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({ requestId })
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || `Failed to ${action} request`);
        }

        return response.json();
    } catch (error) {
        console.error(`Error ${action}ing request:`, error);
        throw error;
    }
} 