<script lang="ts">
    import { onMount } from 'svelte';
    import { Card, Button, Input, Textarea, Modal, Label } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { getFormattedDate } from '$lib/dateFormater';

    export let groupId: number;
    let posts: any[] = [];
    let showCreateModal = false;
    let error = '';
    let loading = true;
    let newPost = {
        title: '',
        content: '',
        group_id: groupId
    };
    let newComments: { [key: number]: string } = {};

    async function loadPosts() {
        try {
            loading = true;
            const response = await fetch(`http://localhost:8080/groups/${groupId}/posts`, {
                credentials: 'include'
            });
            if (response.ok) {
                const data = await response.json();
                posts = Array.isArray(data) ? data : [];
            } else {
                const errorData = await response.json();
                console.error('Load posts error:', errorData);
                error = errorData.error || 'Failed to load posts';
            }
        } catch (err) {
            console.error('Failed to load posts:', err);
            error = err instanceof Error ? err.message : 'Failed to load posts';
        } finally {
            loading = false;
        }
    }

    async function createPost() {
        try {
            if (!newPost.title?.trim() || !newPost.content?.trim()) {
                error = 'Title and content are required';
                return;
            }

            const postData = {
                title: newPost.title.trim(),
                content: newPost.content.trim()
            };

            console.log('Sending post data:', postData, 'to group:', groupId);

            const response = await fetch(`http://localhost:8080/groups/${groupId}/posts`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify(postData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to create post');
            }

            const post = await response.json();
            posts = [post, ...posts];
            showCreateModal = false;
            newPost = { title: '', content: '', group_id: groupId };
            error = '';
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to create post';
            console.error('Create post error:', err);
        }
    }

    async function createComment(postId: number) {
        try {
            const content = newComments[postId];
            if (!content?.trim()) return;

            const response = await fetch(`http://localhost:8080/groups/posts/${postId}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({ content })
            });

            if (!response.ok) {
                throw new Error('Failed to create comment');
            }

            const comment = await response.json();
            posts = posts.map(post => {
                if (post.id === postId) {
                    return {
                        ...post,
                        comments: [...(post.comments || []), comment]
                    };
                }
                return post;
            });
            newComments[postId] = '';
        } catch (err) {
            console.error('Failed to create comment:', err);
        }
    }

    onMount(() => {
        loadPosts();
    });
</script>

<div class="space-y-4">
    <div class="flex justify-between items-center">
        <h3 class="text-xl font-semibold">Posts</h3>
        <Button on:click={() => showCreateModal = true}>Create Post</Button>
    </div>

    {#if error}
        <div class="p-4 text-red-800 bg-red-100 rounded-lg">
            {error}
        </div>
    {/if}

    {#if loading}
        <div class="text-center py-4">
            <div class="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
            <p class="mt-2 text-gray-500">Loading posts...</p>
        </div>
    {:else if posts.length === 0}
        <Card>
            <p class="text-gray-500 text-center">No posts yet. Be the first to create one!</p>
        </Card>
    {:else}
        {#each posts as post}
            <Card>
                <div class="space-y-4">
                    <div>
                        <h4 class="text-lg font-semibold">{post.title}</h4>
                        <p class="text-sm text-gray-500">
                            Posted by {post.author} on {getFormattedDate(new Date(post.created_at)).formated}
                        </p>
                    </div>
                    <p class="whitespace-pre-wrap">{post.content}</p>

                    <!-- Comments section -->
                    <div class="mt-4 space-y-4">
                        <h5 class="font-medium">Comments</h5>
                        {#if post.comments?.length}
                            {#each post.comments as comment}
                                <div class="pl-4 border-l-2 border-gray-200">
                                    <p class="text-sm text-gray-500">
                                        {comment.author} • {getFormattedDate(new Date(comment.created_at)).formated}
                                    </p>
                                    <p>{comment.content}</p>
                                </div>
                            {/each}
                        {/if}

                        <!-- New comment form -->
                        <div class="flex gap-2">
                            <Input
                                type="text"
                                placeholder="Write a comment..."
                                bind:value={newComments[post.id]}
                                class="flex-1"
                            />
                            <Button 
                                size="sm"
                                on:click={() => createComment(post.id)}
                            >
                                Comment
                            </Button>
                        </div>
                    </div>
                </div>
            </Card>
        {/each}
    {/if}
</div>

<Modal bind:open={showCreateModal} size="md">
    <div class="space-y-6">
        <h3 class="text-xl font-medium">Create Post</h3>
        <form on:submit|preventDefault={createPost} class="space-y-4">
            <div>
                <Label for="title">Title</Label>
                <Input
                    id="title"
                    bind:value={newPost.title}
                    required
                    placeholder="Enter post title"
                />
            </div>
            <div>
                <Label for="content">Content</Label>
                <Textarea
                    id="content"
                    bind:value={newPost.content}
                    required
                    placeholder="Write your post..."
                    rows={4}
                />
            </div>
            <div class="flex justify-end space-x-2">
                <Button 
                    color="alternative" 
                    on:click={() => showCreateModal = false}
                >
                    Cancel
                </Button>
                <Button type="submit">Create Post</Button>
            </div>
        </form>
    </div>
</Modal> 