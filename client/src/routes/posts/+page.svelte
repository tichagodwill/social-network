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
    let isLoadingPosts = true;
    let postsPerRow = 3;

    onMount(async () => {
        isLoadingPosts = true;
        await posts.loadPosts();
        isLoadingPosts = false;

        // Load saved layout preference
        const savedLayout = localStorage.getItem('postsPerRow');
        if (savedLayout) {
            postsPerRow = parseInt(savedLayout);
        }
    });

    // Save preference when changed
    $: {
        if (typeof window !== 'undefined') {
            localStorage.setItem('postsPerRow', postsPerRow.toString());
        }
    }

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

<div class="min-h-screen bg-gradient-to-b from-gray-50 to-white">
    <div class="max-w-[90vw] mx-auto px-4 py-8 pt-20">
        <!-- Feed Header -->
        {#if $auth.isAuthenticated}
            <div class="sticky top-20 z-40 mb-8 bg-white/80 backdrop-blur-lg rounded-2xl shadow-lg p-6 border border-gray-100">
                <div class="flex flex-col sm:flex-row justify-between items-center gap-4">
                    <div class="text-center sm:text-left">
                        <h1 class="text-2xl font-bold text-gray-900">Your Feed</h1>
                        <p class="text-gray-600 mt-1">Share your thoughts with the world</p>
                    </div>

                    <div class="flex items-center gap-4">
                        <!-- Layout Control -->
                        <div class="flex items-center gap-2 bg-gray-50 px-4 py-2 rounded-lg">
                            <span class="text-sm text-gray-600">Posts per row:</span>
                            <select
                                    bind:value={postsPerRow}
                                    class="bg-white border border-gray-200 rounded-md px-2 py-1 text-sm"
                            >
                                <option value={1}>1</option>
                                <option value={2}>2</option>
                                <option value={3}>3</option>
                                <option value={4}>4</option>
                            </select>
                        </div>

                        <Button
                                on:click={() => showModal = true}
                                class="transform transition-all duration-200 hover:scale-105 hover:shadow-lg bg-gradient-to-r from-blue-500 to-blue-600"
                        >
                            <span class="flex items-center gap-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                    <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd"/>
                                </svg>
                                Create Post
                            </span>
                        </Button>
                    </div>
                </div>
            </div>
        {/if}
        <!-- Posts Grid -->
        <div class={`grid gap-6 ${
            postsPerRow === 1 ? 'grid-cols-1' :
            postsPerRow === 2 ? 'grid-cols-1 md:grid-cols-2' :
            postsPerRow === 3 ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3' :
            'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'
        }`}>
            {#if isLoadingPosts}
                <!-- Loading Skeletons -->
                {#each Array(postsPerRow) as _, i}
                    <div class="bg-white rounded-2xl shadow-lg p-6 space-y-4 animate-pulse">
                        <div class="flex items-center gap-4">
                            <div class="w-12 h-12 bg-gray-200 rounded-full"></div>
                            <div class="space-y-2">
                                <div class="h-4 w-32 bg-gray-200 rounded"></div>
                                <div class="h-3 w-24 bg-gray-200 rounded"></div>
                            </div>
                        </div>
                        <div class="space-y-2">
                            <div class="h-4 w-3/4 bg-gray-200 rounded"></div>
                            <div class="h-4 w-full bg-gray-200 rounded"></div>
                        </div>
                    </div>
                {/each}
            {:else if $posts.length === 0}
                <!-- Empty State -->
                <div class="col-span-full text-center py-12">
                    <div class="mx-auto w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mb-4">
                        <svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                        </svg>
                    </div>
                    <h3 class="text-xl font-semibold text-gray-900 mb-2">No posts yet</h3>
                    <p class="text-gray-600">Be the first to share something with your network!</p>
                </div>
            {:else}
                {#each $posts as post (post.id)}
                    <div
                            class="bg-white rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 overflow-hidden h-fit"
                            transition:fly|local={{ y: 20, duration: 300 }}
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
                                    <div class="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-bold text-lg">
                                        {post?.authorName?.charAt(0) || 'A'}
                                    </div>
                                {/if}
                                <div class="flex-1 min-w-0">
                                    <h3 class="text-lg font-semibold text-gray-900 truncate">{post?.authorName || 'Author Name'}</h3>
                                    <div class="flex items-center gap-2 text-sm text-gray-500">
                                        <span>{post?.created_at ? getFormattedDate(new Date(post.created_at)).diff : 'Just now'}</span>
                                        <span class="inline-flex items-center gap-1 px-2 py-1 rounded-full bg-gray-100">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                                                {#if post?.privacy === 1}
                                                    <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                                                    <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
                                                {:else}
                                                    <path fill-rule="evenodd" d="M3.707 2.293a1 1 0 00-1.414 1.414l14 14a1 1 0 001.414-1.414l-1.473-1.473A10.014 10.014 0 0019.542 10C18.268 5.943 14.478 3 10 3a9.958 9.958 0 00-4.512 1.074l-1.78-1.781zm4.261 4.26l1.514 1.515a2.003 2.003 0 012.45 2.45l1.514 1.514a4 4 0 00-5.478-5.478z" clip-rule="evenodd"/>
                                                {/if}
                                            </svg>
                                            {post?.privacy === 1 ? 'Public' : 'Private'}
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
            {/if}
        </div>
    </div>
</div>

<!-- Create Post Modal -->
{#if showModal}
    <div
            class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 overflow-y-auto"
            transition:fade={{ duration: 200 }}
            on:click={handleOutsideClick}
            on:keydown={handleModalKeydown}
    >
        <div class="min-h-screen px-4 py-6 flex items-center justify-center">
            <div
                    class="w-full max-w-2xl relative"
                    role="dialog"
                    aria-labelledby="modal-title"
                    transition:fly={{ y: 20, duration: 300 }}
                    on:click|stopPropagation
            >
                <div class="bg-white rounded-2xl shadow-2xl overflow-hidden">
                    <div class="bg-gradient-to-r from-blue-500 to-blue-600 p-6 sticky top-0 z-10">
                        <div class="flex justify-between items-center">
                            <h3 id="modal-title" class="text-2xl font-semibold text-white">Create New Post</h3>
                            <button
                                    type="button"
                                    class="text-white/80 hover:text-white transition-colors"
                                    on:click={() => showModal = false}
                            >
                                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                                </svg>
                            </button>
                        </div>
                    </div>

                    <div class="max-h-[calc(100vh-8rem)] overflow-y-auto">
                        <form on:submit|preventDefault={handleSubmitPost} class="p-6 space-y-6">
                            <div class="space-y-2">
                                <label for="post-title" class="block text-sm font-medium text-gray-700">
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

                            <div class="space-y-2">
                                <label for="post-content" class="block text-sm font-medium text-gray-700">
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

                            <!-- Image Upload Area -->
                            <div class="relative border-2 border-dashed rounded-lg p-6 transition-all duration-200 hover:border-gray-400">
                                <input
                                        id="media-file"
                                        type="file"
                                        accept="image/*"
                                        on:change={(e) => {
                                        const file = e.target.files?.[0];
                                        if (file && file.size > 5 * 1024 * 1024) {
                                            alert('File size must be less than 5MB');
                                            clearFile();
                                        } else {
                                            mediaFile = file;
                                        }
                                    }}
                                        class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                                />
                                <div class="text-center">
                                    <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/>
                                    </svg>
                                    <p class="mt-1 text-sm text-gray-600">Click to upload or drag and drop</p>
                                    <p class="mt-1 text-xs text-gray-500">PNG, JPG, GIF up to 5MB</p>
                                </div>

                                {#if mediaFile}
                                    <div class="mt-4 relative">
                                        <div class="relative w-full overflow-hidden bg-gray-50">
                                            <img
                                                    src={URL.createObjectURL(mediaFile)}
                                                    alt="Preview"
                                                    class="w-full h-auto max-h-[512px] object-contain rounded-lg mx-auto"
                                                    on:click={handleImageClick}
                                            />
                                            <button
                                                    type="button"
                                                    class="absolute top-2 right-2 p-2 bg-black/50 text-white rounded-full hover:bg-black/70 transition-colors"
                                                    on:click={clearFile}
                                            >
                                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                {/if}
                            </div>

                            <!-- Privacy Settings -->
                            <div class="space-y-2">
                                <label class="block text-sm font-medium text-gray-700">Privacy Setting</label>
                                <div class="flex items-center gap-6 bg-gray-50 p-4 rounded-lg">
                                    <div class="flex items-center gap-2">
                                        <Radio bind:group={privacy} value={1} name="privacy">Public</Radio>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        <Radio bind:group={privacy} value={0} name="privacy">Private</Radio>
                                    </div>
                                </div>
                            </div>

                            <!-- Action Buttons -->
                            <div class="flex justify-end gap-3 pt-6 border-t">
                                <Button
                                        type="button"
                                        color="alternative"
                                        on:click={() => showModal = false}
                                >
                                    Cancel
                                </Button>
                                <Button
                                        type="submit"
                                        disabled={isSubmitting}
                                        class="bg-gradient-to-r from-blue-500 to-blue-600"
                                >
                                    {#if isSubmitting}
                                        <span class="flex items-center gap-2">
                                            <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
                                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                                            </svg>
                                            Creating Post...
                                        </span>
                                    {:else}
                                        Create Post
                                    {/if}
                                </Button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}

<!-- Expanded Image Modal -->
{#if showExpandedImage}
    <div
            class="fixed inset-0 bg-black/90 z-50 overflow-hidden"
            transition:fade={{ duration: 200 }}
            on:click={() => showExpandedImage = false}
    >
        <div class="min-h-screen p-4 flex items-center justify-center">
            <button
                    class="absolute top-4 right-4 text-white p-2 hover:bg-white/10 rounded-full"
                    on:click={() => showExpandedImage = false}
            >
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
            </button>
            <img
                    src={expandedImageSrc}
                    alt="Expanded view"
                    class="max-h-[90vh] max-w-[95vw] object-contain rounded-lg"
                    on:click|stopPropagation
            />
        </div>
    </div>
{/if}

<style lang="postcss">
    /* Custom scrollbar for modern browsers */
    :global(.custom-scrollbar) {
        scrollbar-width: thin;
        scrollbar-color: theme('colors.gray.300') theme('colors.gray.100');
    }

    :global(.custom-scrollbar::-webkit-scrollbar) {
        width: 6px;
    }

    :global(.custom-scrollbar::-webkit-scrollbar-track) {
        background: theme('colors.gray.100');
    }

    :global(.custom-scrollbar::-webkit-scrollbar-thumb) {
        background-color: theme('colors.gray.300');
        border-radius: 3px;
    }

    /* Animation classes */
    :global(.fade-enter) {
        opacity: 0;
    }

    :global(.fade-enter-active) {
        opacity: 1;
        transition: opacity 200ms ease-in;
    }

    :global(.fade-exit) {
        opacity: 1;
    }

    :global(.fade-exit-active) {
        opacity: 0;
        transition: opacity 200ms ease-out;
    }
</style>