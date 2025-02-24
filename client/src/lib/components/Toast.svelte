<!-- src/lib/components/Toast.svelte -->
<script lang="ts">
  import { fade } from 'svelte/transition';
  import { Check, X, AlertCircle, Info } from 'lucide-svelte';
  import { toast } from '$lib/stores/toast';
</script>

{#if $toast.visible}
  <div
    class="fixed bottom-4 right-4 z-50 flex items-center p-4 rounded-lg shadow-lg text-sm font-medium"
    class:bg-green-50={$toast.type === 'success'}
    class:bg-red-50={$toast.type === 'error'}
    class:bg-blue-50={$toast.type === 'info'}
    class:bg-amber-50={$toast.type === 'warning'}
    class:text-green-800={$toast.type === 'success'}
    class:text-red-800={$toast.type === 'error'}
    class:text-blue-800={$toast.type === 'info'}
    class:text-amber-800={$toast.type === 'warning'}
    transition:fade={{ duration: 200 }}
  >
    <div class="mr-2 shrink-0">
      {#if $toast.type === 'success'}
        <Check class="w-5 h-5 text-green-500" />
      {:else if $toast.type === 'error'}
        <X class="w-5 h-5 text-red-500" />
      {:else if $toast.type === 'info'}
        <Info class="w-5 h-5 text-blue-500" />
      {:else if $toast.type === 'warning'}
        <AlertCircle class="w-5 h-5 text-amber-500" />
      {/if}
    </div>
    <div class="flex-1 mr-2">{$toast.message}</div>
    <button
      class="shrink-0 ml-auto rounded-full p-1 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      on:click={() => toast.hide()}
      aria-label="Dismiss"
    >
      <X class="w-4 h-4" />
    </button>
  </div>
{/if}