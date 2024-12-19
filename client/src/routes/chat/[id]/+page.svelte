<script lang="ts">
  export let data;
  import Chat from '../+page.svelte';
</script>

<Chat loadContact={+data.id} />