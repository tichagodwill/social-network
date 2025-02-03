import { writable } from 'svelte/store';
import type { Post } from '$lib/types';

function createPostsStore() {
    const { subscribe, set, update } = writable<Post[]>([]);

    return {
        subscribe,
        loadPosts: async () => {
            try {
                const response = await fetch('http://localhost:8080/posts', {
                    credentials: 'include'
                });
                if (response.ok) {
                    const posts = await response.json();
                    set(posts);
                }
            } catch (error) {
                console.error('Failed to load posts:', error);
            }
        },
        addPost: async (postData: any) => {
            try {
                const response = await fetch('http://localhost:8080/posts', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify(postData)
                });

                if (!response.ok) {
                    throw new Error('Failed to create post');
                }

                const newPost = await response.json();

                // Update the posts store by adding the new post at the beginning
                update(currentPosts => {
                    if (!Array.isArray(currentPosts)) {
                        currentPosts = [];
                    }
                    return [newPost, ...currentPosts];
                });

                return newPost;
            } catch (error) {
                console.error('Failed to create post:', error);
                throw error;
            }
        }
    };
}

export const posts = createPostsStore();