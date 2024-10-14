<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let name = 'general'
  export let selectedName = ''
  export let selectedClasses = 'bg-gray-200 dark:bg-gray-400'

  $: extraClasses = (selectedName.toLowerCase() == name.toLowerCase()) ? selectedClasses : ''

  const dispatch = createEventDispatcher();
  function clicked() {
    dispatch('message', { text: name.toLowerCase() })
  }
</script>

<button class="flex hover:bg-gray-200 dark:hover:bg-gray-400 text-black p-2 rounded-lg mx-2 mt-2 {extraClasses}" on:click={clicked}>
  <slot/>{name}
</button>