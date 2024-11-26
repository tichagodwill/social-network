<script lang="ts">
    import { Navbar, NavBrand, NavLi, NavUl, NavHamburger, Button } from 'flowbite-svelte';
    import { auth } from '$lib/stores/auth';
    import NotificationBell from './NotificationBell.svelte';
    import { onMount } from 'svelte';
    import { notifications } from '$lib/stores/notifications';
    import { chat } from '$lib/stores/chat';
    import { goto } from '$app/navigation';

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
</script>

<Navbar let:hidden let:toggle rounded color="primary">
    <NavBrand href="/">
        <span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white">
            Social Network
        </span>
    </NavBrand>
    
    <NavHamburger on:click={toggle} />
    
    <NavUl {hidden}>
        {#if $auth.isAuthenticated}
            <NavLi href="/posts">Posts</NavLi>
            <NavLi href="/groups">Groups</NavLi>
            <NavLi href="/chat">Chat</NavLi>
            <NavLi href="/profile/{$auth.user?.id}">Profile</NavLi>
            <NotificationBell />
            <li>
                <Button 
                    color="alternative" 
                    on:click={handleLogout}
                >
                    Logout
                </Button>
            </li>
        {:else}
            <NavLi href="/login">Login</NavLi>
            <NavLi href="/register">Sign Up</NavLi>
        {/if}
    </NavUl>
</Navbar> 