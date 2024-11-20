<script lang="ts">
    import { Button, Avatar } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import type { User } from '$lib/types';

    export let groupId: number;
    export let members: User[] = [];
    export let isCreator = false;

    async function inviteMember(userId: number) {
        try {
            const response = await fetch('http://localhost:8080/groups/invitation', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    groupId,
                    inviterId: $auth.user?.id,
                    reciverId: userId
                })
            });

            if (!response.ok) {
                throw new Error(await response.text());
            }

            // Refresh members list
            await loadMembers();
        } catch (error) {
            console.error('Failed to invite member:', error);
        }
    }

    async function removeMember(userId: number) {
        try {
            const response = await fetch('http://localhost:8080/groups/leave', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    groupId,
                    userId
                })
            });

            if (!response.ok) {
                throw new Error(await response.text());
            }

            // Refresh members list
            await loadMembers();
        } catch (error) {
            console.error('Failed to remove member:', error);
        }
    }

    async function loadMembers() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/members`, {
                credentials: 'include'
            });
            if (response.ok) {
                members = await response.json();
            }
        } catch (error) {
            console.error('Failed to load members:', error);
        }
    }
</script>

<div class="space-y-4">
    <h3 class="text-xl font-semibold">Members</h3>
    
    {#if members.length === 0}
        <p class="text-gray-500">No members yet</p>
    {:else}
        <div class="space-y-2">
            {#each members as member}
                <div class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800 rounded-lg">
                    <div class="flex items-center space-x-3">
                        <Avatar src={member.avatar || '/default-avatar.png'} size="sm" />
                        <div>
                            <p class="font-medium">{member.username}</p>
                            <p class="text-sm text-gray-500">{member.firstName} {member.lastName}</p>
                        </div>
                    </div>
                    {#if isCreator && member.id !== $auth.user?.id}
                        <Button 
                            size="xs" 
                            color="red" 
                            on:click={() => removeMember(member.id)}
                        >
                            Remove
                        </Button>
                    {/if}
                </div>
            {/each}
        </div>
    {/if}
</div> 