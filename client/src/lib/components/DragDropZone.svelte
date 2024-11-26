<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import type { FileUploadResponse } from '$lib/types';

    export let active = false;
    const dispatch = createEventDispatcher<{
        upload: FileUploadResponse;
    }>();

    let dragging = false;
    let uploading = false;

    function handleDragEnter(e: DragEvent) {
        e.preventDefault();
        dragging = true;
    }

    function handleDragLeave(e: DragEvent) {
        e.preventDefault();
        dragging = false;
    }

    function handleDragOver(e: DragEvent) {
        e.preventDefault();
    }

    async function handleDrop(e: DragEvent) {
        e.preventDefault();
        dragging = false;

        const files = e.dataTransfer?.files;
        if (!files || files.length === 0) return;

        const file = files[0];
        if (file.size > 10 * 1024 * 1024) { // 10MB limit
            alert('File size must be less than 10MB');
            return;
        }

        uploading = true;
        try {
            const formData = new FormData();
            formData.append('file', file);

            const response = await fetch('http://localhost:8080/upload', {
                method: 'POST',
                credentials: 'include',
                body: formData
            });

            if (!response.ok) {
                throw new Error('Upload failed');
            }

            const data = await response.json();
            dispatch('upload', {
                url: data.url,
                fileName: file.name,
                fileType: file.type
            });
        } catch (error) {
            console.error('File upload failed:', error);
            alert('Failed to upload file');
        } finally {
            uploading = false;
        }
    }
</script>

{#if active}
    <div
        role="presentation"
        class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center transition-opacity"
        class:opacity-0={!dragging}
        class:pointer-events-none={!dragging}
        on:dragenter={handleDragEnter}
        on:dragleave={handleDragLeave}
        on:dragover|preventDefault
        on:drop={handleDrop}
    >
        <div class="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-lg text-center">
            {#if uploading}
                <p class="text-lg">Uploading...</p>
            {:else}
                <p class="text-lg">Drop your file here</p>
                <p class="text-sm text-gray-500 mt-2">Maximum file size: 10MB</p>
            {/if}
        </div>
    </div>
{/if}

<div
    role="presentation"
    class="min-h-full"
    on:dragenter={handleDragEnter}
    on:dragleave={handleDragLeave}
    on:dragover|preventDefault
    on:drop={handleDrop}
>
    <slot />
</div> 