<script lang="ts">
    import Navbar from '$lib/components/Navbar.svelte';
    import { auth } from '$lib/stores/auth';
    import { onMount } from 'svelte';

    export let data: PageData;
    const isLoggedIn = data.user?.loggedIn;

    // Optional: Add animation on mount
    onMount(() => {
        // Any initialization code can go here
    });
</script>

<svelte:head>
    <title>SocialNet - Connect with Friends</title>
    <meta name="description" content="Connect with friends, share your thoughts, and join communities on SocialNet" />
</svelte:head>

<div class="min-h-screen bg-gray-50">
    <!-- Navbar can be imported here -->
    <main class="pt-16"> <!-- Added padding-top to account for fixed navbar -->
        <!-- Hero Section -->
        <div class="container mx-auto px-4 sm:px-6 lg:px-8 py-12 sm:py-16">
            <div class="text-center">
                <h1 class="text-4xl sm:text-5xl md:text-6xl font-extrabold mb-6">
                    <span class="block text-gray-900">Welcome to</span>
                    <span class="block text-blue-600 mt-2">SocialNet</span>
                </h1>

                <div class="max-w-2xl mx-auto">
                    <p class="text-lg sm:text-xl text-gray-600 mb-8 leading-relaxed">
                        Connect with friends, share your thoughts, and join communities of like-minded people in a space designed for meaningful connections.
                    </p>

                    {#if !$auth.isAuthenticated}
                        <div class="space-y-4 sm:space-y-0 sm:space-x-4">
                            <a
                                    href="/register"
                                    class="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700
                       transition-all duration-200 transform hover:scale-105 font-semibold shadow-lg"
                            >
                                Get Started
                            </a>
                            <a
                                    href="/login"
                                    class="inline-block border-2 border-blue-600 text-blue-600 px-8 py-3 rounded-lg
                       hover:bg-blue-50 transition-all duration-200 transform hover:scale-105 font-semibold"
                            >
                                Sign In
                            </a>
                        </div>
                    {:else}
                        <div class="space-y-4 sm:space-y-0 sm:space-x-4">
                            <a
                                    href="/posts"
                                    class="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700
                       transition-all duration-200 transform hover:scale-105 font-semibold shadow-lg"
                            >
                                View Posts
                            </a>
                            <a
                                    href="/groups"
                                    class="inline-block border-2 border-blue-600 text-blue-600 px-8 py-3 rounded-lg
                       hover:bg-blue-50 transition-all duration-200 transform hover:scale-105 font-semibold"
                            >
                                Browse Groups
                            </a>
                        </div>
                    {/if}
                </div>
            </div>
        </div>

        <!-- Features Grid -->
        <div class="container mx-auto px-4 sm:px-6 lg:px-8 py-12">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
                <!-- Connect Feature -->
                <div class="group">
                    <div class="p-8 bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300
                      border border-gray-100 h-full transform hover:-translate-y-1">
                        <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-6
                        group-hover:bg-blue-200 transition-colors duration-300">
                            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                      d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                            </svg>
                        </div>
                        <h2 class="text-xl font-semibold mb-4 text-gray-900 group-hover:text-blue-600 transition-colors duration-300">
                            Connect
                        </h2>
                        <p class="text-gray-600 leading-relaxed">
                            Find and connect with friends, family, and new people who share your interests. Build your network
                            and stay connected with what matters most to you.
                        </p>
                    </div>
                </div>

                <!-- Share Feature -->
                <div class="group">
                    <div class="p-8 bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300
                      border border-gray-100 h-full transform hover:-translate-y-1">
                        <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-6
                        group-hover:bg-blue-200 transition-colors duration-300">
                            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                      d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                            </svg>
                        </div>
                        <h2 class="text-xl font-semibold mb-4 text-gray-900 group-hover:text-blue-600 transition-colors duration-300">
                            Share
                        </h2>
                        <p class="text-gray-600 leading-relaxed">
                            Share your moments, thoughts, and experiences with your network. Post updates, photos, and stories
                            that matter to you and your community.
                        </p>
                    </div>
                </div>

                <!-- Groups Feature -->
                <div class="group">
                    <div class="p-8 bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300
                      border border-gray-100 h-full transform hover:-translate-y-1">
                        <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-6
                        group-hover:bg-blue-200 transition-colors duration-300">
                            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                      d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                            </svg>
                        </div>
                        <h2 class="text-xl font-semibold mb-4 text-gray-900 group-hover:text-blue-600 transition-colors duration-300">
                            Join Groups
                        </h2>
                        <p class="text-gray-600 leading-relaxed">
                            Find and join communities of people who share your passions and interests. Create meaningful
                            connections in spaces designed for collaboration and discussion.
                        </p>
                    </div>
                </div>
            </div>
        </div>

        {#if isLoggedIn}
            <div class="container mx-auto px-4 sm:px-6 lg:px-8 py-12">
                <!-- Feed Section -->
                <div class="max-w-3xl mx-auto bg-white rounded-xl shadow-sm border border-gray-100 p-6">
                    <!-- Add your feed component here -->
                    <div class="space-y-4">
                        <!-- This is a placeholder for the feed content -->
                        <h2 class="text-2xl font-semibold text-gray-900">Your Feed</h2>
                        <!-- Add your feed items here -->
                    </div>
                </div>
            </div>
        {/if}
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-gray-100 mt-16">
        <div class="container mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div class="text-center text-gray-600">
                <p>&copy; {new Date().getFullYear()} SocialNet. All rights reserved.</p>
            </div>
        </div>
    </footer>
</div>

<style>
    /* Add any custom styles here if needed */
    :global(body) {
        background-color: rgb(249, 250, 251); /* bg-gray-50 */
    }
</style>