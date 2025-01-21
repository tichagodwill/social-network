<script lang="ts">
    import {onMount} from 'svelte';
    import {posts} from '$lib/stores/posts';
    import {auth} from '$lib/stores/auth';
    import {Button, Card, Textarea, Radio} from 'flowbite-svelte';
    import {getFormattedDate} from '$lib/dateFormater';
    import {fade, fly} from 'svelte/transition';

    let newPostTitle = '';
    let newPostContent = '';
    let mediaFile: File | null = null;
    let privacy = 1;
    let isSubmitting = false;
    let showModal = false;
    let showExpandedImage = false;
    let expandedImageSrc = '';

    onMount(() => {
        posts.loadPosts();
    });

    function handleModalKeydown(event: KeyboardEvent) {
        if (event.key === 'Escape') {
            showModal = false;
        }
    }

    function handleOutsideClick(event: MouseEvent) {
        if (event.target === event.currentTarget) {
            showModal = false;
        }
    }

    function handleTabKey(event: KeyboardEvent) {
        if (event.key === 'Tab') {
            const modal = event.currentTarget as HTMLElement;
            const focusableElements = modal.querySelectorAll(
                'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
            );
            const firstFocusable = focusableElements[0] as HTMLElement;
            const lastFocusable = focusableElements[focusableElements.length - 1] as HTMLElement;

            if (event.shiftKey && document.activeElement === firstFocusable) {
                event.preventDefault();
                lastFocusable.focus();
            } else if (!event.shiftKey && document.activeElement === lastFocusable) {
                event.preventDefault();
                firstFocusable.focus();
            }
        }
    }

    async function handleSubmitPost() {
        if (!newPostTitle.trim() || !newPostContent.trim()) return;

        isSubmitting = true;
        try {
            let mediaBase64 = '';
            if (mediaFile) {
                mediaBase64 = await convertFileToBase64(mediaFile);
            }

            const postData = {
                title: newPostTitle.trim(),
                content: newPostContent.trim(),
                privacy: Number(privacy),
                media: mediaBase64 || undefined,
                author: $auth?.user?.id || 0
            };

            await posts.addPost(postData);
            await posts.loadPosts();

            newPostTitle = '';
            newPostContent = '';
            mediaFile = null;
            privacy = 1;
            showModal = false;
        } catch (error) {
            console.error('Error adding post:', error);
        } finally {
            isSubmitting = false;
        }
    }

    function convertFileToBase64(file: File): Promise<string> {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => resolve(reader.result as string);
            reader.onerror = (error) => reject(error);
        });
    }

    function clearFile() {
        mediaFile = null;
        const fileInput = document.getElementById('media-file') as HTMLInputElement;
        if (fileInput) fileInput.value = '';
    }

    function handleImageClick() {
        if (mediaFile) {
            expandedImageSrc = URL.createObjectURL(mediaFile);
            showExpandedImage = true;
        }
    }
</script>

<div class="min-h-screen bg-gray-50">
    <div class="max-w-4xl mx-auto px-4 py-8">
        <!-- Posts container with enhanced sticky header -->
        <div class="relative">
            {#if $auth.isAuthenticated}
                <!-- Stylized sticky header -->
                <div class="sticky top-4 z-40 mb-8">
                    <div class="bg-white/80 backdrop-blur-md shadow-lg rounded-full px-6 py-3 flex justify-between items-center border border-gray-100">
                        <div class="text-lg font-medium text-gray-700">Your Feed</div>
                        <Button
                                on:click={() => showModal = true}
                                class="transform transition-all duration-200 hover:scale-105 hover:shadow-md bg-gradient-to-r from-blue-500 to-blue-600 text-white font-medium rounded-full px-6"
                        >
                            <span class="flex items-center gap-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20"
                                     fill="currentColor">
                                    <path fill-rule="evenodd"
                                          d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"
                                          clip-rule="evenodd"/>
                                </svg>
                                Create Post
                            </span>
                        </Button>
                    </div>
                </div>

                <!-- Enhanced Modal with proper accessibility -->
                {#if showModal}
                    <div
                            class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
                            transition:fade={{ duration: 200 }}
                            on:click={handleOutsideClick}
                            on:keydown={handleModalKeydown}
                            role="presentation"
                    >
                        <div
                                class="w-full max-w-2xl"
                                role="dialog"
                                aria-labelledby="modal-title"
                                aria-modal="true"
                                on:keydown={handleTabKey}
                        >
                            <div class="bg-white rounded-2xl shadow-2xl overflow-hidden"
                                 transition:fly={{ y: 20, duration: 300 }}>
                                <div class="bg-gradient-to-r from-blue-500 to-blue-600 p-6">
                                    <div class="flex justify-between items-center">
                                        <h3 id="modal-title" class="text-2xl font-semibold text-white">Create New
                                            Post</h3>
                                        <button
                                                type="button"
                                                class="text-white/80 hover:text-white transition-colors"
                                                on:click={() => showModal = false}
                                                aria-label="Close modal"
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none"
                                                 viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                                      d="M6 18L18 6M6 6l12 12"/>
                                            </svg>
                                        </button>
                                    </div>
                                </div>
                                <div class="p-6">
                                    <form on:submit|preventDefault={handleSubmitPost} class="space-y-6">
                                        <div>
                                            <label for="post-title"
                                                   class="block text-sm font-medium text-gray-700 mb-2">
                                                Post Title
                                            </label>
                                            <input
                                                    id="post-title"
                                                    type="text"
                                                    bind:value={newPostTitle}
                                                    placeholder="Give your post a title"
                                                    class="w-full px-4 py-3 rounded-lg border border-gray-200 focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                                                    required
                                            />
                                        </div>

                                        <div>
                                            <label for="post-content"
                                                   class="block text-sm font-medium text-gray-700 mb-2">
                                                Post Content
                                            </label>
                                            <Textarea
                                                    id="post-content"
                                                    bind:value={newPostContent}
                                                    placeholder="What's on your mind?"
                                                    rows={4}
                                                    class="w-full px-4 py-3 rounded-lg"
                                                    required
                                            />
                                        </div>

                                        <div>
                                            <label for="media-file"
                                                   class="block text-sm font-medium text-gray-700 mb-2">
                                                Upload Image (optional)
                                            </label>
                                            <div class="relative">
                                                <!-- Hidden file input -->
                                                <input
                                                        id="media-file"
                                                        type="file"
                                                        accept="image/*"
                                                        on:change={(e) => {
                                                        const file = e.target.files?.[0];
                                                        if (file && file.size > 5 * 1024 * 1024) { // 5MB limit
                                                            alert('File size must be less than 5MB');
                                                            mediaFile = null;
                                                        } else {
                                                            mediaFile = file;
                                                        }
                                                    }}
                                                        class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                                                />
                                                <!-- Custom styled file upload button -->
                                                <div class="w-full px-4 py-3 rounded-lg border border-gray-200 focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all flex items-center justify-between bg-white hover:bg-gray-50">
                                                    <span class="text-gray-500">
                                                        {#if mediaFile}
                                                            {mediaFile.name}
                                                        {:else}
                                                            Choose an image...
                                                        {/if}
                                                    </span>
                                                    <span class="text-blue-600 font-medium">Browse</span>
                                                </div>
                                            </div>
                                            <!-- Clear button (visible only when a file is selected) -->
                                            {#if mediaFile}
                                                <div class="mt-2">
                                                    <Button on:click={clearFile} color="red" size="sm"
                                                            class="w-full sm:w-auto">
                                                        Clear Image
                                                    </Button>
                                                </div>
                                            {/if}
                                            <!-- Image preview -->
                                            {#if mediaFile}
                                                <div class="mt-4">
                                                    <img
                                                            src={URL.createObjectURL(mediaFile)}
                                                            alt="Preview"
                                                            class="mt-2 rounded-lg max-h-32 w-auto object-cover cursor-pointer hover:opacity-80 transition-opacity"
                                                            on:click={handleImageClick}
                                                    />
                                                </div>
                                            {/if}
                                        </div>

                                        <div>
                                            <label class="block text-sm font-medium text-gray-700 mb-3">
                                                Privacy Setting
                                            </label>
                                            <div class="flex items-center gap-6 bg-gray-50 p-4 rounded-lg">
                                                <div class="flex items-center gap-2">
                                                    <Radio bind:group={privacy} value={1} name="privacy">Public</Radio>
                                                </div>
                                                <div class="flex items-center gap-2">
                                                    <Radio bind:group={privacy} value={0} name="privacy">Private</Radio>
                                                </div>
                                            </div>
                                        </div>

                                        <div class="flex justify-end gap-3 pt-6 border-t">
                                            <Button type="button" color="alternative"
                                                    on:click={() => showModal = false}>
                                                Cancel
                                            </Button>
                                            <Button type="submit" disabled={isSubmitting}
                                                    class="bg-gradient-to-r from-blue-500 to-blue-600 text-white">
                                                {isSubmitting ? 'Posting...' : 'Create Post'}
                                            </Button>
                                        </div>
                                    </form>
                                </div>
                            </div>
                        </div>
                    </div>
                {/if}
            {/if}

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

            <!-- Posts List -->
            <div class="space-y-6">
                {#each $posts as post (post.id)}
                    <Card class="border-0 shadow-lg hover:shadow-xl transition-shadow duration-300">
                        <div class="flex justify-between items-start">
                            <div class="flex items-center gap-4">
                                <!-- Author Avatar -->
                                {#if post?.authorAvatar}
                                    <img
                                            src={post.authorAvatar}
                                            alt="Author Avatar"
                                            class="w-12 h-12 rounded-full object-cover"
                                    />
                                {:else}
                                    <div class="w-12 h-12 rounded-full bg-blue-500 flex items-center justify-center text-white font-semibold text-lg">
                                        {post?.authorName?.charAt(0) || 'A'}
                                    </div>
                                {/if}
                                <div>
                                    <h3 class="text-xl font-semibold mb-2 text-gray-800">{post?.authorName || 'Author Name'}</h3>
                                    <div class="flex items-center gap-3 text-sm text-gray-500">
                                        <span>{post?.created_at ? getFormattedDate(new Date(post.created_at)).diff : 'Just now'}</span>
                                        <span class="inline-flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"
                     aria-hidden="true">
                    {#if post?.privacy === 1}
                        <path fill-rule="evenodd" d="M10 12a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd"/>
                        <path fill-rule="evenodd"
                              d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z"
                              clip-rule="evenodd"/>
                    {:else}
                        <path fill-rule="evenodd"
                              d="M3.707 2.293a1 1 0 00-1.414 1.414l14 14a1 1 0 001.414-1.414l-1.473-1.473A10.014 10.014 0 0019.542 10C18.268 5.943 14.478 3 10 3a9.958 9.958 0 00-4.512 1.074l-1.78-1.781zm4.261 4.26l1.514 1.515a2.003 2.003 0 012.45 2.45l1.514 1.514a4 4 0 00-5.478-5.478z"
                              clip-rule="evenodd"/>
                        <path d="M12.454 16.697L9.75 13.992a4 4 0 01-3.742-3.741L2.335 6.578A9.98 9.98 0 00.458 10c1.274 4.057 5.065 7 9.542 7 .847 0 1.669-.105 2.454-.303z"/>
                    {/if}
                </svg>
                                            {post?.privacy === 1 ? 'Public' : 'Private'}
            </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <h3 class="text-xl font-semibold mt-4 text-gray-800">{post?.title || 'Untitled'}</h3>
                        <p class="mt-4 text-gray-700 leading-relaxed">{post?.content || ''}</p>
                        {#if post?.media}
                            <img src={post.media} alt="Post media"
                                 class="mt-4 rounded-lg max-h-96 w-auto object-cover shadow-md cursor-pointer hover:opacity-80 transition-opacity"
                                 on:click={() => { expandedImageSrc = post.media; showExpandedImage = true; }}/>
                        {/if}
                    </Card>
                {/each}
            </div>
        </div>
    </div>
</div>