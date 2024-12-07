import type { PageLoad } from './$types';
import type {User} from "$lib/types";
import { transformUser } from '$lib/utils/transformer'

export const load: PageLoad = async ({ params, fetch }) => {
    try {
        const response = await fetch(`http://localhost:8080/user/${params.id}`, {
            credentials: 'include'
        });
        
        if (response.ok) {
            const userData = await response.json();
            const user: User = transformUser(userData);
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
