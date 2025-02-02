<script lang="ts">
    import {Navbar, NavBrand, NavLi, NavUl, NavHamburger, Button} from 'flowbite-svelte';
    import {auth} from '$lib/stores/auth';
    import NotificationBell from './NotificationBell.svelte';
    import {onMount} from 'svelte';
    import {notifications} from '$lib/stores/notifications';
    import {chat} from '$lib/stores/chat';
    import {goto} from '$app/navigation';
    import { page } from '$app/stores';

    onMount(() => {
        if ($auth.isAuthenticated) {
            const socket = new WebSocket('ws://localhost:8080/ws');
            notifications.initialize(socket);
            chat.initialize();
            notifications.loadNotifications();
        }
    });

    async function handleLogout() {
        try {
            await auth.logout();
            goto('/');
        } catch (error) {
            console.error('Logout failed:', error);
        }
    }

    // Track active route for navigation highlighting using page store
    $: currentPath = $page.url.pathname;
</script>

<div class="navbar-spacer"></div>
<Navbar
        let:hidden
        let:toggle
        rounded={false}
        color="primary"
        class="fixed top-0 left-0 right-0 z-50 border-b border-gray-200 bg-white shadow-sm"
>
    <NavBrand href="/">
        <span class="flex items-center space-x-2">
            <!-- Logo -->
            <svg class="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4"/>
            </svg>
            <span class="self-center text-xl font-semibold text-gray-900 hover:text-blue-600 transition-colors duration-200">
                SocialNet
            </span>
        </span>
    </NavBrand>

    <NavHamburger
            on:click={toggle}
            class="focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
    />

    <NavUl
            {hidden}
            class="md:flex md:items-center md:space-x-1"
    >
        {#if $auth.isAuthenticated}
            <NavLi
                    href="/posts"
                    class="nav-item {currentPath === '/posts' ? 'active' : ''}"
            >
                Posts
            </NavLi>
            <NavLi
                    href="/groups"
                    class="nav-item {currentPath === '/groups' ? 'active' : ''}"
            >
                Groups
            </NavLi>
            <NavLi
                    href="/chat"
                    class="nav-item {currentPath === '/chat' ? 'active' : ''}"
            >
                Chat
            </NavLi>
            <NavLi
                    href="/profile/{$auth.user?.id}"
                    class="nav-item {currentPath.startsWith('/profile') ? 'active' : ''}"
            >
                Profile
            </NavLi>
            <NavLi
                    href="/explore"
                    class="nav-item {currentPath.startsWith('/explore') ? 'active' : ''}"
            >
                Explore
            </NavLi>

            <div class="flex items-center space-x-2 md:ml-4">
                <NotificationBell/>
                <Button
                        color="alternative"
                        class="nav-button"
                        on:click={handleLogout}
                >
                    <span class="flex items-center space-x-1">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
                        </svg>
                        <span>Logout</span>
                    </span>
                </Button>
            </div>
        {:else}
            <div class="flex items-center space-x-2 md:ml-4">
                <Button
                        color="alternative"
                        href="/login"
                        class="nav-button-secondary"
                >
                    Login
                </Button>
                <Button
                        color="primary"
                        href="/register"
                        class="nav-button-primary"
                >
                    Sign Up
                </Button>
            </div>
        {/if}
    </NavUl>
</Navbar>

<style lang="postcss">
    /* Add navbar spacer to prevent content from being covered */
    .navbar-spacer {
        height: 64px; /* Adjust this value based on your navbar height */
    }

    /* Enhanced clickable elements and interactions */
    :global(.nav-item) {
        @apply px-3 py-2 text-gray-600 rounded-lg transition-all duration-200
        hover:text-blue-600 hover:bg-blue-50
        active:bg-blue-100
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2;
    }

    :global(.nav-item.active) {
        @apply text-blue-600 bg-blue-50 font-medium;
    }

    :global(.nav-button) {
        @apply px-4 py-2 text-gray-700 bg-gray-100 rounded-lg
        transition-all duration-200
        hover:bg-gray-200 hover:text-gray-900
        active:bg-gray-300
        focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2;
    }

    :global(.nav-button-secondary) {
        @apply px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg
        transition-all duration-200
        hover:bg-gray-50 hover:border-gray-400
        active:bg-gray-100
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2;
    }

    :global(.nav-button-primary) {
        @apply px-4 py-2 text-white bg-blue-600 rounded-lg
        transition-all duration-200
        hover:bg-blue-700
        active:bg-blue-800
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2;
    }
</style>