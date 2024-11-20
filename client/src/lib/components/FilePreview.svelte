<script lang="ts">
    import { Button } from 'flowbite-svelte';
    // import { XMarkOutline } from 'flowbite-svelte-icons';
    import type { FilePreview } from '$lib/types';

    export let file: FilePreview;
    export let onCancel: () => void;
</script>

<div class="relative bg-gray-100 dark:bg-gray-800 rounded-lg p-4 mt-2">
    <Button
        size="xs"
        color="red"
        class="absolute -top-2 -right-2"
        on:click={onCancel}
    >
        <!-- <XMarkOutline class="w-3 h-3" /> -->
    </Button>

    <div class="max-w-sm">
        {#if file.type === 'image'}
            <img
                src={file.preview}
                alt={file.file.name}
                class="rounded-lg max-h-48 w-auto"
            />
        {:else if file.type === 'video'}
            <video
                src={file.preview}
                controls
                class="rounded-lg max-h-48 w-auto"
            >
                <track kind="captions">
            </video>
        {:else if file.type === 'audio'}
            <audio
                src={file.preview}
                controls
                class="w-full"
            >
                <track kind="captions">
            </audio>
        {:else}
            <div class="flex items-center space-x-2">
                <span class="text-4xl">📄</span>
                <div>
                    <p class="font-medium">{file.file.name}</p>
                    <p class="text-sm text-gray-500">
                        {(file.file.size / 1024 / 1024).toFixed(2)} MB
                    </p>
                </div>
            </div>
        {/if}
    </div>
</div> 