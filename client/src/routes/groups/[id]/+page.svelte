<script lang="ts">
    import { auth } from '$lib/stores/auth';
    import { Button, Card, Modal, Label, Input, Textarea } from 'flowbite-svelte';
    import { goto } from '$app/navigation';
    import GroupMembership from '$lib/components/GroupMembership.svelte';
    import GroupEvents from '$lib/components/GroupEvents.svelte';
    import GroupJoinRequests from '$lib/components/GroupJoinRequests.svelte';

    export let data;

    let group = data.group;
    let members = data.members || [];
    $: isCreator = $auth.user ? group.creator_id === $auth.user.id : false;
    $: isMember = $auth.user && members ? members.some(m => m.id === $auth.user?.id) : false;
    let showEditModal = false;
    let error = '';

    let editForm = {
        title: group.title,
        description: group.description
    };

    $: canViewContent = isCreator || isMember;

    async function handleEdit() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${group.id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify(editForm)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to update group');
            }

            // Update local data
            group.title = editForm.title;
            group.description = editForm.description;
            showEditModal = false;
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to update group';
        }
    }

    async function handleDelete() {
        if (!confirm('Are you sure you want to delete this group? This action cannot be undone.')) {
            return;
        }

        try {
            const response = await fetch(`http://localhost:8080/groups/${group.id}`, {
                method: 'DELETE',
                credentials: 'include'
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to delete group');
            }

            goto('/groups');
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to delete group';
        }
    }
</script>

<div class="max-w-4xl mx-auto p-4 space-y-8">
    {#if !$auth.user}
        <div class="text-center p-8">
            <p class="text-lg mb-4">Please log in to view this group</p>
            <Button href="/login">Log In</Button>
        </div>
    {:else}
        {#if error}
            <div class="p-4 text-red-800 bg-red-100 rounded-lg">
                {error}
            </div>
        {/if}

        <Card>
            <div class="space-y-4">
                <h1 class="text-3xl font-bold">{group.title}</h1>
                <p class="text-gray-600 dark:text-gray-400">{group.description}</p>
                
                {#if isCreator}
                    <div class="flex space-x-2">
                        <Button color="blue" on:click={() => showEditModal = true}>
                            Edit Group
                        </Button>
                        <Button color="red" on:click={handleDelete}>
                            Delete Group
                        </Button>
                    </div>
                {/if}
            </div>
        </Card>

        <div class="grid md:grid-cols-2 gap-8">
            <div class="space-y-8">
                {#if isCreator}
                    <GroupJoinRequests 
                        groupId={group.id}
                        {isCreator}
                        {isMember}
                    />
                {/if}
                <GroupMembership 
                    groupId={group.id}
                    {members}
                    {isCreator}
                />
            </div>
            {#if canViewContent}
                <GroupEvents groupId={group.id} />
            {:else}
                <Card>
                    <p class="text-gray-500">Join the group to view events and other content</p>
                </Card>
            {/if}
        </div>
    {/if}
</div>

<Modal bind:open={showEditModal} size="md">
    <div class="space-y-6">
        <h3 class="text-xl font-medium">Edit Group</h3>
        <form on:submit|preventDefault={handleEdit} class="space-y-4">
            <div>
                <Label for="title">Group Title</Label>
                <Input
                    id="title"
                    bind:value={editForm.title}
                    required
                />
            </div>
            <div>
                <Label for="description">Description</Label>
                <Textarea
                    id="description"
                    bind:value={editForm.description}
                    required
                />
            </div>
            <div class="flex justify-end space-x-2">
                <Button color="alternative" on:click={() => showEditModal = false}>
                    Cancel
                </Button>
                <Button type="submit">Save Changes</Button>
            </div>
        </form>
    </div>
</Modal> 