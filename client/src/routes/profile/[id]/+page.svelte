<script lang="ts">
    import { onMount } from 'svelte';
    import { followers } from '$lib/stores/followers';
    import { auth } from '$lib/stores/auth';
    import { goto } from '$app/navigation';
    import { chat } from '$lib/stores/chat';
    import { Button, Avatar, Badge, Tabs, TabItem, Modal, Input, Radio } from 'flowbite-svelte';
    import { fade, slide, fly } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';
    import type { PageData } from './$types';
    import { error } from '@sveltejs/kit';

    export let data: PageData;
    const userId = parseInt(data.params.id);

    let isOwnProfile = false;
    let isFollowing = false;
    let hasPendingRequest = false;
    let isLoading = false;
    let errorMessage = '';
    let showSettingsModal = false;
    let newProfilePhoto: string = data.user?.avatar ?? "";
    let privacySetting: string = data.user?.isPrivate ? "private" : "public";
    let userDescription: string = data.user?.aboutMe || '';
    let userPosts: Array<any> = [];
    let showExpandedImage = false;
    let expandedImageSrc = '';
    let showUnfollowModal = false;
    let activeTab = 'posts';
    let previousTab = 'posts';

    $: if ($auth.user) {
        isOwnProfile = $auth.user.id === userId;
    }

    // Function to load follow status (whether the user is following, has a pending request, or not)
    async function loadFollowStatus() {
        if (isOwnProfile) {
            return
        }
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
                isFollowing = followStatus.isFollowing;
                hasPendingRequest = followStatus.hasPendingRequest;
            } else {
                console.error('Failed to fetch follow status');
            }
        } catch (error) {
            console.error('Failed to fetch follow status:', error);
        }
    }

    // Check if messaging is allowed (if either user follows the other)
    let canMessage = false;

    // Function to check if either user follows the other
    async function checkCanMessage() {
        if (isOwnProfile) {
            return;
        }
        try {
            // Try to get or create a chat - this will fail with 403 if no follow relationship exists
            const response = await fetch(`http://localhost:8080/chat/check-follow`, {
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
        }
    }

    onMount(async () => {
        try {
            // Load follow status
            await loadFollowStatus();
            // Check if messaging is allowed
            await checkCanMessage();
            // Load user posts
            await loadUserPosts();
        } catch (error) {
            console.error('Failed to load data:', error);
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
        }
    }

    // Function to handle opening chat
    async function handleMessageClick() {
        if (!canMessage) {
            errorMessage = 'You need to follow each other to send messages';
            return;
        }

        try {
            // Get or create chat with the user
            const result = await chat.getOrCreateDirectChat(userId);
            if (result?.chatId) {
                goto(`/chat/${result.chatId}`);
            } else if (result?.error) {
                errorMessage = result.error;
            }
        } catch (error) {
            console.error('Failed to open chat:', error);
            errorMessage = 'Failed to open chat';
        }
    }

    // Function to generate avatar with the first letter of the username
    function generateAvatar(username: string): string {
        const firstLetter = username ? username.charAt(0).toUpperCase() : 'U';
        return `https://ui-avatars.com/api/?name=${firstLetter}&background=0ea5e9&color=fff&size=128`;
    }

    // Function to handle file upload for profile photo
    function handleFileUpload(event: Event) {
        const file = (event.target as HTMLInputElement).files?.[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (e) => {
                newProfilePhoto = e.target?.result as string;
            };
            reader.readAsDataURL(file);
        }
    }

    // Function to clear the selected photo
    function clearPhoto() {
        newProfilePhoto = '';
        const fileInput = document.getElementById('profile-photo') as HTMLInputElement;
        if (fileInput) fileInput.value = '';
    }

    // Function to show settings modal
    function showSettings() {
        newProfilePhoto = data.user?.avatar ?? "";
        privacySetting = data.user?.isPrivate ? "private" : "public";
        userDescription = data.user?.aboutMe || '';
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
            const response = await fetch(`http://localhost:8080/getMyPosts`, {
                credentials: 'include'
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
                    {#if !isOwnProfile}
                        <div class="flex flex-col sm:flex-row gap-4 mt-6">
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
                                        Request Pending
                                    {:else}
                                        Follow
                                    {/if}
                                </span>
                            </Button>

                            <div class="relative group">
                                <Button
                                    class="bg-white/10 {canMessage ? 'hover:bg-white/20' : 'opacity-50 cursor-not-allowed'} text-white font-semibold py-2 px-6 rounded-lg transform transition-all duration-200 {canMessage ? 'hover:scale-105' : ''}"
                                    color="none"
                                    on:click={handleMessageClick}
                                    disabled={!canMessage}
                                >
                                    Message
                                </Button>
                                {#if !canMessage}
                                    <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-3 py-1.5
                                                      bg-gray-800 text-white text-sm rounded-lg opacity-0 group-hover:opacity-100
                                                      transition-all duration-200 shadow-lg scale-95 group-hover:scale-100 whitespace-nowrap">
                                        <div class="relative">
                                            You need to follow each other to send messages
                                            <div class="absolute -bottom-5 left-1/2 transform -translate-x-1/2
                                                      border-8 border-transparent
                                                      border-t-gray-800">
                                            </div>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {:else}
                        <Button
                            class="mt-6 bg-white/10 hover:bg-white/20 text-white font-semibold py-2 px-6 rounded-lg transform transition-all duration-200 hover:scale-105"
                            color="none"
                            on:click={showSettings}
                        >
                            Edit Profile
                        </Button>
                    {/if}
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
                    {#if isOwnProfile && data.Requests && data.Requests.length > 0}
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
                                            <div class="p-4">
                                                <p class="text-gray-800 dark:text-gray-200">{post.content}</p>
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
                        {:else if activeTab === 'requests' && isOwnProfile && data.Requests && data.Requests.length > 0}
                            <div class="space-y-4">
                                {#each data.Requests as request, index}
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
<Modal bind:open={showSettingsModal} size="lg" class="dark:bg-gray-800">
    <div class="p-6" transition:fade={{ duration: 200 }}>
        <h3 class="text-2xl font-bold mb-6 text-gray-900 dark:text-white">Edit Profile</h3>
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
                                src={newProfilePhoto || data.user?.avatar || generateAvatar(data.user?.username)}
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
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                </svg>
                                <span class="text-sm text-gray-500 dark:text-gray-400">
                                    {newProfilePhoto ? 'Change photo' : 'Upload photo'}
                                </span>
                            </div>
                        </div>
                        {#if newProfilePhoto}
                            <Button color="red" size="xs" class="w-full" on:click={clearPhoto}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                                Remove photo
                            </Button>
                        {/if}
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
