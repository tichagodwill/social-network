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
        const [groupResponse, membersResponse] = await Promise.all([
            fetch(`http://localhost:8080/groups/${params.id}`, {
                credentials: 'include'
            }),
            fetch(`http://localhost:8080/groups/${params.id}/members`, {
                credentials: 'include'
            })
        ]);

        if (!groupResponse.ok) {
            if (groupResponse.status === 404) {
                throw error(404, 'Group not found');
            }
            throw error(groupResponse.status, 'Failed to load group');
        }

        const group = await groupResponse.json();
        const members = membersResponse.ok ? await membersResponse.json() : [];

        return {
            group,
            members,
            id: params.id
        };
    } catch (err) {
        console.error('Error loading group:', err);
        if (err.status) {
            throw err;
        }
        throw error(500, 'Error loading group');
    }
}; 