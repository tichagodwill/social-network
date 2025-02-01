import type { PageLoad } from './$types';
import { auth } from '$lib/stores/auth';
import { get } from 'svelte/store';
import { fetchGroups } from '$lib/api/groupApi';

export const load: PageLoad = async () => {
    try {
        const authState = get(auth);
        
        if (!authState.isAuthenticated) {
            return {
                groups: [],
                error: null
            };
        }

        const groups = await fetchGroups();
        return { 
            groups,
            error: null
        };
    } catch (err) {
        console.error('Error loading groups:', err);
        return {
            groups: [],
            error: err instanceof Error ? err.message : 'Error loading groups'
        };
    }
};
