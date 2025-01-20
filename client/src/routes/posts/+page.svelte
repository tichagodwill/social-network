<script lang="ts">
    import { onMount } from 'svelte';
    import { posts } from '$lib/stores/posts';
    import { auth } from '$lib/stores/auth';
    import { Button, Card, Textarea, Radio } from 'flowbite-svelte';
    import { getFormattedDate } from '$lib/dateFormater';

    let newPostTitle = '';
    let newPostContent = '';
    let mediaUrl = '';
    let privacy = 1; // Default to public
    let isSubmitting = false;

    onMount(() => {
        posts.loadPosts();
    });

    async function handleSubmitPost() {
        if (!newPostTitle.trim() || !newPostContent.trim()) return;

        isSubmitting = true;
        try {
            // Create the post data object with all required fields
            const postData = {
                title: newPostTitle.trim(),
                content: newPostContent.trim(),
                privacy: Number(privacy), // Ensure it's a number
                media: mediaUrl.trim() || undefined,
                author: $auth?.user?.id || 0
            };

            // Add the post using the store's addPost method
            await posts.addPost(postData);

            // Refresh the posts list
            await posts.loadPosts();

            // Reset form
            newPostTitle = '';
            newPostContent = '';
            mediaUrl = '';
            privacy = 1;
        } catch (error) {
            console.error('Error adding post:', error);
        } finally {
            isSubmitting = false;
        }
    }
</script>

<div class="container mx-auto px-4 py-8">
    {#if $auth.isAuthenticated}
        <Card class="mb-8">
            <form on:submit|preventDefault={handleSubmitPost} class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">
                        Post Title
                    </label>
                    <input
                            type="text"
                            bind:value={newPostTitle}
                            placeholder="Enter your post title"
                            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            required
                    />
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">
                        Post Content
                    </label>
                    <Textarea
                            bind:value={newPostContent}
                            placeholder="What's on your mind?"
                            rows={3}
                            required
                    />
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">
                        Media URL (optional)
                    </label>
                    <input
                            type="text"
                            bind:value={mediaUrl}
                            placeholder="Enter image URL"
                            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                    />
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">
                        Privacy Setting
                    </label>
                    <div class="flex items-center space-x-6">
                        <div class="flex items-center space-x-2">
                            <Radio bind:group={privacy} value={1} name="privacy">Public</Radio>
                        </div>
                        <div class="flex items-center space-x-2">
                            <Radio bind:group={privacy} value={0} name="privacy">Private</Radio>
                        </div>
                    </div>
                </div>

                <Button type="submit" disabled={isSubmitting} class="w-full">
                    {isSubmitting ? 'Posting...' : 'Post'}
                </Button>
            </form>
        </Card>
    {/if}

    <div class="space-y-6">
        {#each $posts as post (post.id)}
            <Card>
                <div class="flex justify-between items-start">
                    <div>
                        <h3 class="text-xl font-semibold mb-1">{post?.title || 'Untitled'}</h3>
                        <p class="text-sm text-gray-500">
                            {post?.created_at ? getFormattedDate(new Date(post.created_at)).diff : 'Just now'}
                        </p>
                    </div>
                    <div class="text-sm text-gray-500">
                        {post?.privacy === 1 ? 'Public' : 'Private'}
                    </div>
                </div>
                <p class="mt-4 text-gray-700">{post?.content || ''}</p>
                {#if post?.media}
                    <img src={post.media} alt="Post media" class="mt-4 rounded-lg max-h-96 w-auto" />
                {/if}
            </Card>
        {/each}
    </div>
</div>