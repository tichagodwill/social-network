<script lang="ts">

  import { getLastDate } from '$lib/dateFormater'

  export let ownMessage = false
  export let message = '...'
  export let createTime = new Date()

  const messageClass = ownMessage ? 'flex-row-reverse ml-auto px-2' : 'flex-row mr-auto px-2'
  const messageDate = getLastDate(new Date(createTime))
</script>

<span class="message flex gap-2 mt-2 p-0.5 bg-white rounded-lg shadow-sm text-black {messageClass}">
  <span>{message}</span>
  <span class="info text-sm text-gray-500 mt-auto">{messageDate}</span>
</span>

<style>
    .message .info {
        min-width: 0;
        width: 0;
        height: 0;
        visibility: hidden;
        opacity: 0;
        overflow: hidden; /* Ensure smooth transition without affecting layout */
        transition: 400ms ease-in-out;
    }

    .message:hover .info {
        min-width: 2em;
        width: revert;
        height: revert;
        visibility: revert;
        opacity: revert;
    }
</style>