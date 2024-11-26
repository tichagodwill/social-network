<script lang="ts">
    import { onMount } from 'svelte';
    import { Card, Button } from 'flowbite-svelte';
    import { getFormattedDate } from '$lib/dateFormater';
    import CreateGroupModal from '$lib/components/CreateGroupModal.svelte';
    import type { Group } from '$lib/types';

    let groups: Group[] = [];
    let loading = true;
    let error = '';
    let showCreateModal = false;

    async function loadGroups() {
        try {
            const response = await fetch('http://localhost:8080/groups', {
                credentials: 'include'
            });
            if (response.ok) {
                groups = await response.json();
            } else {
                error = 'Failed to load groups';
            }
        } catch (err) {
            error = 'Error connecting to server';
            console.error('Failed to fetch groups:', err);
        } finally {
            loading = false;
        }
    }

    onMount(loadGroups);

    function handleGroupCreated() {
        loadGroups();
    }
</script>

<div class="container mx-auto px-4 py-8">
    <div class="flex justify-between items-center mb-6">
        <h1 class="text-3xl font-bold dark:text-white">Groups</h1>
        <Button on:click={() => showCreateModal = true}>Create Group</Button>
    </div>

    {#if error}
        <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4">
            {error}
        </div>
    {/if}

    {#if loading}
        <div class="text-center">
            <p>Loading groups...</p>
        </div>
    {:else if groups.length === 0}
        <div class="text-center">
            <p>No groups found. Create one to get started!</p>
        </div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {#each groups as group}
                <Card>
                    <div class="flex flex-col h-full">
                        <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
                            {group.title}
                        </h5>
                        <p class="mb-3 font-normal text-gray-700 dark:text-gray-400 flex-grow">
                            {group.description}
                        </p>
                        <div class="mt-4 flex justify-between items-center">
                            <span class="text-sm text-gray-500">
                                {#if group.createdAt}
                                    Created {getFormattedDate(group.createdAt).diff}
                                {/if}
                            </span>
                            <Button href="/groups/{group.id}">
                                View Group
                            </Button>
                        </div>
                    </div>
                </Card>
            {/each}
        </div>
    {/if}
</div>

<CreateGroupModal
    bind:open={showCreateModal}
    onClose={() => showCreateModal = false}
    onGroupCreated={handleGroupCreated}
/> 