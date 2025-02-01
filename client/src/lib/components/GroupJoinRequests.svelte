<script lang="ts">
    import { Button, Card } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { handleJoinRequest } from '$lib/api/groupApi';

    export let groupId: number;
    export let isCreator: boolean;
    export let isMember: boolean;
    export let role: string;

    interface JoinRequest {
        id: number;
        username: string;
        createdAt: string;
    }

    let requests: JoinRequest[] = [];
    let error = '';
    let success = '';
    let loading = true;

    // Check if user has admin privileges
    $: hasAdminPrivileges = isCreator || role === 'admin';

    async function loadRequests() {
        if (!groupId || !hasAdminPrivileges) return;
        
        try {
            loading = true;
            const response = await fetch(`http://localhost:8080/groups/${groupId}/join-requests`, {
                credentials: 'include'
            });
            
            if (response.ok) {
                requests = await response.json();
            } else {
                const errorData = await response.json();
                error = errorData.error || 'Failed to load requests';
                requests = [];
            }
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to load requests';
            requests = [];
        } finally {
            loading = false;
        }
    }

    // Load requests when component mounts and when auth/hasAdminPrivileges changes
    $: if ($auth.isAuthenticated && hasAdminPrivileges && groupId) {
        loadRequests();
    }

    // Only show card if we have admin privileges and there are requests
    $: showCard = hasAdminPrivileges && Array.isArray(requests) && requests.length > 0;

    async function handleRequest(requestId: number, action: 'accept' | 'reject') {
        try {
            error = '';
            success = '';
            await handleJoinRequest(groupId, requestId, action);
            success = `Request ${action}ed successfully`;
            await loadRequests();
        } catch (err) {
            error = err instanceof Error ? err.message : `Failed to ${action} request`;
        }
    }
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
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </Card>
{/if} 