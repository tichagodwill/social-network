<script lang="ts">
    import { onMount } from 'svelte';
    import { Card, Button, Input, Textarea, Modal, Label } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { getFormattedDate } from '$lib/dateFormater';
    import { fade, slide } from 'svelte/transition';
    import { groups } from '$lib/stores/groups';
    import type { FileUploadResponse } from '$lib/types';

    export let groupId: number;
    let posts: any[] = [];
    let showCreateModal = false;
    let error = '';
    let loading = true;
    let newPost = {
        title: '',
        content: '',
        group_id: groupId,
        media: null as File | null,
        mediaType: ''
    };
    let newComments: { [key: number]: string } = {};
    let expandedComments: { [key: number]: boolean } = {};
    let editingPost: any = null;
    let showCommentSections: { [key: number]: boolean } = {};
    let previewUrl: string | null = null;
    let fileError: string | null = null;
    let fileInput: HTMLInputElement;

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
                error = errorData.error || 'Failed to load posts';
            }
        } catch (err) {
            console.error('Failed to load posts:', err);
            error = err instanceof Error ? err.message : 'Failed to load posts';
        } finally {
            loading = false;
        }
    }

    async function handleSubmit(event: Event) {
        event.preventDefault();
        
        if (!newPost.title || !newPost.content) {
            error = 'Please fill in all required fields';
            return;
        }

        try {
            loading = true;
            const formData = new FormData();
            formData.append('title', newPost.title);
            formData.append('content', newPost.content);
            
            // Log the file being uploaded
            console.log('File to upload:', newPost.media);
            
            if (newPost.media) {
                formData.append('media', newPost.media, newPost.media.name);
            }

            // Log the FormData contents
            for (let pair of formData.entries()) {
                console.log(pair[0], pair[1]);
            }

            const response = await fetch(`http://localhost:8080/groups/${groupId}/posts`, {
                method: 'POST',
                credentials: 'include',
                body: formData
            });

            const responseText = await response.text();
            console.log('Server response:', responseText);

            if (!response.ok) {
                throw new Error(responseText || 'Failed to create post');
            }

            const post = JSON.parse(responseText);
            posts = [post, ...posts];
            showCreateModal = false;
            newPost = {
                title: '',
                content: '',
                group_id: groupId,
                media: null,
                mediaType: ''
            };
            previewUrl = null;
            if (fileInput) fileInput.value = '';
        } catch (error) {
            console.error('Error creating post:', error);
            this.error = error instanceof Error ? error.message : 'Failed to create post';
        } finally {
            loading = false;
        }
    }

    function handleFileSelect(event: Event) {
        const input = event.target as HTMLInputElement;
        const file = input.files?.[0];
        
        if (!file) return;
        
        // Check file size (10MB limit)
        if (file.size > 10 * 1024 * 1024) {
            fileError = 'File size must be less than 10MB';
            return;
        }

        // Check file type
        const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'application/pdf'];
        if (!allowedTypes.includes(file.type)) {
            fileError = 'Invalid file type. Only JPG, PNG, GIF, and PDF files are allowed.';
            return;
        }

        newPost.media = file;
        newPost.mediaType = file.type;
        fileError = null;

        // Create preview for images
        if (file.type.startsWith('image/')) {
            const reader = new FileReader();
            reader.onload = (e) => {
                previewUrl = e.target?.result as string;
            };
            reader.readAsDataURL(file);
        } else {
            previewUrl = null;
        }
    }

    async function createComment(postId: number) {
        try {
            const content = newComments[postId];
            if (!content?.trim()) return;

            const response = await fetch(`http://localhost:8080/groups/${groupId}/posts/${postId}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    content: content.trim()
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to create comment');
            }

            const comment = await response.json();
            
            // Update the posts array with the new comment
            posts = posts.map(post => {
                if (post.id === postId) {
                    const updatedComments = post.comments || [];
                    return {
                        ...post,
                        comments: [...updatedComments, comment]
                    };
                }
                return post;
            });

            // Clear the comment input
            newComments[postId] = '';
            expandedComments[postId] = true;
        } catch (err) {
            console.error('Failed to create comment:', err);
            error = err instanceof Error ? err.message : 'Failed to create comment';
        }
    }

    function toggleComments(postId: number) {
        expandedComments[postId] = !expandedComments[postId];
        expandedComments = {...expandedComments};
    }

    function toggleCommentSection(postId: number) {
        showCommentSections[postId] = !showCommentSections[postId];
        showCommentSections = {...showCommentSections};
    }

    onMount(() => {
        loadPosts();
    });
</script>

<style>
    .post-card {
        @apply transform transition-all duration-300;
    }

    .post-card:hover {
        @apply shadow-lg -translate-y-1;
    }

    .comment-section {
        @apply border-l-4 border-gray-200 dark:border-gray-700 pl-4 ml-4 mt-2;
    }

    .fade-bg {
        @apply bg-gradient-to-b from-transparent to-white dark:to-gray-800;
    }

    .comment-input {
        @apply relative;
    }

    .comment-input::before {
        content: '';
        @apply absolute -left-4 h-full w-0.5 bg-blue-500 opacity-0 transition-opacity duration-300;
    }

    .comment-input:focus-within::before {
        @apply opacity-100;
    }

    .post-content {
        @apply text-gray-700 dark:text-gray-300 leading-relaxed;
    }

    .animate-pulse {
        animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
    }

    @keyframes pulse {
        0%, 100% {
            opacity: 1;
        }
        50% {
            opacity: .5;
        }
    }

    :global(.file-input-wrapper) {
        @apply relative;
    }

    :global(.file-input) {
        @apply block w-full text-sm text-gray-900 border border-gray-300 rounded-lg cursor-pointer 
               bg-gray-50 dark:text-gray-400 focus:outline-none dark:bg-gray-700 
               dark:border-gray-600 dark:placeholder-gray-400 file:mr-4 file:py-2 
               file:px-4 file:border-0 file:text-sm file:font-semibold file:bg-primary-50 
               file:text-primary-700 hover:file:bg-primary-100;
    }
</style>

<div class="space-y-4">
    <div class="flex justify-between items-center">
        <h3 class="text-xl font-semibold">Posts</h3>
        <Button 
            gradient
            color="blue"
            class="transform hover:scale-105 transition-transform duration-200"
            on:click={() => showCreateModal = true}
        >
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
            </svg>
            Create Post
        </Button>
    </div>

    {#if error}
        <div transition:fade>
            <div class="p-4 text-red-800 bg-red-100 rounded-lg">
                {error}
            </div>
        </div>
    {/if}

    {#if loading}
        <div class="space-y-4">
            {#each Array(3) as _}
                <div class="animate-pulse">
                    <Card>
                        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-4"></div>
                        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-2"></div>
                        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
                    </Card>
                </div>
            {/each}
        </div>
    {:else if posts.length === 0}
        <Card>
            <div class="text-center py-8">
                <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
                <p class="text-gray-500 text-lg">No posts yet. Be the first to create one!</p>
            </div>
        </Card>
    {:else}
        {#each posts as post (post.id)}
            <div transition:slide>
                <Card class="post-card">
                    <div class="space-y-4">
                        <div class="flex justify-between items-start">
                            <div>
                                <h4 class="text-xl font-semibold hover:text-blue-600 transition-colors duration-200">
                                    {post.title}
                                </h4>
                                <div class="flex items-center space-x-2 text-sm text-gray-500 mt-1">
                                    <span class="font-medium text-blue-600 dark:text-blue-400">{post.author}</span>
                                    <span>•</span>
                                    <span>{getFormattedDate(new Date(post.created_at)).formated}</span>
                                </div>
                            </div>
                        </div>

                        <div class="post-content">
                            <p class="whitespace-pre-wrap text-gray-700 dark:text-gray-300">{post.content}</p>
                        </div>

                        {#if post.media}
                            <div class="mt-4">
                                {#if post.media.toLowerCase().endsWith('.pdf')}
                                    <a 
                                        href={`http://localhost:8080/uploads/group_posts/${post.media}`} 
                                        target="_blank" 
                                        rel="noopener noreferrer"
                                        class="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                                    >
                                        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                        </svg>
                                        View PDF
                                    </a>
                                {:else}
                                    <img 
                                        src={`http://localhost:8080/uploads/group_posts/${post.media}`}
                                        alt="Post attachment"
                                        class="max-h-96 rounded-lg object-contain"
                                        loading="lazy"
                                    />
                                {/if}
                            </div>
                        {/if}

                        <div class="flex items-center space-x-4 pt-2">
                            <Button 
                                size="xs"
                                color="light"
                                class="flex items-center space-x-1"
                                on:click={() => toggleComments(post.id)}
                            >
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
                                </svg>
                                <span>{post.comments?.length || 0} Comments</span>
                            </Button>
                            
                            <Button
                                size="xs"
                                color="light"
                                class="flex items-center space-x-1"
                                on:click={() => toggleCommentSection(post.id)}
                            >
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                                </svg>
                                <span>Add Comment</span>
                            </Button>
                        </div>

                        {#if expandedComments[post.id] && post.comments?.length > 0}
                            <div class="comment-section space-y-3" transition:slide>
                                {#each post.comments as comment}
                                    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                                        <div class="flex items-center justify-between mb-1">
                                            <span class="font-medium text-blue-600 dark:text-blue-400">
                                                {comment.author}
                                            </span>
                                            <span class="text-xs text-gray-500">
                                                {getFormattedDate(new Date(comment.created_at)).formated}
                                            </span>
                                        </div>
                                        <p class="text-gray-700 dark:text-gray-300">{comment.content}</p>
                                    </div>
                                {/each}
                            </div>
                        {/if}

                        {#if showCommentSections[post.id]}
                            <div class="comment-input mt-4" transition:slide>
                                <div class="flex space-x-2">
                                    <Input
                                        type="text"
                                        placeholder="Write a comment..."
                                        bind:value={newComments[post.id]}
                                        class="flex-1"
                                    />
                                    <Button 
                                        size="sm"
                                        gradient
                                        color="blue"
                                        class="transform hover:scale-105 transition-transform duration-200"
                                        on:click={() => createComment(post.id)}
                                    >
                                        Comment
                                    </Button>
                                </div>
                            </div>
                        {/if}
                    </div>
                </Card>
            </div>
        {/each}
    {/if}
</div>

<Modal bind:open={showCreateModal} size="lg">
    <div class="p-6">
        <h3 class="text-xl font-semibold mb-4">Create New Post</h3>
        {#if error}
            <div class="text-red-500 mb-4">{error}</div>
        {/if}
        <form on:submit={handleSubmit} class="space-y-4">
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
                    rows={6}
                />
            </div>
            
            <div class="space-y-2">
                <Label for="media">Upload File (optional)</Label>
                <div class="file-input-wrapper">
                    <input
                        type="file"
                        id="media"
                        bind:this={fileInput}
                        accept=".jpg,.jpeg,.png,.gif,.pdf"
                        on:change={handleFileSelect}
                        class="file-input"
                    />
                </div>
                {#if fileError}
                    <p class="text-red-500 text-sm">{fileError}</p>
                {/if}
                <p class="text-sm text-gray-500">
                    Supported formats: JPG, PNG, GIF, PDF (max 10MB)
                </p>
            </div>
            
            {#if previewUrl}
                <div class="relative">
                    <img
                        src={previewUrl}
                        alt="Preview"
                        class="max-h-48 rounded-lg"
                    />
                    <button
                        type="button"
                        class="absolute top-2 right-2 p-1 bg-red-500 text-white rounded-full hover:bg-red-600"
                        on:click={() => {
                            newPost.media = null;
                            previewUrl = null;
                            if (fileInput) fileInput.value = '';
                        }}
                    >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
            {/if}
            
            <div class="flex justify-end space-x-2">
                <Button
                    color="alternative"
                    on:click={() => {
                        showCreateModal = false;
                        newPost = {
                            title: '',
                            content: '',
                            group_id: groupId,
                            media: null,
                            mediaType: ''
                        };
                        previewUrl = null;
                        if (fileInput) fileInput.value = '';
                    }}
                >
                    Cancel
                </Button>
                <Button type="submit">Create Post</Button>
            </div>
        </form>
    </div>
</Modal> 