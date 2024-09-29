<script lang="ts">
	import { Card, Dropdown, DropdownItem, Textarea, ToolbarButton } from 'flowbite-svelte'
	import { AngleDownOutline, DotsHorizontalOutline, PlusOutline, AngleUpOutline, PaperPlaneOutline } from 'flowbite-svelte-icons'
	import placeholder from '$lib/assets/angy.png'

	export let text = 'Here are the biggest enterprise technology acquisitions of 2021 so far, in reverse chronological order.'
	export let replayCount = 2

	export let showReplies = false
</script>

<Card size="xl">
	<footer class="flex items-center gap-2">
		<img class="w-7 h-7 rounded-full" src={placeholder} alt="Bonnie Green avatar" />
		<span class="font-semibold text-sm text-black dark:text-white"> Bonnie Green </span>
		<span class="text-sm ml-1">
			<time datetime="Tue Sep 24 2024" title="Tue Sep 24 2024">14 days ago</time>
		</span>

		<span class="ml-auto">
			<DotsHorizontalOutline class="dots-menu dark:text-white" />
		<Dropdown triggeredBy=".dots-menu" placement="right-start">
			<DropdownItem>Edit</DropdownItem>
			<DropdownItem>Remove</DropdownItem>
			<DropdownItem slot="footer">Report</DropdownItem>
		</Dropdown>
		</span>
	</footer>
	<div class="mt-3">
		<p class="font-normal text-black dark:text-gray-400 leading-tight">{text}</p>
	</div>

	<button on:click="{() => showReplies = !showReplies}"
		 class="mt-3 inline-flex items-center font-medium text-sm text-primary-600 dark:text-primary-500 hover:underline mb-2">
		{#if replayCount > 0}
			{replayCount} {replayCount === 1 ? 'reply' : 'replies'}
			{#if !showReplies}
				<AngleDownOutline size="sm" class="ml-1 mt-auto" />
			{:else}
				<AngleUpOutline size="sm" class="ml-1 mt-auto" />
			{/if}
		{:else}
			reply
			<PlusOutline size="sm" class="ml-1 mt-auto" />
		{/if}
	</button>

	{#if showReplies}
		<form>
			<label for="chat" class="sr-only">Your message</label>
			<div class="flex items-center py-2 rounded-lg dark:bg-gray-700">
				<Textarea id="chat" class="bg-white dark:bg-gray-800" rows="1" placeholder="Your message..." />
				<ToolbarButton type="submit" color="blue" class="rounded-full text-primary-600 dark:text-primary-500 ml-4">
					<PaperPlaneOutline class="w-6 h-6 rotate-45" />
					<span class="sr-only">Send message</span>
				</ToolbarButton>
			</div>
		</form>
	{/if}
</Card>
