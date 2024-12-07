<script lang="ts">
    import {onMount} from 'svelte';
    import {followers} from '$lib/stores/followers';
    import {auth} from '$lib/stores/auth';
    import {Button, Avatar, Badge, Tabs, TabItem} from 'flowbite-svelte';
    import type {PageData} from './$types';
    import defualtProfileImg from '$lib/assets/defualt-profile.jpg'

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
        await followers.followUser(userId, $auth.user.id);
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
    <!-- Profile Header -->
    <div
            class="rounded-lg shadow-md p-6 bg-gradient-to-r from-[rgba(239,86,47,1)] to-[rgba(239,86,47,0.8)] text-white"
    >
        <div class="flex flex-col md:flex-row items-center md:items-start md:space-x-6">
            <Avatar src={data.user?.avatar} size="xl" alt="User Avatar"/>
            <div class="flex-1 text-center md:text-left mt-4 md:mt-0">
                <h1 class="text-4xl font-extrabold">{data.user?.username}</h1>
                <p class="text-gray-200 dark:text-gray-300 mt-2">{data.user?.aboutMe}</p>
            </div>
            {#if !isOwnProfile}
                <Button
                        class="mt-4 md:mt-0 transition-transform hover:scale-105"
                        color={isFollowing ? 'alternative' : 'primary'}
                        disabled={hasPendingRequest}
                        on:click={handleFollow}
                        aria-label="Follow/Unfollow Button"
                >
                    {#if hasPendingRequest}
                        <Badge color="yellow">Request Pending</Badge>
                    {:else if isFollowing}
                        <Badge color="green">Following</Badge>
                    {:else}
                        Follow
                    {/if}
                </Button>
            {/if}
        </div>
    </div>

    <!-- Tabs Section -->
    <Tabs class="mt-8">
        <TabItem title="Followers" active>
            <div class="rounded-lg shadow-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-6">
                <h3 class="text-2xl font-semibold mb-4">Followers</h3>
                {#if $followers.followers.length > 0}
                    <div class="space-y-4">
                        {#each $followers.followers as follower}
                            <div class="flex items-center space-x-4 hover:bg-gray-100 dark:hover:bg-gray-700 p-4 rounded-lg transition">
                                <Avatar src={follower.avatar || '/default-avatar.png'} alt="Follower Avatar"/>
                                <div>
                                    <p class="font-semibold text-lg">{follower.username}</p>
                                    <p class="text-sm text-gray-600 dark:text-gray-400">
                                        {follower.firstName} {follower.lastName}
                                    </p>
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="text-gray-500 dark:text-gray-400">No followers yet.</p>
                {/if}
            </div>
        </TabItem>

        <TabItem title="Following">
            <div class="rounded-lg shadow-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-6">
                <h3 class="text-2xl font-semibold mb-4">Following</h3>
                {#if $followers.following.length > 0}
                    <div class="space-y-4">
                        {#each $followers.following as following}
                            <div class="flex items-center space-x-4 hover:bg-gray-100 dark:hover:bg-gray-700 p-4 rounded-lg transition">
                                <Avatar src={following.avatar || '/default-avatar.png'} alt="Following Avatar"/>
                                <div>
                                    <p class="font-semibold text-lg">{following.username}</p>
                                    <p class="text-sm text-gray-600 dark:text-gray-400">
                                        {following.firstName} {following.lastName}
                                    </p>
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="text-gray-500 dark:text-gray-400">Not following anyone yet.</p>
                {/if}
            </div>
        </TabItem>

        {#if isOwnProfile && $followers.requests.length > 0}
            <TabItem title="Follow Requests">
                <div class="rounded-lg shadow-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-6">
                    <h3 class="text-2xl font-semibold mb-4">Follow Requests</h3>
                    {#each $followers.requests as request}
                        <div class="flex items-center justify-between hover:bg-gray-100 dark:hover:bg-gray-700 p-4 rounded-lg transition">
                            <div class="flex items-center space-x-4">
                                <Avatar src={request.followerUser?.avatar || '/default-avatar.png'}
                                        alt="Request Avatar"/>
                                <p class="font-semibold text-lg">{request.followerUser?.username}</p>
                            </div>
                            <div class="space-x-2">
                                <Button
                                        size="sm"
                                        color="primary"
                                        on:click={() => followers.handleRequest(request.id, true)}
                                        aria-label="Accept Request Button"
                                >
                                    Accept
                                </Button>
                                <Button
                                        size="sm"
                                        color="alternative"
                                        on:click={() => followers.handleRequest(request.id, false)}
                                        aria-label="Decline Request Button"
                                >
                                    Decline
                                </Button>
                            </div>
                        </div>
                    {/each}
                </div>
            </TabItem>
        {/if}
    </Tabs>
</div>
