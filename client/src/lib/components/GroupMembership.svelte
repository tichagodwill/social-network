<script lang="ts">
    import { Button, Card, Modal, Label, Input } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';

    export let groupId: number;
    export let members: any[] = [];
    export let isCreator: boolean = false;

    let showInviteModal = false;
    let inviteEmail = '';
    let error = '';
    let success = '';

    async function inviteMember() {
        error = '';
        success = '';
        
        try {
            // First, find user by email
            const userResponse = await fetch(`http://localhost:8080/user/by-email/${inviteEmail}`, {
                credentials: 'include'
            });

            if (!userResponse.ok) {
                throw new Error('User not found');
            }

            const user = await userResponse.json();

            // Send invitation
            const response = await fetch('http://localhost:8080/groups/invitation', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    groupId: groupId,
                    inviteeId: user.id
                })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to send invitation');
            }

            success = 'Invitation sent successfully';
            inviteEmail = '';
            showInviteModal = false;
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to send invitation';
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

            // Update local member data
            members = members.map(member => 
                member.id === memberId 
                    ? { ...member, role: newRole }
                    : member
            );
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to update role';
        }
    }

    async function removeMember(memberId: number) {
        if (!confirm('Are you sure you want to remove this member?')) {
            return;
        }

        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/members/${memberId}`, {
                method: 'DELETE',
                credentials: 'include'
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to remove member');
            }

            // Update local member list
            members = members.filter(member => member.id !== memberId);
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to remove member';
        }
    }
</script>

<Card>
    <div class="space-y-4">
        <div class="flex justify-between items-center">
            <h3 class="text-xl font-semibold">Members</h3>
            {#if isCreator}
                <Button on:click={() => showInviteModal = true}>Invite Member</Button>
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

        <div class="space-y-2">
            {#if Array.isArray(members) && members.length > 0}
                {#each members as member}
                    <div class="flex justify-between items-center p-2 bg-gray-50 dark:bg-gray-800 rounded">
                        <div>
                            <p class="font-medium">{member.username}</p>
                            <p class="text-sm text-gray-500">{member.role}</p>
                        </div>
                        {#if isCreator && member.id !== $auth.user?.id}
                            <div class="flex space-x-2">
                                <select 
                                    class="text-sm border rounded"
                                    value={member.role}
                                    on:change={(e) => updateMemberRole(member.id, e.target.value)}
                                >
                                    <option value="member">Member</option>
                                    <option value="moderator">Moderator</option>
                                    <option value="admin">Admin</option>
                                </select>
                                <Button 
                                    size="xs" 
                                    color="red"
                                    on:click={() => removeMember(member.id)}
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
    </div>
</Card>

<Modal bind:open={showInviteModal} size="md">
    <div class="space-y-6">
        <h3 class="text-xl font-medium">Invite Member</h3>
        <form on:submit|preventDefault={inviteMember} class="space-y-4">
            <div>
                <Label for="email">Member Email</Label>
                <Input
                    id="email"
                    type="email"
                    bind:value={inviteEmail}
                    required
                    placeholder="Enter member's email"
                />
            </div>
            <div class="flex justify-end space-x-2">
                <Button color="alternative" on:click={() => showInviteModal = false}>
                    Cancel
                </Button>
                <Button type="submit">Send Invitation</Button>
            </div>
        </form>
    </div>
</Modal> 