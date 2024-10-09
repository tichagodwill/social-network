<script lang="ts">

  export let ownMessage = false
  export let message = '...'

  const messageClass = ownMessage ? 'ml-auto px-2' : 'mr-auto px-2'

</script>


<span class="mt-2 p-0.5 bg-white rounded-lg shadow-sm text-black {messageClass}">
  {message}
</span>