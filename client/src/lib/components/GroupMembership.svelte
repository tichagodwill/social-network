<script lang="ts">
    import { Button, Card, Modal, Label, Input } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { onMount } from 'svelte';
    import { groups } from '$lib/stores/groups';
    import type { GroupMember } from '$lib/types';
    import { inviteToGroup } from '$lib/api/groupApi';
    import { createEventDispatcher } from 'svelte';

    export let groupId: number;
    export let members: GroupMember[] = [];
    export let isCreator: boolean = false;

    let showInviteModal = false;
    let showRemoveModal = false;
    let inviteUsername = '';
    let error = '';
    let modalError = '';
    let success = '';
    let loading = false;
    let hasInvitation = false;
    let hasRequest = false;
    let invitationData: any = null;
    let memberToRemove: any = null;
    let currentUserRole = '';
    let wasRemoved = false;

    const dispatch = createEventDispatcher();

    // Update currentUserRole when members or auth changes
    $: {
        if ($auth?.user && Array.isArray(members) && members.length > 0) {
            const currentMember = members.find(m => m.id === $auth.user?.id);
            currentUserRole = currentMember?.role || '';
        }
    }

    // Function to check if current user is admin or creator
    function hasAdminPrivileges(): boolean {
        if (!$auth?.user || !Array.isArray(members)) return false;
        const currentMember = members.find(m => m.id === $auth.user?.id);
        return isCreator || (currentMember?.role === 'admin');
    }

    // Function to check if current user is a member
    function isMember(): boolean {
        if (!$auth?.user || !Array.isArray(members)) return false;
        return members.some(m => m.id === $auth.user?.id);
    }

    async function checkInvitationStatus() {
        if (!$auth.isAuthenticated) return;
        
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/invitations/status`, {
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                hasInvitation = data.hasInvitation;
                hasRequest = data.hasRequest;
                if (data.invitation) {
                    invitationData = data.invitation;
                }
            }
        } catch (err) {
            console.error('Failed to check invitation status:', err);
        }
    }

    async function handleInvite() {
        modalError = '';
        success = '';
        loading = true;
        
        try {
            if (!inviteUsername.trim()) {
                modalError = 'Please enter a username';
                return;
            }

            await inviteToGroup(groupId, inviteUsername.trim());
            success = 'Invitation sent successfully';
            inviteUsername = '';
            showInviteModal = false;
            
            // Notify parent component
            dispatch('memberInvited');
        } catch (err) {
            modalError = err instanceof Error ? err.message : 'Failed to send invitation';
        } finally {
            loading = false;
        }
    }

    async function requestJoin() {
        try {
            loading = true;
            error = '';
            success = '';
            const response = await fetch(`http://localhost:8080/groups/${groupId}/join`, {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Failed to request join');
            }

            hasRequest = true;
            success = 'Join request sent successfully';
            setTimeout(() => success = '', 3000);
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to request join';
            console.error('Join request error:', error);
        } finally {
            loading = false;
        }
    }

    async function handleInvitation(accept: boolean) {
        try {
            loading = true;
            error = '';
            success = '';

            const response = await fetch(
                `http://localhost:8080/groups/${groupId}/invitations/${invitationData.invitation_id}/${accept ? 'accept' : 'reject'}`, 
                {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                }
            );

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || `Failed to ${accept ? 'accept' : 'reject'} invitation`);
            }

            hasInvitation = false;
            invitationData = null;

            if (accept) {
                success = 'Successfully joined the group!';
                setTimeout(() => window.location.reload(), 1500);
            } else {
                success = 'Invitation rejected';
            }
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to handle invitation';
            console.error('Invitation handling error:', error);
        } finally {
            loading = false;
        }
    }

    async function updateMemberRole(memberId: number, newRole: string) {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/members/${memberId}/role`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({ role: newRole })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to update role');
            }

            members = members.map(member => 
                member.id === memberId 
                    ? { ...member, role: newRole }
                    : member
            );
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to update role';
        }
    }

    async function handleRemoveMember(member: any) {
        memberToRemove = member;
        showRemoveModal = true;
    }

    async function confirmRemoveMember() {
        if (!memberToRemove) return;

        try {
            loading = true;
            const response = await fetch(`http://localhost:8080/groups/${groupId}/members/${memberToRemove.id}`, {
                method: 'DELETE',
                credentials: 'include'
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to remove member');
            }

            // Remove member from local state
            members = members.filter(m => m.id !== memberToRemove.id);
            success = `${memberToRemove.username} has been removed from the group`;
            showRemoveModal = false;
            memberToRemove = null;
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to remove member';
        } finally {
            loading = false;
        }
    }

    function closeModal() {
        showInviteModal = false;
        inviteUsername = '';
        modalError = '';
    }

    function closeRemoveModal() {
        showRemoveModal = false;
        memberToRemove = null;
    }

    // Check invitation status on mount and when auth state changes
    $: {
        if ($auth.isAuthenticated && !isMember()) {
            checkInvitationStatus();
        }
    }

    async function checkMembershipStatus() {
        if (!$auth.isAuthenticated) return;
        
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/members/status`, {
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                wasRemoved = data.wasRemoved;
            }
        } catch (err) {
            console.error('Failed to check membership status:', err);
        }
    }

    onMount(() => {
        if ($auth.isAuthenticated) {
            checkMembershipStatus();
        }
    });
</script>

<Card>
    <div class="space-y-4">
        {#if hasRequest}
            <div class="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg border border-yellow-200 dark:border-yellow-800">
                <div class="flex justify-between items-center">
                    <p class="text-yellow-800 dark:text-yellow-200">
                        Your request to join this group is pending approval
                    </p>
                </div>
            </div>
        {/if}

        {#if wasRemoved && !isMember() && !hasRequest}
            <div class="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg border border-yellow-200 dark:border-yellow-800">
                <div class="flex flex-col sm:flex-row justify-between items-center gap-4">
                    <p class="text-yellow-800 dark:text-yellow-200">
                        You were previously removed from this group
                    </p>
                    <Button 
                        color="blue"
                        size="sm"
                        on:click={requestJoin}
                        disabled={loading}
                    >
                        Request to Join Again
                    </Button>
                </div>
            </div>
        {/if}

        <div class="flex justify-between items-center">
            <h3 class="text-xl font-semibold">Members ({members.length})</h3>
            {#if hasAdminPrivileges()}
                <Button color="blue" on:click={() => showInviteModal = true}>
                    Invite Member
                </Button>
            {:else if !isMember() && !hasRequest}
                <Button 
                    on:click={requestJoin}
                    disabled={loading}
                >
                    {loading ? 'Sending Request...' : 'Request to Join'}
                </Button>
            {/if}
        </div>

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

        {#if isMember() || isCreator}
            <div class="space-y-2">
                {#if Array.isArray(members) && members.length > 0}
                    {#each members as member}
                        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center p-2 bg-gray-50 dark:bg-gray-800 rounded gap-2">
                            <div>
                                <p class="username username-{member.role.toLowerCase()} flex items-center gap-2">
                                    {member.username}
                                    {#if member.role === 'creator'}
                                        <svg class="w-4 h-4 text-purple-500" fill="currentColor" viewBox="0 0 20 20">
                                            <path d="M10 2a1 1 0 0 1 .78.375l2.143 2.5a1 1 0 0 0 .672.3H16a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V7.176a2 2 0 0 1 2-2h2.405a1 1 0 0 0 .672-.3L9.22 2.375A1 1 0 0 1 10 2z"/>
                                        </svg>
                                    {/if}
                                </p>
                                <span class="role-badge role-{member.role.toLowerCase()}">
                                    {member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                                </span>
                            </div>
                            {#if isCreator && member.id !== $auth.user?.id}
                                <div class="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
                                    <select 
                                        class="text-sm border rounded p-1 bg-white dark:bg-gray-700"
                                        value={member.role}
                                        on:change={e => updateMemberRole(member.id, e.currentTarget.value)}
                                    >
                                        <option value="member">Member</option>
                                        <option value="moderator">Moderator</option>
                                        <option value="admin">Admin</option>
                                    </select>
                                    <Button 
                                        size="xs" 
                                        color="red"
                                        class="w-full sm:w-auto"
                                        on:click={() => handleRemoveMember(member)}
                                    >
                                        Remove
                                    </Button>
                                </div>
                            {/if}
                        </div>
                    {/each}
                {:else}
                    <p class="text-gray-500">No members found</p>
                {/if}
            </div>
        {:else}
            <p class="text-gray-500">Join the group to see members</p>
        {/if}
    </div>
</Card>

<!-- Invite Modal -->
<Modal bind:open={showInviteModal} size="xs" autoclose={false}>
    <div class="space-y-4">
        <h3 class="text-xl font-medium">Invite New Member</h3>
        
        {#if modalError}
            <div class="p-4 text-red-800 bg-red-100 rounded-lg text-sm">
                {modalError}
            </div>
        {/if}

        <div>
            <Label for="username">Username</Label>
            <Input
                id="username"
                type="text"
                placeholder="Enter username"
                bind:value={inviteUsername}
            />
        </div>

        <div class="flex justify-end space-x-2">
            <Button color="light" on:click={() => {
                showInviteModal = false;
                modalError = '';
                inviteUsername = '';
            }}>
                Cancel
            </Button>
            <Button 
                color="blue"
                disabled={loading || !inviteUsername.trim()}
                on:click={handleInvite}
            >
                {loading ? 'Sending...' : 'Send Invitation'}
            </Button>
        </div>
    </div>
</Modal>

<!-- Remove member confirmation modal -->
<Modal bind:open={showRemoveModal} size="sm">
    <div class="text-center">
        <h3 class="mb-4 text-lg font-medium">Remove Member</h3>
        
        {#if memberToRemove}
            <p class="mb-6 text-gray-700 dark:text-gray-300">
                Are you sure you want to remove {memberToRemove.username} from the group?
            </p>
            
            <div class="flex justify-end space-x-2">
                <Button 
                    color="alternative" 
                    on:click={closeRemoveModal}
                    disabled={loading}
                >
                    Cancel
                </Button>
                <Button 
                    color="red"
                    on:click={confirmRemoveMember}
                    disabled={loading}
                >
                    {loading ? 'Removing...' : 'Remove Member'}
                </Button>
            </div>
        {/if}
    </div>
</Modal>

{#if success}
    <div class="fixed bottom-4 right-4 p-4 bg-green-100 text-green-800 rounded-lg shadow-lg z-50 animate-fade-out">
        {success}
    </div>
{/if}

<style>
    @keyframes fadeOut {
        from { opacity: 1; }
        to { opacity: 0; }
    }

    .animate-fade-out {
        animation: fadeOut 0.5s ease-out 2.5s forwards;
    }

    :global(.modal-content) {
        max-width: 90vw;
        margin: 0 auto;
    }

    @media (min-width: 640px) {
        :global(.modal-content) {
            max-width: 28rem;
        }
    }

    .username {
        @apply font-medium;
    }

    .username-creator {
        @apply text-purple-700 dark:text-purple-300 font-bold;
    }

    .username-admin {
        @apply text-blue-700 dark:text-blue-300 font-semibold;
    }

    .username-moderator {
        @apply text-green-700 dark:text-green-300;
    }

    .username-member {
        @apply text-gray-700 dark:text-gray-300;
    }

    .role-badge {
        @apply text-xs font-semibold px-2.5 py-0.5 rounded-full;
    }

    .role-creator {
        @apply bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300;
    }

    .role-admin {
        @apply bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300;
    }

    .role-member {
        @apply bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300;
    }

    .role-moderator {
        @apply bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300;
    }
</style> 