<script lang="ts">
    import { Modal, Button, Label, Input, Textarea } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import type { Group } from '$lib/types';

    export let open = false;
    export let onClose = () => {};
    export let onGroupCreated = () => {};

    let title = '';
    let description = '';
    let error = '';
    let loading = false;

    async function handleSubmit() {
        if (!title || !description) {
            error = 'Please fill in all required fields';
            return;
        }

        loading = true;
        error = '';

        try {
            const response = await fetch('http://localhost:8080/groups', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    title,
                    description,
                    creator_id: $auth.user?.id
                })
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText);
            }

            // Reset form
            title = '';
            description = '';
            
            // Close modal and notify parent
            onClose();
            onGroupCreated();
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to create group';
            console.error('Failed to create group:', err);
        } finally {
            loading = false;
        }
    }

    function handleClose() {
        title = '';
        description = '';
        error = '';
        onClose();
    }
</script>

<Modal bind:open size="md" class="w-full max-w-2xl">
    <div class="space-y-6">
        <h3 class="text-xl font-medium text-gray-900 dark:text-white">Create New Group</h3>
        
        {#if error}
            <div class="p-4 mb-4 text-sm text-red-800 bg-red-100 rounded-lg dark:bg-red-900 dark:text-red-400">
                {error}
            </div>
        {/if}

        <form on:submit|preventDefault={handleSubmit} class="space-y-6">
            <div>
                <Label for="title" class="mb-2">Group Title</Label>
                <Input
                    type="text"
                    id="title"
                    placeholder="Enter group title"
                    required
                    bind:value={title}
                />
            </div>
            <div>
                <Label for="description" class="mb-2">Description</Label>
                <Textarea
                    id="description"
                    placeholder="Describe your group..."
                    required
                    bind:value={description}
                    rows={4}
                />
            </div>
            <div class="flex justify-end space-x-2">
                <Button color="alternative" on:click={handleClose}>
                    Cancel
                </Button>
                <Button type="submit" disabled={loading}>
                    {loading ? 'Creating...' : 'Create Group'}
                </Button>
            </div>
        </form>
    </div>
</Modal> 