<script lang="ts">
    import { Button, Card } from 'flowbite-svelte';
    import { onMount } from 'svelte';

    export let groupId: number;
    export let isCreator: boolean = false;
    export let isMember: boolean = false;

    let requests: any[] = [];
    let error = '';
    let success = '';
    let loading = true;

    async function loadPendingRequests() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/requests`, {
                credentials: 'include'
            });
            
            if (response.ok) {
                requests = await response.json() || [];
            } else {
                requests = [];
                if (response.status !== 404) {  // Don't show error for 404
                    const data = await response.json();
                    throw new Error(data.error || 'Failed to load requests');
                }
            }
        } catch (err) {
            console.error('Failed to load requests:', err);
            error = err instanceof Error ? err.message : 'Failed to load requests';
            requests = [];
        } finally {
            loading = false;
        }
    }

    async function handleRequest(requestId: number, action: 'accept' | 'reject') {
        try {
            const response = await fetch(`http://localhost:8080/groups/${action}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({ requestId })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || `Failed to ${action} request`);
            }

            success = `Request ${action}ed successfully`;
            await loadPendingRequests();
        } catch (err) {
            error = err instanceof Error ? err.message : `Failed to ${action} request`;
        }
    }

    onMount(loadPendingRequests);

    $: showCard = isCreator && requests.length > 0;
</script>

{#if showCard}
    <Card>
        <div class="space-y-4">
            <h3 class="text-xl font-semibold">Pending Join Requests</h3>

            {#if error}
                <div class="p-4 text-red-800 bg-red-100 rounded-lg">
                    {error}
                </div>
            {/if}

            {#if success}
                <div class="p-4 text-green-800 bg-green-100 rounded-lg">
                    {success}
                </div>
            {/if}

            {#if loading}
                <div class="text-center py-4">
                    <div class="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
                </div>
            {:else if !requests || requests.length === 0}
                <p class="text-gray-500">No pending requests</p>
            {:else}
                <div class="space-y-2">
                    {#each requests as request}
                        <div class="flex justify-between items-center p-2 bg-gray-50 dark:bg-gray-800 rounded">
                            <div>
                                <p class="font-medium">{request.username}</p>
                                <p class="text-sm text-gray-500">
                                    Requested {new Date(request.createdAt).toLocaleDateString()}
                                </p>
                            </div>
                            {#if isCreator}
                                <div class="flex space-x-2">
                                    <Button 
                                        size="xs" 
                                        color="green"
                                        on:click={() => handleRequest(request.id, 'accept')}
                                    >
                                        Accept
                                    </Button>
                                    <Button 
                                        size="xs" 
                                        color="red"
                                        on:click={() => handleRequest(request.id, 'reject')}
                                    >
                                        Reject
                                    </Button>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </Card>
{/if} 