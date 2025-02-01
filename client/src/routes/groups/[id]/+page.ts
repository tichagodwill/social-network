import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params, fetch }) => {
    try {
        const groupResponse = await fetch(`http://localhost:8080/groups/${params.id}`, {
            credentials: 'include'
        });
        
        if (!groupResponse.ok) {
            throw new Error('Failed to load group');
        }
        
        const group = await groupResponse.json();
        
        const membersResponse = await fetch(`http://localhost:8080/groups/${params.id}/members`, {
            credentials: 'include'
        });
        
        if (!membersResponse.ok) {
            throw new Error('Failed to load members');
        }
        
        const members = await membersResponse.json();
        
        return {
            group,
            members
        };
    } catch (err) {
        throw error(404, {
            message: 'Not found'
        });
    }
}; 