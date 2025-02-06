<script lang="ts">
    import type { Message } from '$lib/types';

    export let message: Message;

    $: isImage = message.fileType?.startsWith('image/');
    $: isVideo = message.fileType?.startsWith('video/');
    $: isAudio = message.fileType?.startsWith('audio/');

    let imageLoaded = false;
</script>

{#if message.fileUrl}
    <div class="mb-2">
        {#if isImage}
            <div class="relative">
                <div class="absolute inset-0 bg-gray-200 dark:bg-gray-700 rounded-lg" class:hidden={imageLoaded}></div>
                <img 
                    src={message.fileUrl} 
                    alt={message.fileName || 'Image'} 
                    class="max-w-sm rounded-lg cursor-zoom-in hover:opacity-95"
                    on:click={() => window.open(message.fileUrl, '_blank')}
                    on:load={() => imageLoaded = true}
                />
            </div>
        {:else if isVideo}
            <video 
                src={message.fileUrl} 
                controls 
                class="max-w-sm rounded-lg"
                controlsList="nodownload"
            >
                <track kind="captions">
            </video>
        {:else if isAudio}
            <div class="bg-black/5 dark:bg-white/5 rounded-lg p-2">
                <audio 
                    src={message.fileUrl} 
                    controls 
                    class="w-full"
                    controlsList="nodownload"
                >
                    <track kind="captions">
                </audio>
            </div>
        {:else}
            <div class="bg-black/5 dark:bg-white/5 rounded-lg p-2 hover:bg-black/10 dark:hover:bg-white/10">
                <a 
                    href={message.fileUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="flex items-center space-x-2"
                >
                    <div class="bg-black/10 dark:bg-white/10 rounded p-1.5">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                            <path fill-rule="evenodd" d="M8 4a3 3 0 00-3 3v4a5 5 0 0010 0V7a1 1 0 112 0v4a7 7 0 11-14 0V7a5 5 0 0110 0v4a3 3 0 11-6 0V7a1 1 0 012 0v4a1 1 0 102 0V7a3 3 0 00-3-3z" clip-rule="evenodd" />
                        </svg>
                    </div>
                    <div class="min-w-0">
                        <p class="font-medium truncate">{message.fileName}</p>
                        <p class="text-xs opacity-75">Click to download</p>
                    </div>
                </a>
            </div>
        {/if}
    </div>
{/if}
{#if message.content}
    <p class="whitespace-pre-wrap break-words">{message.content}</p>
{/if}