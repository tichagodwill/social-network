<script lang="ts">
    import { onMount } from 'svelte';
    import { followers } from '$lib/stores/followers';
    import { auth } from '$lib/stores/auth';
    import { Button, Card, Avatar } from 'flowbite-svelte';
    import type { PageData } from './$types';

    export let data: PageData;
    const userId = parseInt(data.params.id);
    
    let isOwnProfile = false;
    let isFollowing = false;
    let hasPendingRequest = false;

    $: if ($auth.user) {
        isOwnProfile = $auth.user.id === userId;
    }

    onMount(async () => {
        await followers.loadFollowers(userId);
    });

    async function handleFollow() {
        await followers.followUser(userId);
    }

    async function getUser(userId: number) {
        try {
            const response = await fetch(`http://localhost:8080/user/${userId}`, {
                credentials: 'include'
            });
            if (response.ok) {
                return await response.json();
            }
        } catch (error) {
            console.error('Failed to fetch user:', error);
        }
        return null;
    }

    $: if ($followers.requests.length > 0) {
        $followers.requests.forEach(async (request) => {
            const followerUser = await getUser(request.followerId);
            request.followerUser = followerUser;
        });
    }
</script>

<div class="container mx-auto px-4 py-8">
    <Card class="mb-8">
        <div class="flex items-center space-x-4">
            <Avatar src={data.user?.avatar || '/default-avatar.png'} size="xl" />
            <div>
                <h2 class="text-2xl font-bold">{data.user?.username}</h2>
                <p class="text-gray-600 dark:text-gray-400">{data.user?.aboutMe}</p>
            </div>
            {#if !isOwnProfile}
                <Button 
                    color={isFollowing ? 'alternative' : 'primary'}
                    disabled={hasPendingRequest}
                    on:click={handleFollow}
                >
                    {#if hasPendingRequest}
                        Request Pending
                    {:else if isFollowing}
                        Following
                    {:else}
                        Follow
                    {/if}
                </Button>
            {/if}
        </div>
    </Card>

    <div class="grid md:grid-cols-2 gap-8">
        <Card>
            <h3 class="text-xl font-semibold mb-4">Followers</h3>
            {#each $followers.followers as follower}
                <div class="flex items-center space-x-4 mb-4">
                    <Avatar src={follower.avatar || '/default-avatar.png'} />
                    <div>
                        <p class="font-semibold">{follower.username}</p>
                        <p class="text-sm text-gray-600 dark:text-gray-400">
                            {follower.firstName} {follower.lastName}
                        </p>
                    </div>
                </div>
            {/each}
        </Card>

        <Card>
            <h3 class="text-xl font-semibold mb-4">Following</h3>
            {#each $followers.following as following}
                <div class="flex items-center space-x-4 mb-4">
                    <Avatar src={following.avatar || '/default-avatar.png'} />
                    <div>
                        <p class="font-semibold">{following.username}</p>
                        <p class="text-sm text-gray-600 dark:text-gray-400">
                            {following.firstName} {following.lastName}
                        </p>
                    </div>
                </div>
            {/each}
        </Card>
    </div>

    {#if isOwnProfile && $followers.requests.length > 0}
        <Card class="mt-8">
            <h3 class="text-xl font-semibold mb-4">Follow Requests</h3>
            {#each $followers.requests as request}
                <div class="flex items-center justify-between mb-4">
                    <div class="flex items-center space-x-4">
                        <Avatar src={request.followerUser?.avatar || '/default-avatar.png'} />
                        <p class="font-semibold">{request.followerUser?.username}</p>
                    </div>
                    <div class="space-x-2">
                        <Button 
                            size="sm" 
                            color="primary"
                            on:click={() => followers.handleRequest(request.id, true)}
                        >
                            Accept
                        </Button>
                        <Button 
                            size="sm" 
                            color="alternative"
                            on:click={() => followers.handleRequest(request.id, false)}
                        >
                            Decline
                        </Button>
                    </div>
                </div>
            {/each}
        </Card>
    {/if}
</div> 