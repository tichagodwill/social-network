<script lang="ts">
    import { Button, Card } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';

    export let data;
    let groups = data.groups || [];
</script>

<div class="max-w-4xl mx-auto p-4 space-y-8">
    <div class="flex justify-between items-center">
        <h1 class="text-3xl font-bold">Groups</h1>
        {#if $auth.isAuthenticated}
            <Button href="/groups/create">Create Group</Button>
        {/if}
    </div>

    {#if !$auth.isAuthenticated}
        <div class="p-4 text-center">
            <p class="text-lg mb-4">Please log in to view and interact with groups</p>
            <Button href="/login">Log In</Button>
        </div>
    {:else if groups.length === 0}
        <p class="text-gray-500">No groups found</p>
    {:else}
        <div class="grid gap-6">
            {#each groups as group}
                <Card>
                    <div class="flex justify-between items-start">
                        <div>
                            <h2 class="text-xl font-semibold">
                                <a href="/groups/{group.id}" class="hover:underline">
                                    {group.title}
                                </a>
                            </h2>
                            <p class="text-gray-600 dark:text-gray-400 mt-2">
                                {group.description}
                            </p>
                            <p class="text-sm text-gray-500 mt-2">
                                Created by: {group.creator_username}
                            </p>
                        </div>
                        {#if !group.isMember && !group.hasPendingRequest}
                            <Button 
                                size="sm"
                                href="/groups/{group.id}"
                            >
                                View Group
                            </Button>
                        {:else if group.hasPendingRequest}
                            <span class="text-sm text-gray-500">Request Pending</span>
                        {:else}
                            <Button 
                                size="sm"
                                href="/groups/{group.id}"
                            >
                                View Group
                            </Button>
                        {/if}
                    </div>
                </Card>
            {/each}
        </div>
    {/if}
</div> 