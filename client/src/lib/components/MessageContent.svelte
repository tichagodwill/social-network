<script lang="ts">
    import type { Message } from '$lib/types';

    export let message: Message;

    $: isImage = message.fileType?.startsWith('image/');
    $: isVideo = message.fileType?.startsWith('video/');
    $: isAudio = message.fileType?.startsWith('audio/');
</script>

{#if message.fileUrl}
    <div class="mb-2">
        {#if isImage}
            <img 
                src={message.fileUrl} 
                alt={message.fileName || 'Image'} 
                class="max-w-sm rounded-lg"
            />
        {:else if isVideo}
            <video 
                src={message.fileUrl} 
                controls 
                class="max-w-sm rounded-lg"
            >
                <track kind="captions">
            </video>
        {:else if isAudio}
            <audio 
                src={message.fileUrl} 
                controls 
                class="w-full"
            >
                <track kind="captions">
            </audio>
        {:else}
            <a 
                href={message.fileUrl}
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center space-x-2 text-primary-600 hover:underline"
            >
                <span>📎</span>
                <span>{message.fileName}</span>
            </a>
        {/if}
    </div>
{/if}
{#if message.content}
    <p>{message.content}</p>
{/if} 