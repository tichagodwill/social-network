<script lang="ts">
    import { Button, Card, Input, Label, Textarea } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { goto } from '$app/navigation';

    let formData = {
        title: '',
        description: '',
        creator_id: $auth.user?.id
    };

    let error = '';

    async function handleSubmit(event: SubmitEvent) {
        event.preventDefault();
        error = '';

        if (!formData.title || !formData.description) {
            error = 'Please fill in all required fields';
            return;
        }

        try {
            const response = await fetch('http://localhost:8080/groups', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify(formData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to create group');
            }

            // Redirect to groups page after successful creation
            goto('/groups');
        } catch (err) {
            console.error('Failed to create group:', err);
            error = err instanceof Error ? err.message : 'Failed to create group';
        }
    }
</script>

<div class="max-w-2xl mx-auto p-4">
    <Card>
        <h1 class="text-2xl font-bold mb-6">Create New Group</h1>

        {#if error}
            <div class="p-4 mb-4 text-red-800 bg-red-100 rounded-lg">
                {error}
            </div>
        {/if}

        <form on:submit={handleSubmit} class="space-y-6">
            <div>
                <Label for="title">Group Title</Label>
                <Input
                    id="title"
                    bind:value={formData.title}
                    required
                    placeholder="Enter group title"
                />
            </div>

            <div>
                <Label for="description">Description</Label>
                <Textarea
                    id="description"
                    bind:value={formData.description}
                    required
                    placeholder="Describe your group"
                    rows={4}
                />
            </div>

            <div class="flex justify-end space-x-2">
                <Button color="alternative" href="/groups">
                    Cancel
                </Button>
                <Button type="submit">
                    Create Group
                </Button>
            </div>
        </form>
    </Card>
</div> 