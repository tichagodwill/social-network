import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';
import { auth } from '$lib/stores/auth';
import { get } from 'svelte/store';

export const load: PageLoad = async ({ fetch }) => {
    const authState = get(auth);
    
    if (!authState.isAuthenticated) {
        return {
            groups: []
        };
    }

    try {
        const response = await fetch('http://localhost:8080/groups', {
            credentials: 'include'
        });

        if (response.status === 401) {
            return {
                groups: []
            };
        }

        if (!response.ok) {
            throw error(response.status, 'Failed to load groups');
        }

        const groups = await response.json();
        return { groups };
    } catch (err) {
        console.error('Error loading groups:', err);
        throw error(500, 'Error loading groups');
    }
}; 