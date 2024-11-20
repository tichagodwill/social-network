<script lang="ts">
    import { Button, Card, Modal, Label, Input, Textarea } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { getFormattedDate } from '$lib/dateFormater';

    export let groupId: number;
    let events: any[] = [];
    let showCreateModal = false;
    let newEvent = {
        title: '',
        description: '',
        eventDate: '',
    };

    async function loadEvents() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/events`, {
                credentials: 'include'
            });
            if (response.ok) {
                events = await response.json();
            }
        } catch (error) {
            console.error('Failed to load events:', error);
        }
    }

    async function createEvent() {
        try {
            const response = await fetch(`http://localhost:8080/groups/${groupId}/events`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    ...newEvent,
                    creatorId: $auth.user?.id
                })
            });

            if (!response.ok) {
                throw new Error(await response.text());
            }

            showCreateModal = false;
            newEvent = {
                title: '',
                description: '',
                eventDate: ''
            };
            await loadEvents();
        } catch (error) {
            console.error('Failed to create event:', error);
        }
    }

    async function respondToEvent(eventId: number, status: 'going' | 'not_going') {
        try {
            const response = await fetch(`http://localhost:8080/groups/events/${eventId}/respond`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    status,
                    userId: $auth.user?.id
                })
            });

            if (!response.ok) {
                throw new Error(await response.text());
            }

            await loadEvents();
        } catch (error) {
            console.error('Failed to respond to event:', error);
        }
    }

    onMount(loadEvents);
</script>

<div class="space-y-4">
    <div class="flex justify-between items-center">
        <h3 class="text-xl font-semibold">Events</h3>
        <Button on:click={() => showCreateModal = true}>Create Event</Button>
    </div>

    {#if events.length === 0}
        <p class="text-gray-500">No events scheduled</p>
    {:else}
        <div class="space-y-4">
            {#each events as event}
                <Card>
                    <div class="space-y-2">
                        <h4 class="text-lg font-semibold">{event.title}</h4>
                        <p>{event.description}</p>
                        <p class="text-sm text-gray-500">
                            {getFormattedDate(event.eventDate).formated}
                        </p>
                        <div class="flex space-x-2 mt-4">
                            <Button 
                                size="sm" 
                                color="green"
                                on:click={() => respondToEvent(event.id, 'going')}
                            >
                                Going ({event.goingCount})
                            </Button>
                            <Button 
                                size="sm" 
                                color="red"
                                on:click={() => respondToEvent(event.id, 'not_going')}
                            >
                                Not Going ({event.notGoingCount})
                            </Button>
                        </div>
                    </div>
                </Card>
            {/each}
        </div>
    {/if}
</div>

<Modal bind:open={showCreateModal} size="md">
    <div class="space-y-6">
        <h3 class="text-xl font-medium">Create Event</h3>
        <form on:submit|preventDefault={createEvent} class="space-y-4">
            <div>
                <Label for="title">Event Title</Label>
                <Input
                    id="title"
                    bind:value={newEvent.title}
                    required
                />
            </div>
            <div>
                <Label for="description">Description</Label>
                <Textarea
                    id="description"
                    bind:value={newEvent.description}
                    required
                />
            </div>
            <div>
                <Label for="eventDate">Date & Time</Label>
                <Input
                    id="eventDate"
                    type="datetime-local"
                    bind:value={newEvent.eventDate}
                    required
                />
            </div>
            <div class="flex justify-end space-x-2">
                <Button color="alternative" on:click={() => showCreateModal = false}>
                    Cancel
                </Button>
                <Button type="submit">Create Event</Button>
            </div>
        </form>
    </div>
</Modal> 