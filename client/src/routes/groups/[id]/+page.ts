import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params, fetch }) => {
    try {
        const authResponse = await fetch('http://localhost:8080/user/current', {
            credentials: 'include'
        });

        if (!authResponse.ok) {
            throw error(401, 'Unauthorized');
        }

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
        if (err.status) {
            throw err;
        }
        console.error('Error loading group:', err);
        throw error(500, 'Error loading group');
    }
}; 