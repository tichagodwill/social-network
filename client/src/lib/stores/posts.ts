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
        addPost: async (post: Omit<Post, 'id' | 'createdAt' | 'author'>) => {
            try {
                const response = await fetch('http://localhost:8080/posts', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify(post)
                });
                
                if (response.ok) {
                    const newPost = await response.json();
                    update(posts => [newPost, ...posts]);
                }
            } catch (error) {
                console.error('Failed to create post:', error);
            }
        }
    };
}

export const posts = createPostsStore(); 