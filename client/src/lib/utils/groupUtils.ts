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

        const result = await response.json();
        return result;
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

        return await response.json();
    } catch (error) {
        console.error(`Error ${action}ing request:`, error);
        throw error;
    }
} 