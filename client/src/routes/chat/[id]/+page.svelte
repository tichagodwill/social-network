<script lang="ts">
  export let data;
  import { onMount } from 'svelte';
  import { chat } from '$lib/stores/chat';
  import { auth } from '$lib/stores/auth';
  import Chat from '../+page.svelte';
  
  const userId = $auth.user?.id;
  const targetContactId = +data.id;

  onMount(async () => {
    if (!userId || !targetContactId) return;
    
    try {
      // Initialize chat and load contacts
      await chat.initialize();
      await chat.loadContacts(userId);
      
      // Try to create/get direct chat
      const result = await chat.getOrCreateDirectChat(targetContactId);
      if (result.error) {
        console.error('Failed to create chat:', result.error);
        return;
      }
      
      // Reload contacts to ensure we have the latest list
      await chat.loadContacts(userId);
      
      // Load messages for this contact
      await chat.loadMessages(userId, targetContactId);
    } catch (error) {
      console.error('Failed to initialize chat:', error);
    }
  });
</script>

<Chat loadContact={targetContactId} />