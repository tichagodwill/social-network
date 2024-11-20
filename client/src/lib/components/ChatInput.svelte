<script lang="ts">
    import { Input } from 'flowbite-svelte';
    
    export let value = '';
    export let placeholder = '';
    
    let inputElement: any;

    export function focus() {
        inputElement?.element?.focus();
    }

    export function getCursorPosition(): number {
        return inputElement?.element?.selectionStart || 0;
    }
</script>

<Input
    type="text"
    bind:value
    bind:this={inputElement}
    {placeholder}
    class="flex-1"
    on:keypress
/> 