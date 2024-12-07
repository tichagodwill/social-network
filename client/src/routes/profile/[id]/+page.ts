import type { PageLoad } from './$types';
import type {User} from "$lib/types";

export const load: PageLoad = async ({ params, fetch }) => {
    try {
        const response = await fetch(`http://localhost:8080/user/${params.id}`, {
            credentials: 'include'
        });
        
        if (response.ok) {
            var userData = await response.json();
            const user = transformUser(userData);
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

function transformUser(data: any): User {
    return {
        id: data.id,
        email: data.email,
        firstName: data.first_name,
        lastName: data.last_name,
        dateOfBirth: data.date_of_birth,
        username: data.username,
        aboutMe: data.about_me,
        createdAt: data.created_at,
        isPrivate: data.IsPrivate,
        avatar: data.avatar
    };
}