<script lang="ts">
  import { Input } from 'flowbite-svelte'
  import { PaperPlaneSolid } from 'flowbite-svelte-icons'
  import ChatMessage from '$lib/components/chat-message.svelte'

  let messages = [
    { message: 'Let me know if you need anything.', ownMessage: false, createTime: '2024-09-27 09:43:47' },
    { message: 'What time is the meeting today?', ownMessage: false, createTime: '2024-09-29 08:48:56' },
    { message: 'That\'s an interesting perspective.', ownMessage: false, createTime: '2024-09-30 16:37:45' },
    { message: 'Can we reschedule our call?', ownMessage: false, createTime: '2024-10-02 07:55:20' },
    { message: 'I\'ll get back to you on that.', ownMessage: false, createTime: '2024-10-02 18:58:36' },
    { message: 'Did you see the latest update?', ownMessage: true, createTime: '2024-10-03 12:42:52' },
    { message: 'I\'ll send over the documents soon.', ownMessage: true, createTime: '2024-10-04 11:14:33' },
    { message: 'Can you review the report by EOD?', ownMessage: false, createTime: '2024-10-07 21:31:02' },
    { message: 'This project is looking great!', ownMessage: false, createTime: '2024-10-10 14:25:09' },
    { message: 'Hey! How\'s it going?', ownMessage: false, createTime: '2024-10-13 03:42:37' },
  ].sort((a, b) => a.createTime < b.createTime);

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