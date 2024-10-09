<script lang="ts">
  import { Input } from 'flowbite-svelte'
  import { PaperPlaneSolid } from 'flowbite-svelte-icons'
  import ChatMessage from '$lib/components/chat-message.svelte'

  let messages = [
    { message: 'Hey! How\'s it going?', ownMessage: false },
    { message: 'What time is the meeting today?', ownMessage: false },
    { message: 'Did you see the latest update?', ownMessage: true },
    { message: 'I\'ll get back to you on that.', ownMessage: false },
    { message: 'Can we reschedule our call?', ownMessage: false },
    { message: 'This project is looking great!', ownMessage: false },
    { message: 'Let me know if you need anything.', ownMessage: false },
    { message: 'I\'ll send over the documents soon.', ownMessage: true },
    { message: 'Can you review the report by EOD?', ownMessage: false },
    { message: 'That\'s an interesting perspective.', ownMessage: false },
  ]
</script>

<div class="flex flex-col h-full">
  <!-- message box -->
  <div class="flex flex-col-reverse grow overflow-y-scroll mt-2 mx-12">
    {#each messages as message}
      <ChatMessage {...message}></ChatMessage>
    {/each}
  </div>

  <!-- input box -->
  <div class="mx-12 my-3">
    <div class="flex flex-row align-middle">
      <Input type="text" id="input" placeholder="Message..." required size="lg" color="white" />
      <PaperPlaneSolid class="w-10 h-10 rotate-90 my-auto ml-2 p-2 bg-blue-500 rounded-full" color="white" />
    </div>
  </div>
</div>