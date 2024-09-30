<script lang="ts">
	import { Card, Dropdown, DropdownItem, Textarea, ToolbarButton } from 'flowbite-svelte'
	import { AngleDownOutline, CloseOutline, DotsHorizontalOutline, PlusOutline, AngleUpOutline, PaperPlaneOutline, FaceGrinOutline, ImageOutline } from 'flowbite-svelte-icons'
	import placeholder from '$lib/assets/angy.png'
	import { getFormattedDate } from '$lib/dateFormater'

	export let text: string
	export let replayCount = 2
	export let date: Date
	export let showReplies = false
	export let enableReplay = false
	// export let parent = false

	const postDate = getFormattedDate(date)
</script>

<Card size="xl">
	<footer class="flex items-center gap-2">
		<img class="w-7 h-7 rounded-full" src={placeholder} alt="Bonnie Green avatar" />
		<span class="font-semibold text-sm text-black dark:text-white"> Bonnie Green </span>
		<span class="text-sm ml-1">
			<time datetime="{postDate.formated}" title="{postDate.formated}">{postDate.diff}</time>
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
			{replayCount}
			{#if !showReplies}
				show {replayCount === 1 ? 'reply' : 'replies'} <AngleDownOutline size="sm" class="ml-1 mt-auto" />
			{:else}
				hide {replayCount === 1 ? 'reply' : 'replies'} <AngleUpOutline size="sm" class="ml-1 mt-auto" />
			{/if}
		{:else}
			reply
			<PlusOutline size="sm" class="ml-1 mt-auto" />
		{/if}
	</button>

	{#if enableReplay}
		<form>
			<label for="chat" class="sr-only">Replay</label>
			<div class="flex items-center px-3 py-2 rounded-lg bg-gray-50 dark:bg-gray-700">
					{#if enableReplay}
						<ToolbarButton color="dark" class="text-gray-500 dark:text-gray-400">
							<CloseOutline class="" />
							<span class="sr-only">Close</span>
						</ToolbarButton>
					{/if}
				<ToolbarButton color="dark" class="text-gray-500 dark:text-gray-400">
					<ImageOutline class="w-6 h-6" />
					<span class="sr-only">Upload image</span>
				</ToolbarButton>
				<ToolbarButton color="dark" class="text-gray-500 dark:text-gray-400">
					<FaceGrinOutline class="w-6 h-6" />
					<span class="sr-only">Add emoji</span>
				</ToolbarButton>
				<Textarea id="chat" class="bg-white dark:bg-gray-800 ml-1" rows="1" placeholder="Your replay...">
				</Textarea>
				<ToolbarButton type="submit" color="blue" class="rounded-full text-primary-600 dark:text-primary-500 ml-4">
					<PaperPlaneOutline class="w-6 h-6 rotate-90" />
				</ToolbarButton>
			</div>
		</form>
	{/if}
	<slot />
</Card>
