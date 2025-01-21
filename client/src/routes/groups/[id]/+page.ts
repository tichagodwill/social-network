import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';
import { auth } from '$lib/stores/auth';
import { get } from 'svelte/store';

export const load: PageLoad = async ({ params, fetch }) => {
    const authState = get(auth);
    
    if (!authState.isAuthenticated) {
        return {
            group: null,
            members: [],
            error: 'Please login to view this group'
        };
    }

    try {
        // Load group data
        const groupResponse = await fetch(`http://localhost:8080/groups/${params.id}`, {
            credentials: 'include'
        });

        if (!groupResponse.ok) {
            if (groupResponse.status === 404) {
                throw error(404, 'Group not found');
            }
            const errorData = await groupResponse.json();
            throw error(groupResponse.status, errorData.error || 'Failed to load group');
        }

        const group = await groupResponse.json();

        // Load members data
        const membersResponse = await fetch(`http://localhost:8080/groups/${params.id}/members`, {
            credentials: 'include'
        });

        const members = membersResponse.ok ? await membersResponse.json() : [];

        if (!group) {
            throw error(404, 'Group not found');
        }

        return {
            group,
            members,
            error: null
        };
    } catch (err) {
        console.error('Error loading group:', err);
        if (err.status) {
            throw err;
        }
        throw error(500, 'Error loading group');
    }
}; 