<script lang="ts">
    import { Button, Card } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { handleJoinRequest } from '$lib/api/groupApi';
    import { createEventDispatcher } from 'svelte';

    export let groupId: number;
    export let isCreator: boolean;
    export let isMember: boolean;
    export let role: string;

    let requests: any[] = [];
    let loading = false;
    let error = '';
    let success = '';

    const dispatch = createEventDispatcher();

    async function fetchRequests() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/join-requests`, {
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
            error = '';
            const response = await fetch(`http://localhost:8080/groups/${groupId}/join-requests/${action}`, {
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

            // Remove the handled request from the list
            requests = requests.filter(r => r.id !== requestId);
            success = `Request ${action}ed successfully`;
            
            // Notify parent component to refresh members list if request was accepted
            if (action === 'accept') {
                dispatch('memberAdded');
            }

            setTimeout(() => {
                success = '';
            }, 3000);
        } catch (err) {
            error = err instanceof Error ? err.message : `Failed to ${action} request`;
            console.error(`Error ${action}ing request:`, err);
        } finally {
            loading = false;
        }
    }

    // Fetch requests when component mounts and when role/isCreator changes
    $: if (groupId && (isCreator || role === 'admin')) {
        fetchRequests();
    }
</script>

{#if isCreator || role === 'admin'}
    <Card class="mb-4">
        <div class="space-y-4">
            <h3 class="text-xl font-semibold flex items-center justify-between">
                Pending Join Requests
                {#if requests.length > 0}
                    <span class="bg-blue-100 text-blue-800 text-xs font-medium px-2.5 py-0.5 rounded-full dark:bg-blue-900 dark:text-blue-300">
                        {requests.length}
                    </span>
                {/if}
            </h3>
            
            {#if error}
                <div class="p-4 mb-4 text-sm text-red-800 bg-red-100 rounded-lg dark:bg-red-900 dark:text-red-400" transition:fade>
                    {error}
                </div>
            {/if}

            {#if success}
                <div class="p-4 mb-4 text-sm text-green-800 bg-green-100 rounded-lg dark:bg-green-900 dark:text-green-400" transition:fade>
                    {success}
                </div>
            {/if}
            
            {#if requests.length === 0}
                <p class="text-gray-500 dark:text-gray-400">No pending requests</p>
            {:else}
                <div class="space-y-3">
                    {#each requests as request}
                        <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md transition-shadow duration-200">
                            <div>
                                <p class="font-medium text-gray-900 dark:text-white">
                                    {request.username}
                                </p>
                                <p class="text-sm text-gray-500 dark:text-gray-400">
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
                                    <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                                    </svg>
                                    {loading ? 'Processing...' : 'Accept'}
                                </Button>
                                <Button
                                    size="xs"
                                    color="red"
                                    disabled={loading}
                                    on:click={() => handleRequest(request.id, 'reject')}
                                >
                                    <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                                    </svg>
                                    {loading ? 'Processing...' : 'Reject'}
                                </Button>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </Card>
{/if}

<style>
    /* Add smooth transitions for hover effects */
    :global(.btn-transition) {
        @apply transform transition-all duration-200;
    }

    :global(.btn-transition:hover) {
        @apply -translate-y-0.5;
    }
</style> 