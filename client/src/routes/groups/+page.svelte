<script lang="ts">
    import { Button, Card } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import { Calendar, Users, Sparkles } from 'lucide-svelte';
    import { fade, fly } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';

    export let data;

    $: groups = data.groups || [];
    $: error = data.error;
</script>

<div class="max-w-7xl mx-auto p-4 space-y-8">
    <div class="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 pb-6"
         in:fly="{{ y: -20, duration: 600, delay: 200 }}">
        <div>
            <h1 class="text-3xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 dark:from-indigo-400 dark:to-purple-400 bg-clip-text text-transparent flex items-center gap-2">
                <Sparkles class="w-8 h-8 text-indigo-600 dark:text-indigo-400 animate-pulse" />
                Groups
            </h1>
            <p class="text-gray-600 dark:text-gray-400 mt-2 animate-fade-in">
                Discover and join amazing communities
            </p>
        </div>
        {#if $auth.isAuthenticated}
            <Button href="/groups/create" 
                   class="create-button">
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                </svg>
                Create Group
            </Button>
        {/if}
    </div>

    {#if error}
        <div class="p-4 mb-4 text-red-800 bg-red-100 rounded-lg shadow-sm border border-red-200"
             in:fly="{{ y: 20, duration: 400 }}"
             out:fade>
            <div class="flex items-center">
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                {error}
            </div>
        </div>
    {/if}

    {#if !$auth.isAuthenticated}
        <div class="text-center p-12 bg-gradient-to-br from-gray-50 to-white dark:from-gray-800 dark:to-gray-900 rounded-lg shadow-sm"
             in:fly="{{ y: 20, duration: 400 }}">
            <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
            </svg>
            <p class="text-xl font-semibold mb-4 text-gray-800 dark:text-white">Please log in to view groups</p>
            <Button href="/login" size="xl" color="blue" class="shadow-sm">
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1"/>
                </svg>
                Log In
            </Button>
        </div>
    {:else if groups.length === 0}
        <div in:fly="{{ y: 20, duration: 400 }}">
            <Card class="text-center bg-gray-50 dark:bg-gray-800">
                <div class="p-8">
                    <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14v6m-3-3h6M6 10h2a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v2a2 2 0 002 2zm10 0h2a2 2 0 002-2V6a2 2 0 00-2-2h-2a2 2 0 00-2 2v2a2 2 0 002 2zM6 20h2a2 2 0 002-2v-2a2 2 0 00-2-2H6a2 2 0 00-2 2v2a2 2 0 002 2z"/>
                    </svg>
                    <p class="text-xl font-semibold text-gray-800 dark:text-white mb-2">No groups found</p>
                    <p class="text-gray-600 dark:text-gray-400">
                        Create a new group or join existing ones to get started
                    </p>
                </div>
            </Card>
        </div>
    {:else}
        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {#each groups as group, i (group.id)}
                <div in:fly="{{ y: 20, duration: 400, delay: i * 100 }}">
                    <Card class="group-card">
                        <div class="p-6">
                            <div class="flex justify-between items-start mb-4">
                                <div class="flex-1">
                                    <h3 class="text-xl font-bold group-title">
                                        <a href="/groups/{group.id}" 
                                           class="text-gradient">
                                            {group.title}
                                        </a>
                                    </h3>
                                    {#if group.is_member}
                                        <span class="member-badge">
                                            <svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                                                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
                                            </svg>
                                            Member
                                        </span>
                                    {/if}
                                </div>
                            </div>
                            
                            <p class="text-gray-600 dark:text-gray-300 mb-6 line-clamp-2 leading-relaxed group-hover:line-clamp-none transition-all duration-300">
                                {group.description}
                            </p>

                            <div class="space-y-3 mb-6 metadata">
                                <div class="flex items-center text-sm text-gray-500 dark:text-gray-400 metadata-item">
                                    <Users class="w-4 h-4 mr-2" />
                                    <span class="font-medium mr-1">Created by</span>
                                    <span class="text-gray-700 dark:text-gray-300">{group.creator_username}</span>
                                </div>
                                <div class="flex items-center text-sm text-gray-500 dark:text-gray-400 metadata-item">
                                    <Calendar class="w-4 h-4 mr-2" />
                                    <span>{new Date(group.created_at).toLocaleDateString('en-US', {
                                        year: 'numeric',
                                        month: 'long',
                                        day: 'numeric'
                                    })}</span>
                                </div>
                            </div>

                            <div class="flex justify-end space-x-3">
                                <Button 
                                    href="/groups/{group.id}" 
                                    class="view-details-button"
                                >
                                    <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                                    </svg>
                                    View Details
                                </Button>
                                {#if !group.is_member}
                                    <Button 
                                        class="join-button"
                                    >
                                        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"/>
                                        </svg>
                                        Join Group
                                    </Button>
                                {/if}
                            </div>
                        </div>
                    </Card>
                </div>
            {/each}
        </div>
    {/if}
</div>

<style lang="postcss">
    .group-card {
        @apply bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 
               shadow-sm hover:shadow-xl transition-all duration-500;
        animation: fadeIn 0.5s ease-out;
    }

    .group-card:hover {
        @apply transform -translate-y-1;
    }

    .text-gradient {
        @apply bg-gradient-to-r from-indigo-600 to-purple-600 dark:from-indigo-400 dark:to-purple-400 
               bg-clip-text text-transparent hover:from-purple-600 hover:to-indigo-600 
               dark:hover:from-purple-400 dark:hover:to-indigo-400 transition-all duration-300;
    }

    .create-button {
        @apply bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600
               text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5
               transition-all duration-300;
    }

    .view-details-button {
        @apply bg-gradient-to-r from-indigo-50 to-purple-50 hover:from-indigo-100 hover:to-purple-100
               text-indigo-700 border border-indigo-100 shadow-sm hover:shadow
               dark:from-indigo-900 dark:to-purple-900 dark:text-indigo-200
               dark:hover:from-indigo-800 dark:hover:to-purple-800 dark:border-indigo-700
               transition-all duration-300 transform hover:-translate-y-0.5;
    }

    .join-button {
        @apply bg-gradient-to-r from-blue-500 to-indigo-500 hover:from-blue-600 hover:to-indigo-600
               text-white shadow-sm hover:shadow transform hover:-translate-y-0.5
               dark:from-blue-600 dark:to-indigo-600 dark:hover:from-blue-700 dark:hover:to-indigo-700
               transition-all duration-300;
    }

    .member-badge {
        @apply inline-flex items-center px-3 py-1 mt-2 text-xs font-medium
               bg-gradient-to-r from-green-100 to-emerald-100 text-green-800
               dark:from-green-900 dark:to-emerald-900 dark:text-green-200
               rounded-full transform hover:scale-105 transition-all duration-300;
    }

    .metadata-item {
        @apply transform hover:translate-x-1 transition-all duration-300;
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(10px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    .animate-fade-in {
        animation: fadeIn 0.5s ease-out;
    }

    :global(.dark) .group-card {
        @apply bg-gray-800 border-gray-700;
    }

    :global(.dark) .text-gray-600 {
        @apply text-gray-300;
    }

    :global(.dark) .text-gray-500 {
        @apply text-gray-400;
    }
</style> 