import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
    try {
        const response = await fetch(`http://localhost:8080/user/${params.id}`, {
            credentials: 'include'
        });
        
        if (response.ok) {
            const user = await response.json();
            return {
                user,
                params
            };
        }
        
        return {
            user: null,
            params
        };
    } catch (error) {
        console.error('Failed to load user profile:', error);
        return {
            user: null,
            params
        };
    }
}; 