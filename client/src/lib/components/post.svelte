<script lang="ts">
	import { Card, Dropdown, DropdownItem } from 'flowbite-svelte'
	import { AngleDownOutline, DotsHorizontalOutline, PlusOutline } from 'flowbite-svelte-icons'
	import placeholder from '$lib/assets/angy.png'

	export let text = 'Here are the biggest enterprise technology acquisitions of 2021 so far, in reverse chronological order.'
	export let replayCount = 2

</script>

<Card size="lg">
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

	<a href="/"
		 class="mt-3 inline-flex items-center font-medium text-sm text-primary-600 dark:text-primary-500 hover:underline">
		{#if replayCount > 0}
			{replayCount} {replayCount === 1 ? 'reply' : 'replies'}
			<AngleDownOutline size="sm" class="ml-1 mt-auto" />
		{:else}
			reply
			<PlusOutline size="sm" class="ml-1 mt-auto" />
		{/if}
	</a>
</Card>