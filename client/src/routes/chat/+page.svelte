<script>
  import { Card } from 'flowbite-svelte'
  import {AddressBookOutline, CogOutline, MessageCaptionOutline, MessagesOutline, UsersGroupOutline,} from 'flowbite-svelte-icons'
  import ChatEntry from '$lib/components/chat-entry.svelte';
  import ChatRoom from '$lib/components/chat-room.svelte';
</script>

<Card size="none" class="flex-grow flex-row h-full" style="padding: 0px">

  <div class="w-40 flex flex-col justify-center bg-primary-50">
    <div class="grow-[0.333]" />

    <button class="flex hover:bg-gray-200 dark:hover:bg-gray-400 text-black p-2 rounded-lg mx-2 mt-2">
      <MessagesOutline class="w-6 h-6 mr-2" />
      General
    </button>
    <button class="flex hover:bg-gray-200 dark:hover:bg-gray-400 text-black p-2 rounded-lg mx-2 mt-2">
      <MessageCaptionOutline class="w-6 h-6 mr-2" />
      Chat
    </button>
    <button class="flex hover:bg-gray-200 dark:hover:bg-gray-400 text-black p-2 rounded-lg mx-2 mt-2">
      <UsersGroupOutline class="w-6 h-6 mr-2" />
      Groups
    </button>
    <button class="flex hover:bg-gray-200 dark:hover:bg-gray-400 text-black p-2 rounded-lg mx-2 mt-2">
      <AddressBookOutline class="w-6 h-6 mr-2" />
      Contacts
    </button>
    <button class="flex hover:bg-gray-200 dark:hover:bg-gray-400 text-black p-2 rounded-lg mx-2 mt-2">
      <CogOutline class="w-6 h-6 mr-2" />
      Settings
    </button>

    <div class="grow-[1]" />
  </div>

  <div class="w-1/3 grow-0 shrink-[0.2] flex flex-col justify-center gap-2">
    <ChatEntry isSelected={true}></ChatEntry>
    <ChatEntry></ChatEntry>
    <ChatEntry></ChatEntry>
  </div>
  <div class="bg-gray-100 w-ful grow shrink-[0.9]">
    <ChatRoom></ChatRoom>
  </div>
</Card>
