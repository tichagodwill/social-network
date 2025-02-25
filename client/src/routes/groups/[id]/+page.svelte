<script lang="ts">
    import { auth } from '$lib/stores/auth';
    import { Button, Card, Modal, Label, Input, Textarea } from 'flowbite-svelte';
    import { goto } from '$app/navigation';
    import GroupMembership from '$lib/components/GroupMembership.svelte';
    import GroupEvents from '$lib/components/GroupEvents.svelte';
    import GroupJoinRequests from '$lib/components/GroupJoinRequests.svelte';
    import GroupPosts from '$lib/components/GroupPosts.svelte';
    import { onMount } from 'svelte';
    import { groups } from '$lib/stores/groups';
    import type { GroupMember } from '$lib/types';
    import { getInvitationStatus } from '$lib/stores/groups';

    export let data;

    let group = data?.group;
    let members: GroupMember[] = data?.members || [];
    let error = '';
    let groupId = group?.id;
    let authChecked = false;
    let userRole: string = '';
    
    function isCreator(): boolean {
        return $auth.user && group ? group.creator_id === $auth.user.id : false;
    }

    function isMember(): boolean {
        return $auth.user && Array.isArray(members) ? members.some(m => m.id === $auth.user?.id) : false;
    }

    let showEditModal = false;
    let showDeleteModal = false;
    let loading = true;

    let editForm = {
        title: group?.title || '',
        description: group?.description || ''
    };

    $: canViewContent = isCreator() || isMember();
    $: groupId = group?.id;

    let invitationStatus: { status: string; invitationId: number | null } = {
        status: '',
        invitationId: null
    };

    // Watch for auth changes
    $: {
        if ($auth.isAuthenticated !== undefined) {
            authChecked = true;
            loading = false;
        }
    }

    onMount(async () => {
        // Wait for auth to be checked
        const checkAuth = new Promise<void>((resolve) => {
            if ($auth.isAuthenticated !== undefined) {
                resolve();
            } else {
                const unsubscribe = auth.subscribe(value => {
                    if (value.isAuthenticated !== undefined) {
                        unsubscribe();
                        resolve();
                    }
                });
            }
        });

        await checkAuth;
        authChecked = true;
        loading = false;

        if ($auth.isAuthenticated && groupId) {
            try {
                const status = await getInvitationStatus(groupId);
                invitationStatus = status;
            } catch (error) {
                console.error('Error fetching invitation status:', error);
            }
        }
    });

    async function handleEdit(event: Event) {
        event.preventDefault();
        error = '';
        
        try {
            if (!groupId) throw new Error('Group ID is missing');
            
            const result = await groups.updateGroup(groupId, {
                title: editForm.title.trim(),
                description: editForm.description.trim()
            });
            
            group = {
                ...group,
                title: editForm.title.trim(),
                description: editForm.description.trim()
            };
            
            showEditModal = false;
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to update group';
            console.error('Update error:', err);
        }
    }

    async function handleDelete() {
        error = '';
        
        try {
            if (!groupId) throw new Error('Group ID is missing');
            
            await groups.deleteGroup(groupId);
            showDeleteModal = false;
            goto('/groups');
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to delete group';
            console.error('Delete error:', err);
        }
    }

    function hasAdminPrivileges(): boolean {
        if (!$auth.user || !group) return false;
        const currentMember = members.find(m => m.id === $auth.user.id);
        return isCreator() || (currentMember?.role === 'admin');
    }

    // Add animation classes for transitions
    let fadeIn = "animate-fade-in";
    let slideIn = "animate-slide-in";

    // Function to get user's role in the group
    async function getUserRole() {
        try {
            if (!$auth.user || !groupId) return null;

            const response = await fetch(`http://localhost:8080/groups/${groupId}/role`, {
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to get user role');
            }

            const data = await response.json();
            return data.role;
        } catch (error) {
            console.error('Error getting user role:', error);
            return null;
        }
    }

    $: if (groupId && $auth.isAuthenticated) {
        getUserRole();
    }

    async function handleInvitation(action: 'accept' | 'reject') {
        if (!invitationStatus.invitationId) return;
        
        try {
            const response = await fetch(
                `http://localhost:8080/groups/${groupId}/invitations/${invitationStatus.invitationId}/${action}`,
                {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                }
            );

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `Failed to ${action} invitation`);
            }

            // Refresh the page or update the UI
            window.location.reload();
        } catch (error) {
            console.error(`Error ${action}ing invitation:`, error);
        }
    }
</script>

<style>
    @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
    }

    @keyframes slideIn {
        from { transform: translateY(20px); opacity: 0; }
        to { transform: translateY(0); opacity: 1; }
    }

    @keyframes pulse {
        0% { transform: scale(1); }
        50% { transform: scale(1.05); }
        100% { transform: scale(1); }
    }

    :global(.animate-fade-in) {
        animation: fadeIn 0.5s ease-out forwards;
    }

    :global(.animate-slide-in) {
        animation: slideIn 0.5s ease-out forwards;
    }

    :global(.group-card) {
        @apply transition-all duration-300 
               bg-gradient-to-br from-white via-gray-50 to-gray-100
               dark:from-gray-800 dark:via-gray-800 dark:to-gray-900
               w-full max-w-none mx-auto mb-8 mt-8
               border border-gray-200 dark:border-gray-700
               shadow-lg hover:shadow-xl
               backdrop-blur-sm
               relative
               overflow-hidden;
    }

    :global(.group-title) {
        @apply text-4xl font-bold bg-clip-text text-transparent 
               bg-gradient-to-r from-blue-600 via-purple-500 to-indigo-600 
               dark:from-blue-400 dark:via-purple-300 dark:to-indigo-400 
               break-words mb-4 
               relative z-10;
    }

    :global(.group-description) {
        @apply text-gray-700 dark:text-gray-300 
               leading-relaxed text-lg
               whitespace-pre-wrap break-words 
               max-w-none mx-auto
               relative z-10
               bg-white/50 dark:bg-gray-800/50
               rounded-lg p-6
               backdrop-blur-sm
               border border-gray-100 dark:border-gray-700;
    }

    :global(.action-button) {
        @apply transform transition-all duration-300 
               hover:scale-105 hover:shadow-md
               relative z-10;
    }

    /* Add decorative elements */
    :global(.group-card::before) {
        content: '';
        @apply absolute top-0 right-0 w-64 h-64
               bg-gradient-to-br from-blue-500/10 to-purple-500/10
               dark:from-blue-400/5 dark:to-purple-400/5
               rounded-full -translate-y-32 translate-x-32
               blur-3xl;
    }

    :global(.group-card::after) {
        content: '';
        @apply absolute bottom-0 left-0 w-64 h-64
               bg-gradient-to-tr from-indigo-500/10 to-cyan-500/10
               dark:from-indigo-400/5 dark:to-cyan-400/5
               rounded-full translate-y-32 -translate-x-32
               blur-3xl;
    }
</style>

{#if error}
    <div class="p-4 mb-4 text-red-800 bg-red-100 rounded-lg">
        {error}
    </div>
{/if}

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    {#if loading}
        <div class="text-center p-8 {fadeIn}">
            <div class="loading-spinner mx-auto"></div>
            <p class="mt-4 text-lg text-gray-600 dark:text-gray-300">Loading group data...</p>
        </div>
    {:else if !authChecked}
        <div class="text-center p-8 {fadeIn}">
            <div class="loading-spinner mx-auto"></div>
            <p class="mt-4 text-lg text-gray-600 dark:text-gray-300">Checking authentication...</p>
        </div>
    {:else if !$auth.isAuthenticated}
        <div class="text-center p-8 {slideIn}">
            <p class="text-xl mb-4 text-gray-700 dark:text-gray-200">Please log in to view this group</p>
            <Button href="/login" class="action-button" size="xl" gradient color="purpleToBlue">
                Log In
            </Button>
        </div>
    {:else if !group}
        <div class="text-center p-8 {slideIn}">
            <p class="text-xl mb-4 text-gray-700 dark:text-gray-200">Group not found</p>
            <Button href="/groups" class="action-button" size="xl" gradient color="purpleToBlue">
                Back to Groups
            </Button>
        </div>
    {:else if invitationStatus.status === 'pending'}
        <div class="container mx-auto px-4 py-8">
            <Card class="max-w-4xl mx-auto">
                <h1 class="text-2xl font-bold mb-4">{group.title}</h1>
                <p class="text-gray-600 dark:text-gray-300">{group.description}</p>
                <div class="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg mt-4">
                    <p class="text-lg font-medium mb-3">You have a pending invitation to this group</p>
                    <div class="flex space-x-3">
                        <Button 
                            color="green" 
                            size="sm"
                            on:click={() => handleInvitation('accept')}
                        >
                            Accept Invitation
                        </Button>
                        <Button 
                            color="red" 
                            size="sm"
                            on:click={() => handleInvitation('reject')}
                        >
                            Reject Invitation
                        </Button>
                    </div>
                </div>
            </Card>
        </div>
    {:else}
        <Card class="group-card {slideIn} w-full max-w-none mb-8 mt-8">
            <div class="space-y-6 max-w-7xl mx-auto p-8">
                <h1 class="group-title break-words">{group.title}</h1>
                <p class="group-description whitespace-pre-wrap break-words">{group.description}</p>
                
                {#if isCreator()}
                    <div class="flex flex-wrap gap-4 mt-8">
                        <Button 
                            gradient
                            color="blue" 
                            class="action-button bg-gradient-to-r from-blue-500 via-indigo-500 to-purple-500 
                                   hover:from-blue-600 hover:via-indigo-600 hover:to-purple-600
                                   shadow-md hover:shadow-xl" 
                            on:click={() => {
                                editForm = {
                                    title: group.title,
                                    description: group.description
                                };
                                showEditModal = true;
                            }}
                        >
                            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                            </svg>
                            Edit Group
                        </Button>
                        <Button 
                            gradient
                            color="red"
                            class="action-button bg-gradient-to-r from-red-600 to-pink-600 hover:from-red-700 hover:to-pink-700" 
                            on:click={() => showDeleteModal = true}
                        >
                            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                            </svg>
                            Delete Group
                        </Button>
                    </div>
                {/if}
            </div>
        </Card>

        <div class="grid md:grid-cols-2 gap-8">
            <div class="space-y-8 {slideIn}" style="animation-delay: 0.2s">
                {#if hasAdminPrivileges()}
                    <GroupJoinRequests 
                        {groupId}
                        isCreator={group.creator_id === $auth.user?.id}
                        isMember={group.is_member}
                        role={group.role}
                    />
                {/if}
                <GroupMembership 
                    {groupId}
                    {members}
                    isCreator={isCreator()}
                />
            </div>
            {#if canViewContent}
                <div class="space-y-8 {slideIn}" style="animation-delay: 0.4s">
                    <GroupPosts groupId={group.id} />
                    <GroupEvents groupId={group.id} />
                </div>
            {:else}
                <Card class="content-section flex items-center justify-center {slideIn}" style="animation-delay: 0.4s">
                    <p class="text-lg text-gray-500 dark:text-gray-400">
                        Join the group to view posts, events and other content
                    </p>
                </Card>
            {/if}
        </div>
    {/if}
</div>

<!-- Enhanced Modal Styles -->
<Modal 
    bind:open={showEditModal} 
    size="md"
    class="transform transition-all duration-300"
>
    <div class="space-y-6 p-2">
        <h3 class="text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-cyan-500 to-blue-500">
            Edit Group
        </h3>
        <form on:submit|preventDefault={handleEdit} class="space-y-4">
            <div class="space-y-2">
                <Label for="title" class="text-lg">Group Title</Label>
                <Input
                    id="title"
                    bind:value={editForm.title}
                    required
                    class="transition-all duration-300 focus:ring-2 focus:ring-blue-500"
                />
            </div>
            <div class="space-y-2">
                <Label for="description" class="text-lg">Description</Label>
                <Textarea
                    id="description"
                    bind:value={editForm.description}
                    required
                    class="transition-all duration-300 focus:ring-2 focus:ring-blue-500"
                />
            </div>
            <div class="flex justify-end space-x-3">
                <Button 
                    type="button"
                    color="alternative" 
                    class="action-button" 
                    on:click={() => showEditModal = false}
                >
                    Cancel
                </Button>
                <Button 
                    type="submit"
                    gradient
                    color="blue"
                    class="action-button bg-gradient-to-r from-cyan-500 to-blue-500 hover:from-cyan-600 hover:to-blue-600"
                >
                    Save Changes
                </Button>
            </div>
        </form>
    </div>
</Modal>

<Modal 
    bind:open={showDeleteModal} 
    size="md"
    class="transform transition-all duration-300"
>
    <div class="space-y-6 p-2">
        <h3 class="text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-red-600 to-pink-600">
            Delete Group
        </h3>
        <p class="text-lg text-gray-600 dark:text-gray-300">
            Are you sure you want to delete this group? This action cannot be undone and all group data will be permanently removed.
        </p>
        <div class="flex justify-end space-x-3">
            <Button 
                color="alternative" 
                class="action-button"
                on:click={() => showDeleteModal = false}
            >
                Cancel
            </Button>
            <Button 
                gradient
                color="red"
                class="action-button bg-gradient-to-r from-red-600 to-pink-600 hover:from-red-700 hover:to-pink-700"
                on:click={handleDelete}
            >
                Delete Group
            </Button>
        </div>
    </div>
</Modal> 