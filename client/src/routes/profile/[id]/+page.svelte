<script lang="ts">
    import { onMount } from 'svelte';
    import { followers } from '$lib/stores/followers';
    import { auth } from '$lib/stores/auth';
    import { Button, Avatar, Badge, Tabs, TabItem, Modal, Input, Radio } from 'flowbite-svelte';
    import type { PageData } from './$types';
    import {error} from "@sveltejs/kit";

    export let data: PageData;
    const userId = parseInt(data.params.id);

    let isOwnProfile = false;
    let isFollowing = false;
    let hasPendingRequest = false;
    let isLoading = false;
    let errorMessage = '';
    let showSettingsModal = false;
    let newProfilePhoto: string = data.user?.avatar ?? "";
    let privacySetting: boolean = data.user?.isPrivate === true;
    let userDescription: string = data.user?.aboutMe || '';
    let userPosts: Array<any> = [];
    let showExpandedImage = false;
    let expandedImageSrc = '';

    $: if ($auth.user) {
        isOwnProfile = $auth.user.id === userId;
    }

    // Function to load follow status (whether the user is following, has a pending request, or not)
    async function loadFollowStatus() {
       if(isOwnProfile){
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
                debugger
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
    onMount(async () => {
        try {
            // Load follow status
            await loadFollowStatus();
            // Load user posts
            await loadUserPosts();
        } catch (error) {
            console.error('Failed to load data:', error);
        }
    });

    // Function to handle follow/unfollow
    async function handleFollow() {
        isLoading = true;
        errorMessage = '';
        try {
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
        privacySetting = data.user?.isPrivate === true;
        userDescription = data.user?.aboutMe || '';
        showSettingsModal = true;
    }

    // Function to update privacy settings and profile photo
    async function updateSettings() {
        try {
            if(!isOwnProfile){
                return
            }
            const imageToSend = newProfilePhoto === null ? data.user?.avatar : newProfilePhoto || '';

            const response = await fetch(`http://localhost:8080/updateProfile`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    image: imageToSend,
                    description: userDescription,
                    privacy: privacySetting
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
            const response = await fetch(`http://localhost:8080/user/${userId}/posts`, {
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
</script>

<div class="container mx-auto px-4 py-24">
    <!-- Profile Header -->
    <div class="rounded-lg shadow-lg p-8 bg-gradient-to-r from-[rgba(239,86,47,1)] to-[rgba(239,86,47,0.8)] text-white">
        <div class="flex flex-col md:flex-row items-center md:items-start md:space-x-8">
            <!-- Avatar -->
            <div class="relative">
                <Avatar
                  src={data.user?.avatar || generateAvatar(data.user?.username)}
                  size="xl"
                  alt="User Avatar"
                  on:click={() => {
                        expandedImageSrc = data.user?.avatar || generateAvatar(data.user?.username);
                        showExpandedImage = true;
                    }}
                  class="cursor-pointer hover:opacity-80 transition-opacity border-4 border-white shadow-lg"
                />
                {#if isOwnProfile}
                    <button
                      class="absolute bottom-0 right-0 p-2 bg-white rounded-full shadow-md hover:bg-gray-100 transition"
                      on:click={() => (showSettings())}
                      aria-label="Settings"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-orange-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37a1.724 1.724 0 002.572-1.065z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                    </button>
                {/if}
            </div>

            <!-- User Details -->
            <div class="flex-1 text-center md:text-left mt-6 md:mt-0">
                <div class="flex items-center justify-center md:justify-start space-x-3">
                    <h1 class="text-4xl font-extrabold">{data.user?.username}</h1>
                    <div class="relative group">
                        {#if privacySetting}
                            <!-- Closed Eye Icon for Private -->
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-200 cursor-pointer" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                            </svg>
                        {:else}
                            <!-- Open Eye Icon for Public -->
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-200 cursor-pointer" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.522 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.478 0-8.268-2.943-9.542-7z" />
                            </svg>
                        {/if}
                        <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-3 py-1 bg-black text-white text-sm rounded opacity-0 group-hover:opacity-100 transition-opacity">
                            {privacySetting ? 'Private' : 'Public'}
                        </div>
                    </div>
                </div>
                {#if userDescription}
                <p class="text-gray-200 dark:text-gray-300 mt-2">{userDescription}</p>
                {/if}
                {#if data.user?.firstName || data.user?.lastName}
                <p class="text-gray-200 dark:text-gray-300 mt-2">
                    {data.user?.firstName} {data.user?.lastName}
                </p>
                {/if}
                {#if data.user?.email}
                    <p class="text-gray-200 dark:text-gray-300 mt-2">
                        {data.user?.email}
                    </p>
                {/if}
                {#if data.user?.dateOfBirth}
                <p class="text-gray-200 dark:text-gray-300 mt-2">
                    Date of Birth: {data.user?.dateOfBirth}
                </p>
                {/if}
            </div>

            <!-- Follow Button -->
            {#if !isOwnProfile}
                <Button
                  class="mt-6 md:mt-0 transition-transform hover:scale-105"
                  color={isFollowing ? 'alternative' : 'primary'}
                  disabled={hasPendingRequest || isLoading}
                  on:click={handleFollow}
                  aria-label="Follow/Unfollow Button"
                >
                    {#if isLoading}
                        <span class="flex items-center">
                            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            Processing...
                        </span>
                    {:else if hasPendingRequest}
                        <Badge color="yellow">Request Pending</Badge>
                    {:else if isFollowing}
                        <Badge color="green">Following</Badge>
                    {:else}
                        Follow
                    {/if}
                </Button>
                {#if errorMessage}
                    <p class="text-red-500 text-sm mt-2">{errorMessage}</p>
                {/if}
            {/if}
        </div>
    </div>

    <!-- Tabs Section -->
    <Tabs class="mt-8">
        <TabItem title="Followers" active>
            <div class="rounded-lg shadow-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-6">
                <h3 class="text-2xl font-semibold mb-4">Followers</h3>
                {#if data.Followers && data.Followers.length > 0}
                    <div class="space-y-4">
                        {#each data.Followers as Followers}
                            <div class="flex items-center space-x-4 hover:bg-gray-100 dark:hover:bg-gray-700 p-4 rounded-lg transition">
                                <Avatar src={Followers.avatar || generateAvatar(Followers.username)} alt="Following Avatar" />
                                <div>
                                    <p class="font-semibold text-lg">{Followers.username}</p>
                                    <p class="text-sm text-gray-600 dark:text-gray-400">
                                        {Followers.firstName} {Followers.lastName}
                                    </p>
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="text-gray-500 dark:text-gray-400">Not following anyone yet.</p>
                {/if}
            </div>
        </TabItem>

        <TabItem title="Following">
            <div class="rounded-lg shadow-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-6">
                <h3 class="text-2xl font-semibold mb-4">Following</h3>
                {#if data.Following && data.Following.length > 0}
                    <div class="space-y-4">
                        {#each data.Following as following}
                            <div class="flex items-center space-x-4 hover:bg-gray-100 dark:hover:bg-gray-700 p-4 rounded-lg transition">
                                <Avatar src={following.avatar || generateAvatar(following.username)} alt="Following Avatar" />
                                <div>
                                    <p class="font-semibold text-lg">{following.username}</p>
                                    <p class="text-sm text-gray-600 dark:text-gray-400">
                                        {following.firstName} {following.lastName}
                                    </p>
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="text-gray-500 dark:text-gray-400">Not following anyone yet.</p>
                {/if}
            </div>
        </TabItem>

        <!-- My Posts Tab -->
        <TabItem title="My Posts">
            <div class="rounded-lg shadow-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-6">
                <h3 class="text-2xl font-semibold mb-4">My Posts</h3>
                {#if userPosts && userPosts.length > 0}
                    <div class="space-y-4">
                        {#each userPosts as post}
                            <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700">
                                <p class="font-semibold text-lg">{post.title}</p>
                                <p class="text-sm text-gray-600 dark:text-gray-400">{post.content}</p>
                                {#if post.media}
                                    <img
                                      src={post.media}
                                      alt="Post media"
                                      class="mt-4 rounded-lg max-h-96 w-auto object-cover shadow-md cursor-pointer hover:opacity-80 transition-opacity"
                                      on:click={() => {
                                            expandedImageSrc = post.media;
                                            showExpandedImage = true;
                                        }}
                                    />
                                {/if}
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="text-gray-500 dark:text-gray-400">No posts yet.</p>
                {/if}
            </div>
        </TabItem>

        {#if isOwnProfile && data.Requests && data.Requests.length > 0}
            <TabItem title="Follow Requests">
                <div class="rounded-lg shadow-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-6">
                    <h3 class="text-2xl font-semibold mb-4">Follow Requests</h3>
                    {#each data.Requests as request}
                        <div class="flex items-center justify-between hover:bg-gray-100 dark:hover:bg-gray-700 p-4 rounded-lg transition">
                            <div class="flex items-center space-x-4">
                                <Avatar src={request.avatar|| generateAvatar(request.username)} alt="Request Avatar" />
                                <div>
                                    <p class="font-semibold text-lg">{request.username}</p>
                                    <p class="text-sm text-gray-600 dark:text-gray-400">
                                        {request.firstName} {request.lastName}
                                    </p>
                                </div>
                            </div>
                            <div class="space-x-2">
                                <Button
                                        size="sm"
                                        color="primary"
                                        on:click={() => followers.handleRequest(request.id, true)}
                                        aria-label="Accept Request Button"
                                >
                                    Accept
                                </Button>
                                <Button
                                        size="sm"
                                        color="alternative"
                                        on:click={() => followers.handleRequest(request.id, false)}
                                        aria-label="Decline Request Button"
                                >
                                    Decline
                                </Button>
                            </div>
                        </div>
                    {/each}
                </div>
            </TabItem>
        {/if}
    </Tabs>

    <!-- Settings Modal -->
    <Modal bind:open={showSettingsModal} title="Settings">
        <div class="space-y-6">
            <!-- Profile Photo Update -->
            <div>
                <label class="block text-sm font-medium mb-2">Profile Photo</label>
                <div class="relative">
                    <input
                      id="profile-photo"
                      type="file"
                      accept="image/*"
                      on:change={handleFileUpload}
                      class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                    />
                    <div class="w-full px-4 py-3 rounded-lg border border-gray-200 focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all flex items-center justify-between bg-white hover:bg-gray-50">
                        <span class="text-gray-500">
                            {#if newProfilePhoto}
                                New Photo Selected
                            {:else}
                                Choose an image...
                            {/if}
                        </span>
                        <span class="text-blue-600 font-medium">Browse</span>
                    </div>
                </div>
                {#if newProfilePhoto}
                    <div class="mt-4">
                        <img
                          src={newProfilePhoto}
                          alt="Profile Photo"
                          class="w-20 h-20 rounded-full cursor-pointer hover:opacity-80 transition-opacity"
                          on:click={() => {
                                expandedImageSrc = newProfilePhoto;
                                showExpandedImage = true;
                            }}
                        />
                        <Button on:click={clearPhoto} color="red" size="sm" class="mt-2">
                            Clear Image
                        </Button>
                    </div>
                {/if}
            </div>

            <!-- Description Update -->
            <div>
                <label class="block text-sm font-medium mb-2">Description</label>
                <Input
                  type="text"
                  bind:value={userDescription}
                  placeholder="Enter your description"
                  class="w-full"
                />
            </div>

            <!-- Privacy Settings -->
            <div>
                <label class="block text-sm font-medium mb-2">Privacy Setting</label>
                <div class="flex items-center gap-6 bg-gray-50 p-4 rounded-lg">
                    <div class="flex items-center gap-2">
                        <Radio bind:group={privacySetting} value={false} name="privacy">Public</Radio>
                    </div>
                    <div class="flex items-center gap-2">
                        <Radio bind:group={privacySetting} value={true} name="privacy">Private</Radio>
                    </div>
                </div>
            </div>

            <!-- Save Button -->
            <Button on:click={updateSettings} class="w-full bg-gradient-to-r from-blue-500 to-blue-600 text-white">
                Save Changes
            </Button>
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
</div>