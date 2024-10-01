<script lang="ts">
  import Post from '$lib/components/post.svelte'
  import { Card } from 'flowbite-svelte'

  const prop = {
    text: 'Here are the biggest enterprise technology acquisitions of 2021 so far, in reverse chronological order.',
    replayCount: 2,
    date: new Date('2024-09-24 12:45:23Z'),
    parent: true,
  }

  const childs = [
    { text: '123', date: new Date('2024-09-24 12:45:23Z') },
    { text: 'Test', date: new Date('2024-09-24 12:45:23Z') },
  ]
</script>

<Card size="xl" class="m-11">
  <Post {...prop}>
    {#each childs as post}
      <Post {...post} Class="mt-6" />
    {/each}
  </Post>
</Card>
