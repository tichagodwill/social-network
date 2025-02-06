<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { Button, Card, Modal, Label, Input, Textarea } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { getFormattedDate } from '$lib/dateFormater';
    import { fade, slide } from 'svelte/transition';
    import { groups } from '$lib/stores/groups';

    export let groupId: number;
    let events: any[] = [];
    let showCreateModal = false;
    let error = '';
    let loading = true;
    let newEvent = {
        title: '',
        description: '',
        eventDate: ''
    };

    // Add response tracking
    let responsesByEvent: { [key: number]: { going: number, notGoing: number } } = {};
    let userResponses: { [key: number]: string } = {};

    let socket: WebSocket;

    function connectWebSocket() {
        socket = new WebSocket('ws://localhost:8080/ws');
        
        socket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'eventRSVP' && data.groupId === groupId) {
                // Update the response counts for the specific event
                responsesByEvent[data.eventId] = {
                    going: data.going,
                    notGoing: data.notGoing
                };
                // Force reactivity
                responsesByEvent = { ...responsesByEvent };
            }
        };

        socket.onclose = () => {
            // Implement reconnection logic if needed
            setTimeout(connectWebSocket, 1000);
        };
    }

    async function loadEvents() {
        try {
            loading = true;
            const response = await fetch(`http://localhost:8080/groups/${groupId}/events`, {
                credentials: 'include'
            });
            if (response.ok) {
                const data = await response.json();
                events = data.map(event => ({
                    ...event,
                    eventDate: new Date(event.eventDate) // Parse the ISO date string
                }));
                
                // Initialize response counts and user responses
                events.forEach(event => {
                    responsesByEvent[event.id] = {
                        going: event.responses?.going || 0,
                        notGoing: event.responses?.notGoing || 0
                    };
                    if (event.userResponse) {
                        userResponses[event.id] = event.userResponse;
                    }
                });

                // Force reactivity
                responsesByEvent = { ...responsesByEvent };
                userResponses = { ...userResponses };
            } else {
                events = [];
            }
        } catch (err) {
            console.error('Failed to load events:', err);
            error = err instanceof Error ? err.message : 'Failed to load events';
        } finally {
            loading = false;
        }
    }

    async function createEvent() {
        try {
            if (!$auth?.user) return;
            
            await groups.createEvent(groupId, {
                title: newEvent.title,
                description: newEvent.description,
                eventDate: newEvent.eventDate,
                creatorId: $auth.user.id
            });

            await loadEvents();
            showCreateModal = false;
            newEvent = { title: '', description: '', eventDate: '' };
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to create event';
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

    async function respondToEvent(eventId: number, rsvpStatus: 'going' | 'not_going') {
        try {
            console.log('1. Starting RSVP request:', { eventId, rsvpStatus });
            
            const fetchResponse = await fetch(`http://localhost:8080/groups/${groupId}/events/${eventId}/respond`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({ status: rsvpStatus })
            });
            
            console.log('2. Response status:', fetchResponse.status);
            console.log('3. Response headers:', Object.fromEntries(fetchResponse.headers.entries()));

            const responseText = await fetchResponse.text();
            console.log('4. Raw response:', responseText);

            let responseData;
            try {
                responseData = JSON.parse(responseText);
            } catch (e) {
                throw new Error(responseText || 'Failed to parse server response');
            }

            if (!fetchResponse.ok) {
                throw new Error(responseData.error || 'Failed to respond to event');
            }

            // Update local state
            userResponses[eventId] = rsvpStatus;
            responsesByEvent[eventId] = {
                going: responseData.going,
                notGoing: responseData.notGoing
            };

            // Force reactivity
            responsesByEvent = { ...responsesByEvent };
            userResponses = { ...userResponses };

        } catch (error) {
            console.error('Error in respondToEvent:', error);
            throw error;
        }
    }

    onMount(() => {
        loadEvents();
        connectWebSocket();
    });

    onDestroy(() => {
        if (socket) {
            socket.close();
        }
    });
</script>

<style>
    .event-card {
        transform: translateY(0);
        transition: all 0.3s ease;
        margin-bottom: 1rem;
    }

    .event-card:hover {
        transform: translateY(-4px);
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    }

    .event-card .admin-buttons {
        opacity: 0;
        transition: opacity 0.2s ease;
    }

    .event-card:hover .admin-buttons {
        opacity: 1;
    }

    .event-badge {
        display: inline-flex;
        align-items: center;
        padding: 0.25rem 0.75rem;
        border-radius: 9999px;
        font-size: 0.75rem;
        font-weight: 500;
    }

    .event-badge.upcoming {
        background-color: rgb(220, 252, 231);
        color: rgb(22, 101, 52);
    }

    .event-badge.past {
        background-color: rgb(243, 244, 246);
        color: rgb(55, 65, 81);
    }

    .response-count {
        font-size: 0.875rem;
        color: rgb(107, 114, 128);
    }

    .animate-pulse {
        animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
    }

    @keyframes pulse {
        0%, 100% {
            opacity: 1;
        }
        50% {
            opacity: .5;
        }
    }

    .response-stats {
        display: inline-flex;
        align-items: center;
        padding: 0.5rem 1rem;
        background-color: rgb(243, 244, 246);
        border-radius: 0.5rem;
        font-weight: 500;
    }

    .response-stats-count {
        color: rgb(79, 70, 229);
        font-weight: 600;
        margin-left: 0.25rem;
    }

    :global(.dark) .response-stats {
        background-color: rgb(31, 41, 55);
    }

    :global(.dark) .response-stats-count {
        color: rgb(129, 140, 248);
    }
</style>

<div class="space-y-4">
    <div class="flex justify-between items-center">
        <h3 class="text-xl font-semibold">Events</h3>
        <Button 
            gradient
            color="blue"
            class="transform hover:scale-105 transition-transform duration-200"
            on:click={() => showCreateModal = true}
        >
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
            </svg>
            Create Event
        </Button>
    </div>

    {#if error}
        <div transition:fade>
            <div class="p-4 text-red-800 bg-red-100 rounded-lg">
                {error}
            </div>
        </div>
    {/if}

    <div class="space-y-4">
        {#if loading}
            <div class="space-y-4">
                {#each Array(2) as _}
                    <div class="animate-pulse">
                        <Card>
                            <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-4"></div>
                            <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-2"></div>
                            <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
                        </Card>
                    </div>
                {/each}
            </div>
        {:else if events.length === 0}
            <Card>
                <div class="text-center py-8">
                    <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
                    </svg>
                    <p class="text-gray-500 text-lg">No events scheduled yet</p>
                </div>
            </Card>
        {:else}
            {#each events as event (event.id)}
                <div transition:slide class="mb-4">
                    <Card class="event-card">
                        <div class="space-y-4">
                            <div class="flex justify-between items-start">
                                <div>
                                    <div class="flex items-center space-x-2">
                                        <h4 class="text-xl font-semibold">{event.title}</h4>
                                        <span class="event-badge {new Date(event.eventDate) > new Date() ? 'upcoming' : 'past'}">
                                            {new Date(event.eventDate) > new Date() ? 'Upcoming' : 'Past'}
                                        </span>
                                    </div>
                                    <p class="text-sm text-gray-500 mt-1">
                                        {getFormattedDate(new Date(event.eventDate)).formated}
                                    </p>
                                </div>
                            </div>

                            <p class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
                                {event.description}
                            </p>

                            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                                <div class="flex flex-wrap gap-2">
                                    <Button 
                                        size="sm"
                                        color={userResponses[event.id] === 'going' ? 'green' : 'light'}
                                        on:click={() => respondToEvent(event.id, 'going')}
                                        class="flex-1 sm:flex-none justify-center"
                                    >
                                        Going ({responsesByEvent[event.id]?.going || 0})
                                    </Button>
                                    <Button 
                                        size="sm"
                                        color={userResponses[event.id] === 'not_going' ? 'red' : 'light'}
                                        on:click={() => respondToEvent(event.id, 'not_going')}
                                        class="flex-1 sm:flex-none justify-center"
                                    >
                                        Not Going ({responsesByEvent[event.id]?.notGoing || 0})
                                    </Button>
                                </div>
                                <div class="response-stats">
                                    <span>Total Responses:</span>
                                    <span class="response-stats-count">
                                        {(responsesByEvent[event.id]?.going || 0) + (responsesByEvent[event.id]?.notGoing || 0)}
                                    </span>
                                </div>
                            </div>
                        </div>
                    </Card>
                </div>
            {/each}
        {/if}
    </div>
</div>

<Modal bind:open={showCreateModal} size="lg" autoclose={false}>
    <div class="space-y-6">
        <h3 class="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            Create New Event
        </h3>
        
        {#if error}
            <div transition:fade>
                <div class="p-4 text-red-800 bg-red-100 rounded-lg">
                    {error}
                </div>
            </div>
        {/if}

        <form on:submit|preventDefault={handleSubmit} class="space-y-4">
            <div>
                <Label for="title" class="text-lg mb-2">Event Title</Label>
                <Input
                    id="title"
                    bind:value={newEvent.title}
                    required
                    placeholder="Enter event title"
                    class="transition-all duration-300 focus:ring-2 focus:ring-blue-500"
                />
            </div>
            <div>
                <Label for="description" class="text-lg mb-2">Description</Label>
                <Textarea
                    id="description"
                    bind:value={newEvent.description}
                    required
                    placeholder="Describe your event..."
                    rows={4}
                    class="transition-all duration-300 focus:ring-2 focus:ring-blue-500"
                />
            </div>
            <div>
                <Label for="eventDate" class="text-lg mb-2">Date & Time</Label>
                <Input
                    id="eventDate"
                    type="datetime-local"
                    bind:value={newEvent.eventDate}
                    required
                    class="transition-all duration-300 focus:ring-2 focus:ring-blue-500"
                />
            </div>
            <div class="flex justify-end space-x-2">
                <Button 
                    color="alternative" 
                    on:click={() => {
                        showCreateModal = false;
                        newEvent = { title: '', description: '', eventDate: '' };
                        error = '';
                    }}
                >
                    Cancel
                </Button>
                <Button 
                    type="submit"
                    gradient
                    color="blue"
                    class="transform hover:scale-105 transition-transform duration-200"
                >
                    Create Event
                </Button>
            </div>
        </form>
    </div>
</Modal> 