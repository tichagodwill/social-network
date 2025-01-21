<script lang="ts">
    import { Button, Card, Modal, Label, Input } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { onMount } from 'svelte';

    export let groupId: number;
    export let members: any[] = [];
    export let isCreator: boolean = false;

    let showInviteModal = false;
    let showRemoveModal = false;
    let inviteIdentifier = '';
    let inviteType = 'email';
    let error = '';
    let modalError = '';
    let success = '';
    let loading = false;
    let hasInvitation = false;
    let hasRequest = false;
    let invitationData: any = null;
    let memberToRemove: any = null;

    async function checkInvitationStatus() {
        if (!$auth.isAuthenticated) return;
        
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/invitation/status`, {
                credentials: 'include'
            });
            if (response.ok) {
                const data = await response.json();
                hasInvitation = Boolean(data.hasInvitation);
                hasRequest = Boolean(data.hasRequest);
                invitationData = data.invitation;
                console.log('Invitation status:', { hasInvitation, hasRequest, invitationData });
            }
        } catch (err) {
            console.error('Failed to check invitation status:', err);
        }
    }

    async function inviteMember() {
        modalError = '';
        success = '';
        loading = true;
        
        try {
            if (!inviteIdentifier.trim()) {
                modalError = `Please enter a valid ${inviteType}`;
                return;
            }

            const response = await fetch('http://localhost:8080/groups/invitation', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    groupId: groupId,
                    identifier: inviteIdentifier,
                    identifierType: inviteType
                })
            });

            const data = await response.json();

            if (!response.ok) {
                switch (response.status) {
                    case 404:
                        modalError = `No user found with this ${inviteType}`;
                        break;
                    case 400:
                        if (data.error.includes('already a member')) {
                            modalError = 'This user is already a member of the group';
                        } else if (data.error.includes('pending invitation')) {
                            modalError = 'This user already has a pending invitation';
                        } else {
                            modalError = data.error || 'Invalid invitation request';
                        }
                        break;
                    case 403:
                        modalError = 'You do not have permission to invite members';
                        break;
                    default:
                        modalError = data.error || 'Failed to send invitation';
                }
                return;
            }

            success = 'Invitation sent successfully';
            inviteIdentifier = '';
            showInviteModal = false;
            
            setTimeout(() => {
                success = '';
            }, 3000);
        } catch (err) {
            modalError = err instanceof Error ? err.message : 'Failed to send invitation';
        } finally {
            loading = false;
        }
    }

    async function requestJoin() {
        try {
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
        }
    }

    async function handleInvitation(accept: boolean) {
        try {
            error = '';
            success = '';
            if (!invitationData?.id) {
                error = 'Invalid invitation';
                return;
            }

            const response = await fetch(`http://localhost:8080/groups/invitation/${invitationData.id}/${accept ? 'accept' : 'reject'}`, {
                method: 'POST',
                credentials: 'include'
            });

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || `Failed to ${accept ? 'accept' : 'reject'} invitation`);
            }

            success = data.message;
            setTimeout(() => success = '', 3000);

            // Update local state
            if (accept) {
                window.location.reload(); // Refresh to update membership status
            } else {
                hasInvitation = false;
                invitationData = null;
            }
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to handle invitation';
            console.error('Invitation handling error:', error);
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
        inviteIdentifier = '';
        modalError = '';
    }

    function closeRemoveModal() {
        showRemoveModal = false;
        memberToRemove = null;
    }

    $: isMember = members.some(m => m.id === $auth.user?.id);

    // Check invitation status on mount and when auth state changes
    $: {
        if ($auth.isAuthenticated && !isMember) {
            checkInvitationStatus();
        }
    }

    onMount(() => {
        if ($auth.isAuthenticated && !isMember) {
            checkInvitationStatus();
        }
    });
</script>

<Card>
    <div class="space-y-4">
        <div class="flex justify-between items-center">
            <h3 class="text-xl font-semibold">Members ({members.length})</h3>
            {#if isCreator}
                <Button on:click={() => showInviteModal = true}>Invite Members</Button>
            {:else if !isMember}
                {#if hasInvitation}
                    <div class="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
                        <Button 
                            color="green" 
                            on:click={() => handleInvitation(true)}
                        >
                            Accept Invitation
                        </Button>
                        <Button 
                            color="red" 
                            on:click={() => handleInvitation(false)}
                        >
                            Reject Invitation
                        </Button>
                    </div>
                {:else if hasRequest}
                    <span class="text-sm text-gray-500">Join request pending</span>
                {:else}
                    <Button on:click={requestJoin}>Request to Join</Button>
                {/if}
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

        {#if isMember || isCreator}
            <div class="space-y-2">
                {#if Array.isArray(members) && members.length > 0}
                    {#each members as member}
                        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center p-2 bg-gray-50 dark:bg-gray-800 rounded gap-2">
                            <div>
                                <p class="font-medium">{member.username}</p>
                                <p class="text-sm text-gray-500">{member.role}</p>
                            </div>
                            {#if isCreator && member.id !== $auth.user?.id}
                                <div class="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
                                    <select 
                                        class="text-sm border rounded p-1"
                                        value={member.role}
                                        on:change={(e) => updateMemberRole(member.id, e.target?.value)}
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

<!-- Invite Modal with responsive design -->
<Modal 
    bind:open={showInviteModal} 
    size="md" 
    class="w-full max-w-md mx-auto"
>
    <div class="space-y-6">
        <h3 class="text-xl font-medium">Invite Member</h3>
        
        {#if modalError}
            <div class="p-4 text-red-800 bg-red-100 rounded-lg">
                {modalError}
            </div>
        {/if}

        <form on:submit|preventDefault={inviteMember} class="space-y-4">
            <div>
                <Label>Invite by</Label>
                <div class="flex space-x-4 mb-4">
                    <label class="flex items-center">
                        <input
                            type="radio"
                            bind:group={inviteType}
                            value="email"
                            class="mr-2"
                        />
                        Email
                    </label>
                    <label class="flex items-center">
                        <input
                            type="radio"
                            bind:group={inviteType}
                            value="username"
                            class="mr-2"
                        />
                        Username
                    </label>
                </div>
            </div>
            <div>
                <Label for="identifier">
                    {inviteType === 'email' ? "Member's Email" : "Member's Username"}
                </Label>
                <Input
                    id="identifier"
                    type={inviteType === 'email' ? 'email' : 'text'}
                    bind:value={inviteIdentifier}
                    required
                    placeholder={`Enter member's ${inviteType}`}
                    disabled={loading}
                />
            </div>
            <div class="flex justify-end space-x-2">
                <Button 
                    color="alternative" 
                    on:click={closeModal}
                    disabled={loading}
                >
                    Cancel
                </Button>
                <Button 
                    type="submit"
                    disabled={loading}
                >
                    {loading ? 'Sending...' : 'Send Invitation'}
                </Button>
            </div>
        </form>
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
</style> 