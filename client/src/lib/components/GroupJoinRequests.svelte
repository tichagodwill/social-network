<script lang="ts">
    import { Button, Card } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { handleJoinRequest } from '$lib/api/groupApi';

    export let groupId: number;
    export let isCreator: boolean;
    export let isMember: boolean;
    export let role: string;

    let requests: any[] = [];
    let loading = false;

    async function fetchRequests() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/requests`, {
                credentials: 'include'
            });
            if (response.ok) {
                requests = await response.json();
            }
        } catch (error) {
            console.error('Error fetching requests:', error);
        }
    }

    async function handleRequest(requestId: number, action: 'accept' | 'reject') {
        try {
            loading = true;
            await handleJoinRequest(groupId, requestId, action);
            // Remove the handled request from the list
            requests = requests.filter(r => r.id !== requestId);
        } catch (error) {
            console.error(`Error ${action}ing request:`, error);
        } finally {
            loading = false;
        }
    }

    $: if (groupId && (isCreator || role === 'admin')) {
        fetchRequests();
    }
</script>

{#if isCreator || role === 'admin'}
    <Card class="mb-4">
        <div class="space-y-4">
            <h3 class="text-xl font-semibold">Pending Join Requests</h3>
            
            {#if requests.length === 0}
                <p class="text-gray-500">No pending requests</p>
            {:else}
                <div class="space-y-2">
                    {#each requests as request}
                        <div class="flex justify-between items-center p-2 bg-gray-50 dark:bg-gray-800 rounded">
                            <div>
                                <p class="font-medium">{request.username}</p>
                                <p class="text-sm text-gray-500">
                                    Requested {new Date(request.created_at).toLocaleDateString()}
                                </p>
                            </div>
                            <div class="flex space-x-2">
                                <Button
                                    size="xs"
                                    color="green"
                                    disabled={loading}
                                    on:click={() => handleRequest(request.id, 'accept')}
                                >
                                    Accept
                                </Button>
                                <Button
                                    size="xs"
                                    color="red"
                                    disabled={loading}
                                    on:click={() => handleRequest(request.id, 'reject')}
                                >
                                    Reject
                                </Button>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </Card>
{/if} 