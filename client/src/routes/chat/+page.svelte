<script lang="ts">
  import { Card } from 'flowbite-svelte'
  import {AddressBookOutline, CogOutline, MessageCaptionOutline, MessagesOutline, UsersGroupOutline} from 'flowbite-svelte-icons'
  import ChatEntry from '$lib/components/chat-entry.svelte'
  import ChatRoom from '$lib/components/chat-room.svelte'
  import OptionButton from '$lib/components/option-button.svelte'

  let selectedName = 'gerenal'
  let handleOptionButton = (event: CustomEvent) => { selectedName = event.detail.text }
</script>

<Card size="none" class="flex-grow flex-row h-full" style="padding: 0px">

  <div class="w-40 flex flex-col justify-center bg-primary-50">
    <div class="grow-[0.333]" />

      <OptionButton name='General' selectedName={selectedName} on:message={handleOptionButton}><MessagesOutline class="w-6 h-6 mr-2" /></OptionButton>
      <OptionButton name='Chat' selectedName={selectedName} on:message={handleOptionButton}><MessageCaptionOutline class="w-6 h-6 mr-2" /></OptionButton>
      <OptionButton name='Groups' selectedName={selectedName} on:message={handleOptionButton}><UsersGroupOutline class="w-6 h-6 mr-2" /></OptionButton>
      <OptionButton name='Contact' selectedName={selectedName} on:message={handleOptionButton}><AddressBookOutline class="w-6 h-6 mr-2" /></OptionButton>
      <OptionButton name='Settings' selectedName={selectedName} on:message={handleOptionButton}><CogOutline class="w-6 h-6 mr-2" /></OptionButton>

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
