import type {PageLoad} from './$types';
import type {Followers, User} from "$lib/types";
import {transformFollowers, transformUser} from '$lib/utils/transformer'
import {error} from "@sveltejs/kit";

export const load: PageLoad = async ({params, fetch}) => {
    const response = await fetch(`http://localhost:8080/user/${params.id}`, {
        credentials: 'include'
    });

    if (response.ok) {
        // "user":      userInfo,
        // "followers": followers,
        //  "following": following,
        const res = await response.json();
        const user: User = transformUser(res.user);
        const Following: Followers[] | null = res.following ? transformFollowers(res.following) as Followers[] : null;
        const Followers: Followers[] | null = res.followers ? transformFollowers(res.followers) as Followers[] : null;
        return {
            user,
            params,
            Following,
            Followers
        };
    }
    throw error(404, 'User not found');

};
