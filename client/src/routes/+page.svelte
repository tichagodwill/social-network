<script lang="ts">
  import Navbar from '$lib/components/Navbar.svelte';
  import { auth } from '$lib/stores/auth';

  export let data: PageData;
  const isLoggedIn = data.user?.loggedIn;
</script>


<main class="container mx-auto px-4 py-8">
    <div class="text-center">
        <h1 class="text-4xl font-bold mb-6 dark:text-white">
            Welcome to Social Network
        </h1>
        
        <div class="max-w-2xl mx-auto">
            <p class="text-lg mb-8 dark:text-gray-300">
                Connect with friends, share your thoughts, and join communities of like-minded people.
            </p>
            
            {#if !$auth.isAuthenticated}
                <div class="space-x-4">
                    <a 
                        href="/register" 
                        class="inline-block bg-primary-600 text-white px-6 py-3 rounded-lg hover:bg-primary-700 transition-colors"
                    >
                        Get Started
                    </a>
                    <a 
                        href="/login" 
                        class="inline-block bg-gray-200 text-gray-800 px-6 py-3 rounded-lg hover:bg-gray-300 transition-colors dark:bg-gray-700 dark:text-white dark:hover:bg-gray-600"
                    >
                        Sign In
                    </a>
                </div>
            {:else}
                <div class="space-x-4">
                    <a 
                        href="/posts" 
                        class="inline-block bg-primary-600 text-white px-6 py-3 rounded-lg hover:bg-primary-700 transition-colors"
                    >
                        View Posts
                    </a>
                    <a 
                        href="/groups" 
                        class="inline-block bg-gray-200 text-gray-800 px-6 py-3 rounded-lg hover:bg-gray-300 transition-colors dark:bg-gray-700 dark:text-white dark:hover:bg-gray-600"
                    >
                        Browse Groups
                    </a>
                </div>
            {/if}
        </div>
    </div>

    <div class="mt-16 grid md:grid-cols-3 gap-8">
        <div class="p-6 bg-white rounded-lg shadow-md dark:bg-gray-800">
            <h2 class="text-xl font-semibold mb-4 dark:text-white">Connect</h2>
            <p class="dark:text-gray-300">Find and connect with friends, family, and new people who share your interests.</p>
        </div>
        
        <div class="p-6 bg-white rounded-lg shadow-md dark:bg-gray-800">
            <h2 class="text-xl font-semibold mb-4 dark:text-white">Share</h2>
            <p class="dark:text-gray-300">Share your thoughts, photos, and experiences with your network.</p>
        </div>
        
        <div class="p-6 bg-white rounded-lg shadow-md dark:bg-gray-800">
            <h2 class="text-xl font-semibold mb-4 dark:text-white">Join Groups</h2>
            <p class="dark:text-gray-300">Find communities of people who share your passions and interests.</p>
        </div>
    </div>

    {#if isLoggedIn}
        <div class="mt-8">
            <!-- Feed or other logged-in content can go here -->
        </div>
    {/if}
</main>
