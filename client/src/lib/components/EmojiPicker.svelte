<script lang="ts">
 import { Button } from 'flowbite-svelte';
//  import { FaceSmileOutline } from 'flowbite-svelte-icons';
 import { createEventDispatcher, onMount } from 'svelte';
    import type { EmojiPickerEvent } from '$lib/types';
    import emojiData from '@emoji-mart/data';

    let pickerVisible = false;
    let picker: any;
    const dispatch = createEventDispatcher();

    onMount(async () => {
        const { Picker } = await import('emoji-mart');
        picker = new Picker({
            data: emojiData,
            onEmojiSelect: (emoji: { native: string }) => {
                dispatch('emoji-select', { emoji });
                pickerVisible = false;
            }
        });
    });

    function togglePicker() {
        pickerVisible = !pickerVisible;
    }
</script>

<div class="relative">
    <Button 
        color="alternative" 
        class="!p-2" 
        on:click={togglePicker}
    >
    </Button>

    {#if pickerVisible}
        <div 
            class="absolute bottom-12 right-0 z-50"
            use:picker
        />
    {/if}
</div>

<style>
    :global(em-emoji-picker) {
        --rgb-background: var(--background-rgb);
        --rgb-input: var(--input-rgb);
        --rgb-color: var(--color-rgb);
    }
</style> 