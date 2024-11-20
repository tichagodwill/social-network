import { writable } from 'svelte/store';
import type { Group, CreateGroupRequest } from '$lib/types';

function createGroupsStore() {
    const { subscribe, set, update } = writable<Group[]>([]);

    return {
        subscribe,
        loadGroups: async () => {
            try {
                const response = await fetch('http://localhost:8080/groups', {
                    credentials: 'include'
                });
                if (response.ok) {
                    const groups = await response.json();
                    set(groups);
                }
            } catch (error) {
                console.error('Failed to load groups:', error);
            }
        },
        createGroup: async (groupData: CreateGroupRequest) => {
            try {
                const response = await fetch('http://localhost:8080/groups', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify(groupData)
                });

                if (!response.ok) {
                    throw new Error(await response.text());
                }

                const newGroup = await response.json();
                update(groups => [...groups, newGroup]);
                return newGroup;
            } catch (error) {
                console.error('Failed to create group:', error);
                throw error;
            }
        }
    };
}

export const groups = createGroupsStore(); 