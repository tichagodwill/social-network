<script lang="ts">
    import { Button } from 'flowbite-svelte';
    import { PaperClipOutline } from 'flowbite-svelte-icons';
    import { createEventDispatcher } from 'svelte';
    import type { FileUploadResponse, FilePreview } from '$lib/types';
    import { createFilePreview, uploadFile } from '$lib/utils/fileUpload';
    import FilePreview from './FilePreview.svelte';

    const dispatch = createEventDispatcher<{
        upload: FileUploadResponse;
    }>();

    let uploading = false;
    let inputElement: HTMLInputElement;
    let filePreview: FilePreview | null = null;

    async function handleFileSelect(event: Event) {
        const target = event.target as HTMLInputElement;
        const files = target.files;
        
        if (!files || files.length === 0) return;
        
        const file = files[0];
        if (file.size > 10 * 1024 * 1024) {
            alert('File size must be less than 10MB');
            return;
        }

        filePreview = await createFilePreview(file);
    }

    async function handleUpload() {
        if (!filePreview) return;
        
        uploading = true;
        try {
            const response = await uploadFile(filePreview.file);
            dispatch('upload', response);
            filePreview = null;
        } catch (error) {
            console.error('File upload failed:', error);
            alert('Failed to upload file');
        } finally {
            uploading = false;
            if (inputElement) {
                inputElement.value = '';
            }
        }
    }

    function cancelUpload() {
        filePreview = null;
        if (inputElement) {
            inputElement.value = '';
        }
    }
</script>

<div>
    <input
        type="file"
        class="hidden"
        bind:this={inputElement}
        on:change={handleFileSelect}
        accept="image/*,video/*,audio/*,.pdf,.doc,.docx,.txt"
    />
    <Button
        color="alternative"
        class="!p-2"
        disabled={uploading}
        on:click={() => inputElement.click()}
    >
        <PaperClipOutline class="w-5 h-5" />
    </Button>

    {#if filePreview}
        <div class="absolute bottom-20 left-0 right-0 px-4">
            <FilePreview
                file={filePreview}
                onCancel={cancelUpload}
            />
            <div class="mt-2 flex justify-end space-x-2">
                <Button
                    size="sm"
                    color="alternative"
                    on:click={cancelUpload}
                >
                    Cancel
                </Button>
                <Button
                    size="sm"
                    on:click={handleUpload}
                    disabled={uploading}
                >
                    {uploading ? 'Sending...' : 'Send'}
                </Button>
            </div>
        </div>
    {/if}
</div> 