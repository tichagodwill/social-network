import type { PageLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { auth } from '$lib/stores/auth';
import { get } from 'svelte/store';
export const load: PageLoad = async () => {
    const authState = get(auth);
    
    if (!authState.isAuthenticated) {
        throw redirect(302, '/login');
    }
    return {};
}; 