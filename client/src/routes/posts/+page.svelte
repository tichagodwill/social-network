<script lang="ts">
    import { onMount } from 'svelte';
    import { posts } from '$lib/stores/posts';
    import { auth } from '$lib/stores/auth';
    import { Button, Card, Textarea } from 'flowbite-svelte';
    import { getFormattedDate } from '$lib/dateFormater';

    let newPostContent = '';
    let isSubmitting = false;

    onMount(() => {
        posts.loadPosts();
    });

    async function handleSubmitPost() {
        if (!newPostContent.trim()) return;
        
        isSubmitting = true;
        try {
            await posts.addPost({
                content: newPostContent,
                privacy: 1, // public
                title: newPostContent.slice(0, 50) // Use first 50 chars as title
            });
            newPostContent = '';
        } finally {
            isSubmitting = false;
        }
    }
</script>

<div class="container mx-auto px-4 py-8">
    {#if $auth.isAuthenticated}
        <Card class="mb-8">
            <form on:submit|preventDefault={handleSubmitPost} class="space-y-4">
                <Textarea
                    bind:value={newPostContent}
                    placeholder="What's on your mind?"
                    rows={3}
                />
                <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? 'Posting...' : 'Post'}
                </Button>
            </form>
        </Card>
    {/if}

    <div class="space-y-6">
        {#each $posts as post}
            <Card>
                <div class="flex justify-between items-start">
                    <div>
                        <h3 class="text-lg font-semibold">{post.author}</h3>
                        <p class="text-sm text-gray-500">
                            {getFormattedDate(new Date(post.createdAt)).diff}
                        </p>
                    </div>
                </div>
                <p class="mt-4">{post.content}</p>
                {#if post.media}
                    <img src={post.media} alt="Post media" class="mt-4 rounded-lg max-h-96 w-auto" />
                {/if}
            </Card>
        {/each}
    </div>
</div> 