<script lang="ts">
    import { page } from '$app/stores';
    import { onMount } from 'svelte';
    import { followers } from '$lib/stores/followers';
    import { auth } from '$lib/stores/auth';
    import { goto } from '$app/navigation';
    import { Button, Avatar, Modal, } from 'flowbite-svelte';
    import { fade, slide, fly } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';
    import type { PageData } from './$types';
    import { error } from '@sveltejs/kit';
    import { getFormattedDate } from '$lib/dateFormater'

    export let data: PageData;
    
    // Get userId from URL and watch for changes
    $: userId = parseInt($page.params.id);
    let isOwnProfile = false;
    let isFollowing = false;
    let hasPendingRequest = false;
    let isLoading = false;
    let errorMessage = '';
    let showSettingsModal = false;
    let userPosts: Array<any> = [];
    let showExpandedImage = false;
    let expandedImageSrc = '';
    let showUnfollowModal = false;
    let activeTab = 'posts';
    let previousTab = 'posts';

    // Initialize the followers store with the requests data
    $: if (data.Requests && data.Requests.length > 0) {
        followers.update(state => ({
            ...state,
            requests: data.Requests
        }));
    }

    // Settings state management
    let originalProfilePhoto: string = data.user?.avatar ?? "";
    let newProfilePhoto: string = originalProfilePhoto;
    let originalPrivacySetting: string = data.user?.isPrivate ? "private" : "public";
    let privacySetting: string = originalPrivacySetting;
    let originalDescription: string = data.user?.aboutMe || '';
    let userDescription: string = originalDescription;
    let hasCustomPhoto = !!data.user?.avatar;
    let isUsingDefault = false;
    let hasNewUpload = false;

    // Reset settings to original values when modal is closed
    $: if (!showSettingsModal) {
        resetSettings();
    }

    function resetSettings() {
        newProfilePhoto = originalProfilePhoto;
        privacySetting = originalPrivacySetting;
        userDescription = originalDescription;
        hasCustomPhoto = !!originalProfilePhoto;
        isUsingDefault = false;
        hasNewUpload = false;
    }

    // Function to generate avatar with the first letter of the username
    function generateAvatar(username: string): string {
        const firstLetter = username ? username.charAt(0).toUpperCase() : 'U';
        return `https://ui-avatars.com/api/?name=${encodeURIComponent(firstLetter)}&background=random`;
    }

    // Function to handle file upload for profile photo
    async function handleFileUpload(event: Event) {
        const input = event.target as HTMLInputElement;
        const file = input.files?.[0];
        
        if (file) {
            const reader = new FileReader();
            reader.onload = (e) => {
                newProfilePhoto = e.target?.result as string;
                hasCustomPhoto = true;
                isUsingDefault = false;
                hasNewUpload = true;
            };
            reader.readAsDataURL(file);
        }
    }

    // Function to remove uploaded photo
    function removeUploadedPhoto() {
        if (originalProfilePhoto) {
            // If there was an original photo, revert to it
            newProfilePhoto = originalProfilePhoto;
            hasCustomPhoto = true;
        } else {
            // If no original photo, clear everything
            newProfilePhoto = '';
            hasCustomPhoto = false;
        }
        hasNewUpload = false;
        isUsingDefault = false;
        const fileInput = document.getElementById('profile-photo') as HTMLInputElement;
        if (fileInput) fileInput.value = '';
    }

    // Function to use default avatar
    function useDefaultAvatar() {
        newProfilePhoto = '';
        hasCustomPhoto = false;
        isUsingDefault = true;
        hasNewUpload = false;
        const fileInput = document.getElementById('profile-photo') as HTMLInputElement;
        if (fileInput) fileInput.value = '';
    }

    // Function to restore original photo
    function restoreOriginalPhoto() {
        if (originalProfilePhoto) {
            newProfilePhoto = originalProfilePhoto;
            hasCustomPhoto = true;
            isUsingDefault = false;
            hasNewUpload = false;
            const fileInput = document.getElementById('profile-photo') as HTMLInputElement;
            if (fileInput) fileInput.value = '';
        }
    }

    // Save profile changes
    async function saveProfile() {
        isLoading = true;
        try {
            const response = await fetch('http://localhost:8080/user/update', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    avatar: hasCustomPhoto ? newProfilePhoto : '',
                    aboutMe: userDescription,
                    isPrivate: privacySetting === 'private',
                }),
            });

            if (response.ok) {
                // Update original values after successful save
                originalProfilePhoto = newProfilePhoto;
                originalPrivacySetting = privacySetting;
                originalDescription = userDescription;
                showSettingsModal = false;
            } else {
                errorMessage = 'Failed to update profile';
            }
        } catch (error) {
            console.error('Failed to update profile:', error);
            errorMessage = 'Failed to update profile';
        } finally {
            isLoading = false;
        }
    }

    // Watch for changes in URL (userId) or auth state
    $: {
        console.log('URL or auth changed:', { userId: $page.params.id, authUser: $auth.user?.id });
        if ($auth.user && userId) {
            isOwnProfile = $auth.user.id === userId;
            isFollowing = false;
            hasPendingRequest = false;
            canMessage = false;
            
            if (!isOwnProfile) {
                console.log('Loading follow status for user:', userId);
                loadFollowStatus();
                checkCanMessage();
            } else {
                console.log('Own profile detected');
            }
            loadUserPosts();
        }
    }

    // Check if messaging is allowed (if either user follows the other)
    let canMessage = false;
    let checkingMessagePermission = false;

    // Function to load follow status (whether the user is following, has a pending request, or not)
    async function loadFollowStatus() {
        if (isOwnProfile) return;
        
        try {
            const response = await fetch(`http://localhost:8080/user/follow-status`, {
                method: 'POST', // Use POST method
                credentials: 'include', // Include cookies for authentication
                headers: {
                    'Content-Type': 'application/json', // Specify the content type
                },
                body: JSON.stringify({
                    // Include any required data in the request body
                    followedId: userId, // Replace with the actual current user's ID
                }),
            });

            if (response.ok) {
                const followStatus = await response.json();
                console.log('Follow status loaded:', followStatus);
                isFollowing = followStatus.isFollowing;
                hasPendingRequest = followStatus.hasPendingRequest;
            } else {
                console.error('Failed to fetch follow status');
            }
        } catch (error) {
            console.error('Failed to fetch follow status:', error);
        }
    }

    // Function to check if user can message
    async function checkCanMessage() {
        if (isOwnProfile || checkingMessagePermission) return;
        
        checkingMessagePermission = true;
        try {
            const response = await fetch('http://localhost:8080/chat/check-follow', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    userId: userId
                }),
            });
            
            canMessage = response.ok;
        } catch (error) {
            console.error('Failed to check message permission:', error);
            canMessage = false;
        } finally {
            checkingMessagePermission = false;
        }
    }

    // Watch for changes in userId and check message permission
    $: if (userId && !isOwnProfile) {
        checkCanMessage();
    }



    onMount(async () => {
        if ($auth.user && data.user) {
            isOwnProfile = $auth.user.id === userId;
            if (!isOwnProfile) {
                await loadFollowStatus();
                await checkCanMessage();
            }
            await loadUserPosts();
        }
    });

    // Function to add pulse animation to button
    function addPulseAnimation(element: HTMLElement) {
        element.classList.add('button-pulse');
        setTimeout(() => {
            element.classList.remove('button-pulse');
        }, 300);
    }

    // Function to handle follow/unfollow
    async function handleFollow(event: MouseEvent) {
        const button = event.currentTarget as HTMLElement;

        // If already following, show confirmation modal
        if (isFollowing) {
            showUnfollowModal = true;
            return;
        }

        isLoading = true;
        errorMessage = '';
        try {
            addPulseAnimation(button);
            const result = await followers.followUser(userId);
            if (result?.status === 'pending') {
                hasPendingRequest = true;
            } else if (result?.status === 'accepted') {
                isFollowing = true;
            }
        } catch (error) {
            errorMessage = 'Failed to update follow status';
            console.error(errorMessage, error);
        } finally {
            isLoading = false;
            await checkCanMessage();
        }
    }

    // Function to handle unfollow confirmation
    async function handleUnfollow() {
        isLoading = true;
        errorMessage = '';
        try {
            const success = await followers.unfollowUser(userId);
            if (success) {
                isFollowing = false;
            } else {
                errorMessage = 'Failed to unfollow user';
            }
        } catch (error) {
            errorMessage = 'Failed to update follow status';
            console.error(errorMessage, error);
        } finally {
            isLoading = false;
            showUnfollowModal = false;
            await checkCanMessage();
        }
    }

    // Handle message button click
    async function handleMessageClick() {
        if (!canMessage) {
            errorMessage = 'You need to follow each other to send messages';
            return;
        }

        // Since we've already checked permissions with checkCanMessage,
        // we can directly navigate to the chat
        goto(`/chat/${userId}`);
    }

    // Function to show settings modal
    function showSettings() {
        newProfilePhoto = data.user?.avatar ?? "";
        privacySetting = data.user?.isPrivate ? "private" : "public";
        userDescription = data.user?.aboutMe || '';
        hasCustomPhoto = !!data.user?.avatar;
        isUsingDefault = false;
        hasNewUpload = false;
        showSettingsModal = true;
    }

    // Function to update privacy settings and profile photo
    async function updateSettings() {
        try {
            if (!isOwnProfile) {
                return
            }
            const imageToSend = newProfilePhoto === null ? data.user?.avatar : newProfilePhoto || '';

            const response = await fetch(`http://localhost:8080/updateProfile`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    image: imageToSend,
                    description: userDescription,
                    privacy: privacySetting === "private"
                }),
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to update profile');
            }

            showSettingsModal = false;
            location.reload(); // Reload the page to reflect changes
        } catch (error) {
            console.error('Failed to update settings:', error);
        }
    }

    // Function to fetch user posts
    async function loadUserPosts() {
        try {
            debugger
            const response = await fetch(`http://localhost:8080/user/getPosts`, {
                credentials: 'include',
                method: 'POST',
                body: JSON.stringify({
                    userId: userId // Send the userId from the URL params
                }),
            });
            if (response.ok) {
                userPosts = await response.json();
            } else {
                //send to 404 page
                throw error(404, 'User not found');
            }
        } catch (error) {
            console.error('Failed to fetch posts:', error);
        }
    }

    const handleTabChange = (tabId: string) => {
        previousTab = activeTab;
        activeTab = tabId.toLowerCase();
    };
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container mx-auto px-4 py-8">
        <!-- Profile Header - Modern gradient with subtle animation -->
        <div class="rounded-2xl shadow-xl p-8 bg-gradient-to-br from-blue-500 via-blue-600 to-blue-700 text-white transform hover:scale-[1.01] transition-all duration-300">
            <div class="flex flex-col md:flex-row items-center md:items-start md:space-x-10">
                <!-- Avatar with hover effect and border -->
                <div class="relative group">
                    <div class="absolute -inset-0.5 bg-gradient-to-r from-pink-600 to-purple-600 rounded-full opacity-50 group-hover:opacity-100 blur transition duration-300"></div>
                    <div class="relative">
                        <Avatar
                            src={data.user?.avatar || generateAvatar(data.user?.username)}
                            class="w-32 h-32 ring-4 ring-white/50 transform transition-all duration-300 group-hover:scale-105"
                            alt={data.user?.username}
                        />
                    </div>
                </div>

                <!-- Profile Info with improved typography -->
                <div class="flex-1 mt-6 md:mt-0 text-center md:text-left">
                    <div class="flex flex-col md:flex-row md:items-center md:space-x-4">
                        <h1 class="text-3xl font-bold mb-2 md:mb-0">{data.user?.username}</h1>
                        <div class="relative group cursor-help">
                            {#if data.user?.isPrivate}
                                <svg xmlns="http://www.w3.org/2000/svg"
                                     class="h-6 w-6 text-orange-500 dark:text-orange-400 hover:text-orange-600 dark:hover:text-orange-300 transition-colors duration-200"
                                     fill="none"
                                     viewBox="0 0 24 24"
                                     stroke="currentColor"
                                >
                                    <path stroke-linecap="round"
                                          stroke-linejoin="round"
                                          stroke-width="2"
                                          d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                                    />
                                </svg>
                            {:else}
                                <svg xmlns="http://www.w3.org/2000/svg"
                                     class="h-6 w-6 text-emerald-500 dark:text-emerald-400 hover:text-emerald-600 dark:hover:text-emerald-300 transition-colors duration-200"
                                     fill="none"
                                     viewBox="0 0 24 24"
                                     stroke="currentColor"
                                >
                                    <path stroke-linecap="round"
                                          stroke-linejoin="round"
                                          stroke-width="2"
                                          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                                    />
                                    <path stroke-linecap="round"
                                          stroke-linejoin="round"
                                          stroke-width="2"
                                          d="M2.458 12C3.732 7.943 7.522 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.478 0-8.268-2.943-9.542-7z"
                                    />
                                </svg>
                            {/if}
                            <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-3 py-1.5
                                      {data.user?.isPrivate ?
                                        'bg-orange-600 dark:bg-orange-700' :
                                        'bg-emerald-600 dark:bg-emerald-700'}
                                      text-white text-sm rounded-lg opacity-0 group-hover:opacity-100
                                      transition-all duration-200 shadow-lg scale-95 group-hover:scale-100">
                                <div class="relative">
                                    {data.user?.isPrivate ? 'Private Account' : 'Public Account'}
                                    <div class="absolute -bottom-5 left-1/2 transform -translate-x-1/2
                                              border-8 border-transparent
                                              {data.user?.isPrivate ?
                                                'border-t-orange-600 dark:border-t-orange-700' :
                                                'border-t-emerald-600 dark:border-t-emerald-700'}">
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <p class="text-lg text-white/90 mt-3 max-w-2xl">
                        {data.user?.aboutMe || 'No bio added yet'}
                    </p>

                    <!-- Stats with hover effects -->
                    <div class="flex flex-wrap justify-center md:justify-start gap-6 mt-6">
                        <div class="text-center hover:transform hover:scale-105 transition-all duration-200">
                            <div class="text-2xl font-bold">{data.Followers?.length || 0}</div>
                            <div class="text-sm text-white/80">Followers</div>
                        </div>
                        <div class="text-center hover:transform hover:scale-105 transition-all duration-200">
                            <div class="text-2xl font-bold">{data.Following?.length || 0}</div>
                            <div class="text-sm text-white/80">Following</div>
                        </div>
                        <div class="text-center hover:transform hover:scale-105 transition-all duration-200">
                            <div class="text-2xl font-bold">{userPosts?.length || 0}</div>
                            <div class="text-sm text-white/80">Posts</div>
                        </div>
                    </div>

                    <!-- Action Buttons with modern styling -->
                    <div class="flex flex-col sm:flex-row gap-4 mt-6">
                        {#if isOwnProfile}
                            <Button
                                class="bg-blue-500 hover:bg-blue-600 text-white font-semibold py-2 px-6 rounded-lg transform transition-all duration-200 hover:scale-105"
                                color="none"
                                on:click={showSettings}
                            >
                                Edit Profile
                            </Button>
                        {:else}
                            <Button
                                class="relative group overflow-hidden {isFollowing ? 'bg-red-500 hover:bg-red-600' : 'bg-blue-500 hover:bg-blue-600'} text-white font-semibold py-2 px-6 rounded-lg transform transition-all duration-200 hover:scale-105"
                                color="none"
                                disabled={isLoading}
                                on:click={handleFollow}
                            >
                                <span class="relative z-10">
                                    {#if isLoading}
                                        Loading...
                                    {:else if isFollowing}
                                        Unfollow
                                    {:else if hasPendingRequest}
                                        Requested
                                    {:else}
                                        Follow
                                    {/if}
                                </span>
                            </Button>

                            {#if !isOwnProfile}
                                <Button
                                    color="alternative"
                                    disabled={!canMessage || checkingMessagePermission}
                                    on:click={handleMessageClick}
                                    class="flex items-center space-x-2 {canMessage ? 'hover:bg-gray-100 dark:hover:bg-gray-700' : 'opacity-50 cursor-not-allowed'}"
                                >
                                    {#if checkingMessagePermission}
                                        <div class="w-4 h-4 border-2 border-gray-300 border-t-blue-500 rounded-full animate-spin"></div>
                                    {:else}
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                                        </svg>
                                    {/if}
                                    <span>Message</span>
                                </Button>
                            {/if}
                        {/if}
                    </div>
                </div>
            </div>
        </div>

        <!-- Tabs Section with modern styling -->
        <div class="mt-8">
            <div class="border-b border-gray-200 dark:border-gray-700">
                <div class="flex space-x-8">
                    <button
                        class="py-4 px-1 relative {activeTab === 'posts' ? 'text-blue-600 dark:text-blue-500' : 'text-gray-500 dark:text-gray-400'} hover:text-blue-600 dark:hover:text-blue-500 transition-colors duration-200"
                        on:click={() => handleTabChange('posts')}
                    >
                        <span class="text-sm font-medium">Posts</span>
                        {#if activeTab === 'posts'}
                            <div class="absolute bottom-0 left-0 w-full h-0.5 bg-blue-600 dark:bg-blue-500" transition:slide></div>
                        {/if}
                    </button>
                    <button
                        class="py-4 px-1 relative {activeTab === 'followers' ? 'text-blue-600 dark:text-blue-500' : 'text-gray-500 dark:text-gray-400'} hover:text-blue-600 dark:hover:text-blue-500 transition-colors duration-200"
                        on:click={() => handleTabChange('followers')}
                    >
                        <span class="text-sm font-medium">Followers</span>
                        {#if activeTab === 'followers'}
                            <div class="absolute bottom-0 left-0 w-full h-0.5 bg-blue-600 dark:bg-blue-500" transition:slide></div>
                        {/if}
                    </button>
                    <button
                        class="py-4 px-1 relative {activeTab === 'following' ? 'text-blue-600 dark:text-blue-500' : 'text-gray-500 dark:text-gray-400'} hover:text-blue-600 dark:hover:text-blue-500 transition-colors duration-200"
                        on:click={() => handleTabChange('following')}
                    >
                        <span class="text-sm font-medium">Following</span>
                        {#if activeTab === 'following'}
                            <div class="absolute bottom-0 left-0 w-full h-0.5 bg-blue-600 dark:bg-blue-500" transition:slide></div>
                        {/if}
                    </button>
                    {#if isOwnProfile && $followers.requests && $followers.requests.length > 0}
                        <button
                            class="py-4 px-1 relative {activeTab === 'requests' ? 'text-blue-600 dark:text-blue-500' : 'text-gray-500 dark:text-gray-400'} hover:text-blue-600 dark:hover:text-blue-500 transition-colors duration-200"
                            on:click={() => handleTabChange('requests')}
                        >
                            <span class="text-sm font-medium">Follow Requests</span>
                            {#if activeTab === 'requests'}
                                <div class="absolute bottom-0 left-0 w-full h-0.5 bg-blue-600 dark:bg-blue-500" transition:slide></div>
                            {/if}
                        </button>
                    {/if}
                </div>
            </div>

            <div class="relative mt-6" style="min-height: 400px;">
                {#key activeTab}
                    <div
                        class="absolute w-full"
                        in:fly={{
                            x: previousTab === activeTab ? 0 : (previousTab > activeTab ? -100 : 100),
                            duration: 300,
                            easing: quintOut
                        }}
                        out:fly={{
                            x: previousTab === activeTab ? 0 : (previousTab > activeTab ? 100 : -100),
                            duration: 300,
                            easing: quintOut
                        }}
                    >
                        {#if activeTab === 'posts'}
                            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                                {#if userPosts && userPosts.length > 0}
                                    {#each userPosts as post, index}
                                        <div
                                            class="bg-white dark:bg-gray-800 rounded-xl shadow-lg overflow-hidden transform transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
                                            in:fly|local={{
                                                y: 20,
                                                duration: 400,
                                                delay: index * 100,
                                                easing: quintOut
                                            }}
                                        >
                                            <div class="p-6">
                                                <!-- Author Section -->
                                                <div class="flex items-center gap-4 mb-4">
                                                    {#if post?.authorAvatar}
                                                        <img
                                                          src={post.authorAvatar}
                                                          alt={post?.authorName || 'Author'}
                                                          class="w-12 h-12 rounded-full object-cover ring-2 ring-gray-100"
                                                        />
                                                    {:else}
                                                        <div
                                                          class="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-bold text-lg">
                                                            {post?.authorName?.charAt(0) || 'A'}
                                                        </div>
                                                    {/if}
                                                    <div class="flex-1 min-w-0">
                                                        <h3 class="text-lg font-semibold text-gray-900 truncate">{post?.authorName || 'Author Name'}</h3>
                                                        <div class="flex items-center gap-2 text-sm text-gray-500">
                                                            <span>{post?.created_at ? getFormattedDate(new Date(post.created_at)).diff : 'Just now'}</span>
                                                            <span class="inline-flex items-center gap-1 px-2 py-1 rounded-full bg-gray-100">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20"
                                                 fill="currentColor">
                                                {#if post?.privacy === 0}
                                                    <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                                                    <path fill-rule="evenodd"
                                                          d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z"
                                                          clip-rule="evenodd" />
                                                {:else}
                                                    <path fill-rule="evenodd"
                                                          d="M3.707 2.293a1 1 0 00-1.414 1.414l14 14a1 1 0 001.414-1.414l-1.473-1.473A10.014 10.014 0 0019.542 10C18.268 5.943 14.478 3 10 3a9.958 9.958 0 00-4.512 1.074l-1.78-1.781zm4.261 4.26l1.514 1.515a2.003 2.003 0 012.45 2.45l1.514 1.514a4 4 0 00-5.478-5.478z"
                                                          clip-rule="evenodd" />
                                                {/if}
                                            </svg>
                                                                {post?.privacy === 0 ? 'Public' : 'Private'}
                                        </span>
                                                        </div>
                                                    </div>
                                                </div>

                                                <!-- Post Content -->
                                                <div class="space-y-4">
                                                    <h2 class="text-xl font-semibold text-gray-900">{post?.title || 'Untitled'}</h2>
                                                    <p class="text-gray-700 leading-relaxed whitespace-pre-line">{post?.content || ''}</p>

                                                    {#if post?.media}
                                                        <div class="relative w-full overflow-hidden bg-gray-50">
                                                            <img
                                                              src={post.media}
                                                              alt="Post media"
                                                              class="w-full h-auto max-h-[512px] object-contain rounded-lg cursor-zoom-in hover:opacity-95 transition-opacity mx-auto"
                                                              on:click={() => {
                                                expandedImageSrc = post.media || '';
                                                showExpandedImage = true;
                                            }}
                                                            />
                                                        </div>
                                                    {/if}
                                                </div>
                                            </div>

                                        </div>
                                    {/each}
                                {:else}
                                    <div class="col-span-full text-center py-10"
                                         in:fade={{ duration: 200 }}>
                                        <p class="text-gray-500 dark:text-gray-400 text-lg">No posts yet</p>
                                    </div>
                                {/if}
                            </div>
                        {:else if activeTab === 'followers'}
                            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                                {#if data.Followers && data.Followers.length > 0}
                                    {#each data.Followers as follower, index}
                                        <div
                                            class="bg-white dark:bg-gray-800 rounded-xl p-4 shadow-lg flex items-center justify-between transform transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
                                            in:fly|local={{
                                                y: 20,
                                                duration: 400,
                                                delay: index * 100,
                                                easing: quintOut
                                            }}
                                        >
                                            <div class="flex items-center space-x-4">
                                                <Avatar
                                                    src={follower.avatar || generateAvatar(follower.username)}
                                                    class="w-12 h-12"
                                                    alt={follower.username}
                                                />
                                                <div>
                                                    <p class="font-semibold text-gray-800 dark:text-gray-200">{follower.username}</p>
                                                </div>
                                            </div>
                                            <Button
                                                size="sm"
                                                color="blue"
                                                class="transform transition-all duration-200 hover:scale-105"
                                                href="/profile/{follower.userId}"
                                            >
                                                View Profile
                                            </Button>
                                        </div>
                                    {/each}
                                {:else}
                                    <div class="col-span-full text-center py-10"
                                         in:fade={{ duration: 200 }}>
                                        <p class="text-gray-500 dark:text-gray-400 text-lg">No followers yet</p>
                                    </div>
                                {/if}
                            </div>
                        {:else if activeTab === 'following'}
                            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                                {#if data.Following && data.Following.length > 0}
                                    {#each data.Following as following, index}
                                        <div
                                            class="bg-white dark:bg-gray-800 rounded-xl p-4 shadow-lg flex items-center justify-between transform transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
                                            in:fly|local={{
                                                y: 20,
                                                duration: 400,
                                                delay: index * 100,
                                                easing: quintOut
                                            }}
                                        >
                                            <div class="flex items-center space-x-4">
                                                <Avatar
                                                    src={following.avatar || generateAvatar(following.username)}
                                                    class="w-12 h-12"
                                                    alt={following.username}
                                                />
                                                <div>
                                                    <p class="font-semibold text-gray-800 dark:text-gray-200">{following.username}</p>
                                                </div>
                                            </div>
                                            <Button
                                                size="sm"
                                                color="blue"
                                                class="transform transition-all duration-200 hover:scale-105"
                                                href="/profile/{following.userId}"
                                            >
                                                View Profile
                                            </Button>
                                        </div>
                                    {/each}
                                {:else}
                                    <div class="col-span-full text-center py-10"
                                         in:fade={{ duration: 200 }}>
                                        <p class="text-gray-500 dark:text-gray-400 text-lg">Not following anyone yet</p>
                                    </div>
                                {/if}
                            </div>
                        {:else if activeTab === 'requests' && isOwnProfile && $followers.requests && $followers.requests.length > 0}
                            <div class="space-y-4">
                                {#each $followers.requests as request, index}
                                    <div
                                        class="bg-white dark:bg-gray-800 rounded-xl p-4 shadow-lg flex items-center justify-between transform transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
                                        in:fly|local={{
                                            y: 20,
                                            duration: 400,
                                            delay: index * 100,
                                            easing: quintOut
                                        }}
                                    >
                                        <div class="flex items-center space-x-4">
                                            <Avatar
                                                src={request.avatar || generateAvatar(request.username)}
                                                class="w-12 h-12"
                                                alt={request.username}
                                            />
                                            <p class="font-semibold text-gray-800 dark:text-gray-200">{request.username}</p>
                                        </div>
                                        <div class="flex space-x-2">
                                            <Button
                                                size="sm"
                                                color="green"
                                                class="transform transition-all duration-200 hover:scale-105"
                                                on:click={() => followers.handleRequest(request.id, true)}
                                            >
                                                Accept
                                            </Button>
                                            <Button
                                                size="sm"
                                                color="red"
                                                class="transform transition-all duration-200 hover:scale-105"
                                                on:click={() => followers.handleRequest(request.id, false)}
                                            >
                                                Reject
                                            </Button>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/key}
            </div>
        </div>
    </div>
</div>

<!-- Settings Modal - Updated with theme colors -->
<Modal 
    bind:open={showSettingsModal} 
    size="lg" 
    class="dark:bg-gray-800"
    on:close={resetSettings}
>
    <div class="p-6" transition:fade={{ duration: 200 }}>
        <h3 class="text-2xl font-bold mb-6 text-gray-900 dark:text-white">Edit Profile</h3>
        {#if errorMessage}
            <div class="mb-4 p-4 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded-lg">
                {errorMessage}
            </div>
        {/if}
        <div class="space-y-8" transition:slide={{ duration: 300, delay: 150 }}>
            <!-- Profile Photo Update -->
            <div class="space-y-4">
                <label class="block text-lg font-medium text-gray-900 dark:text-white">Profile Photo</label>
                <div class="flex items-center space-x-6">
                    <!-- Current/New Photo Preview -->
                    <div class="relative group w-24 h-24">
                        <div class="absolute -inset-0.5 bg-gradient-to-r from-pink-600 to-purple-600 rounded-full opacity-50 group-hover:opacity-100 blur transition duration-300"></div>
                        <div class="relative rounded-full w-24 h-24 overflow-hidden">
                            <img
                                src={hasCustomPhoto ? newProfilePhoto : generateAvatar(data.user?.username)}
                                alt="Profile"
                                class="w-full h-full object-cover"
                            />
                        </div>
                    </div>

                    <!-- Upload Controls -->
                    <div class="flex-1 space-y-2">
                        <div class="relative">
                            <input
                                id="profile-photo"
                                type="file"
                                accept="image/*"
                                on:change={handleFileUpload}
                                class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
                            />
                            <div class="w-full px-4 py-3 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-600 hover:border-purple-500 dark:hover:border-purple-400 transition-colors duration-200 flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-700">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-400 dark:text-gray-300 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                </svg>
                                <span class="text-sm text-gray-500 dark:text-gray-400">
                                    {hasCustomPhoto ? 'Change photo' : 'Upload photo'}
                                </span>
                            </div>
                        </div>
                        <div class="flex flex-wrap gap-2">
                            {#if hasNewUpload}
                                <Button color="red" size="xs" class="flex-1" on:click={removeUploadedPhoto}>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                    </svg>
                                    Remove uploaded photo
                                </Button>
                            {/if}
                            {#if !isUsingDefault}
                                <Button color="alternative" size="xs" class="flex-1" on:click={useDefaultAvatar}>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                    </svg>
                                    Use default avatar
                                </Button>
                            {:else if originalProfilePhoto}
                                <Button color="alternative" size="xs" class="flex-1" on:click={restoreOriginalPhoto}>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                                    </svg>
                                    Restore original photo
                                </Button>
                            {/if}
                        </div>
                    </div>
                </div>
            </div>

            <!-- Description Update -->
            <div class="space-y-2">
                <label class="block text-lg font-medium text-gray-900 dark:text-white">Bio</label>
                <textarea
                    bind:value={userDescription}
                    placeholder="Tell us about yourself..."
                    class="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-purple-500 focus:border-transparent dark:bg-gray-700 dark:text-white transition-all resize-none"
                    rows="4"
                ></textarea>
            </div>

            <!-- Privacy Settings -->
            <div class="space-y-4">
                <label class="block text-lg font-medium text-gray-900 dark:text-white">Account Privacy</label>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <button
                        class="relative p-4 rounded-lg border-2 {privacySetting === 'public' ? 'border-purple-500 bg-purple-50 dark:bg-purple-900/20' : 'border-gray-200 dark:border-gray-700'} transition-all duration-200"
                        on:click={() => privacySetting = 'public'}
                    >
                        <div class="flex items-center space-x-3">
                            <div class="flex-shrink-0">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-purple-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.522 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.478 0-8.268-2.943-9.542-7z" />
                                </svg>
                            </div>
                            <div class="flex-1 text-left">
                                <p class="font-medium text-gray-900 dark:text-white">Public Account</p>
                                <p class="text-sm text-gray-500 dark:text-gray-400">Anyone can see your profile</p>
                            </div>
                            {#if privacySetting === 'public'}
                                <div class="flex-shrink-0">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-purple-500" viewBox="0 0 20 20" fill="currentColor">
                                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                                    </svg>
                                </div>
                            {/if}
                        </div>
                    </button>

                    <button
                        class="relative p-4 rounded-lg border-2 {privacySetting === 'private' ? 'border-purple-500 bg-purple-50 dark:bg-purple-900/20' : 'border-gray-200 dark:border-gray-700'} transition-all duration-200"
                        on:click={() => privacySetting = 'private'}
                    >
                        <div class="flex items-center space-x-3">
                            <div class="flex-shrink-0">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-purple-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                </svg>
                            </div>
                            <div class="flex-1 text-left">
                                <p class="font-medium text-gray-900 dark:text-white">Private Account</p>
                                <p class="text-sm text-gray-500 dark:text-gray-400">Only approved followers can see your profile</p>
                            </div>
                            {#if privacySetting === 'private'}
                                <div class="flex-shrink-0">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-purple-500" viewBox="0 0 20 20" fill="currentColor">
                                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                                    </svg>
                                </div>
                            {/if}
                        </div>
                    </button>
                </div>
            </div>
        </div>

        <!-- Save Button -->
        <div class="mt-8" transition:slide={{ duration: 300, delay: 300 }}>
            <Button
                on:click={updateSettings}
                class="w-full py-3 bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700 text-white font-medium rounded-lg transition-all duration-200 transform hover:scale-[1.02] focus:ring-4 focus:ring-purple-500/50"
            >
                Save Changes
            </Button>
        </div>
    </div>
</Modal>

<!-- Unfollow Confirmation Modal -->
<Modal bind:open={showUnfollowModal} size="xs">
    <div class="text-center">
        <svg class="mx-auto mb-4 text-gray-400 w-12 h-12 dark:text-gray-200" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 11V6m0 8h.01M19 10a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/>
        </svg>
        <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
            Are you sure you want to unfollow this user?
        </h3>
        <div class="flex justify-center gap-4">
            <Button color="red" on:click={handleUnfollow}>
                Yes, unfollow
            </Button>
            <Button color="alternative" on:click={() => showUnfollowModal = false}>
                No, cancel
            </Button>
        </div>
    </div>
</Modal>

<!-- Expanded Image Modal -->
{#if showExpandedImage}
    <div
      class="fixed inset-0 bg-black/80 z-50 flex items-center justify-center p-4"
      on:click={() => showExpandedImage = false}
    >
        <div class="max-w-4xl w-full">
            <img
              src={expandedImageSrc}
              alt="Expanded Image"
              class="rounded-lg max-h-[90vh] w-auto object-contain"
            />
        </div>
    </div>
{/if}

<style lang="postcss">
    @keyframes pulse {
        0% {
            transform: scale(1);
        }
        50% {
            transform: scale(1.05);
        }
        100% {
            transform: scale(1);
        }
    }

    :global(.button-pulse) {
        animation: pulse 0.3s ease-in-out;
    }

    /* Custom scrollbar styles from your theme */
    :global(.custom-scrollbar) {
        scrollbar-width: thin;
        scrollbar-color: theme('colors.primary.300') theme('colors.gray.100');
    }

    :global(.custom-scrollbar::-webkit-scrollbar) {
        width: 6px;
    }

    :global(.custom-scrollbar::-webkit-scrollbar-track) {
        background: theme('colors.gray.100');
    }

    :global(.custom-scrollbar::-webkit-scrollbar-thumb) {
        background-color: theme('colors.primary.300');
        border-radius: 3px;
    }
</style>
