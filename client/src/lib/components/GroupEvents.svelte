<script lang="ts">
    import { onMount } from 'svelte';
    import { Button, Card, Modal, Label, Input, Textarea } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { getFormattedDate } from '$lib/dateFormater';

    export let groupId: number;
    let events: any[] = [];
    let showCreateModal = false;
    let error = '';
    let loading = true;
    let newEvent = {
        title: '',
        description: '',
        eventDate: '',
    };

    async function loadEvents() {
        try {
            loading = true;
            const response = await fetch(`http://localhost:8080/groups/${groupId}/events`, {
                credentials: 'include'
            });
            if (response.ok) {
                events = await response.json() || [];
            } else {
                events = [];
            }
        } catch (error) {
            console.error('Failed to load events:', error);
            events = [];
        } finally {
            loading = false;
        }
    }

    async function handleSubmit(event: SubmitEvent) {
        event.preventDefault();
        try {
            if (!newEvent.title || !newEvent.description || !newEvent.eventDate) {
                error = 'Please fill in all fields';
                return;
            }

            const eventDate = new Date(newEvent.eventDate).toISOString();

            const response = await fetch(`http://localhost:8080/groups/${groupId}/events`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    title: newEvent.title,
                    description: newEvent.description,
                    eventDate: eventDate,
                    creatorId: $auth.user?.id
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to create event');
            }

            showCreateModal = false;
            newEvent = {
                title: '',
                description: '',
                eventDate: ''
            };
            error = '';

            await loadEvents();
        } catch (err) {
            console.error('Failed to create event:', err);
            error = err instanceof Error ? err.message : 'Failed to create event';
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
                    userId: $auth.user?.id,
                    status
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to respond to event');
            }

            await loadEvents();
        } catch (err) {
            console.error('Failed to respond to event:', err);
        }
    }

    onMount(loadEvents);
</script>

<div class="space-y-4">
    <div class="flex justify-between items-center">
        <h3 class="text-xl font-semibold">Events</h3>
        <Button on:click={() => showCreateModal = true}>Create Event</Button>
    </div>

    {#if loading}
        <div class="text-center py-4">
            <div class="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
        </div>
    {:else if events.length === 0}
        <p class="text-gray-500">No events scheduled</p>
    {:else}
        <div class="space-y-4">
            {#each events as event}
                <Card>
                    <div class="space-y-2">
                        <h4 class="text-lg font-semibold">{event.title}</h4>
                        <p>{event.description}</p>
                        <p class="text-sm text-gray-500">
                            {getFormattedDate(new Date(event.eventDate)).formated}
                        </p>
                        <div class="flex space-x-2 mt-4">
                            <Button 
                                size="sm" 
                                color="green"
                                on:click={() => respondToEvent(event.id, 'going')}
                            >
                                Going ({event.goingCount || 0})
                            </Button>
                            <Button 
                                size="sm" 
                                color="red"
                                on:click={() => respondToEvent(event.id, 'not_going')}
                            >
                                Not Going ({event.notGoingCount || 0})
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
        {#if error}
            <div class="p-4 text-red-800 bg-red-100 rounded-lg">
                {error}
            </div>
        {/if}
        <form on:submit={handleSubmit} class="space-y-4">
            <div>
                <Label for="title">Event Title</Label>
                <Input
                    id="title"
                    bind:value={newEvent.title}
                    required
                    placeholder="Enter event title"
                />
            </div>
            <div>
                <Label for="description">Description</Label>
                <Textarea
                    id="description"
                    bind:value={newEvent.description}
                    required
                    placeholder="Describe your event"
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
                <Button color="alternative" on:click={() => {
                    showCreateModal = false;
                    error = '';
                }}>
                    Cancel
                </Button>
                <Button type="submit">Create Event</Button>
            </div>
        </form>
    </div>
</Modal> 