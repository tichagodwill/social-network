<script lang="ts">
	import { Dropdown, DropdownItem, Textarea, ToolbarButton } from 'flowbite-svelte'
	import { AngleDownOutline, CloseOutline, DotsHorizontalOutline, AngleUpOutline, PaperPlaneSolid, FaceGrinOutline, ImageOutline } from 'flowbite-svelte-icons'
	import placeholder from '$lib/assets/angy.png'
	import { getFormattedDate } from '$lib/dateFormater'
	import { auth } from '$lib/stores/auth'
	import { onMount } from 'svelte'

	export let Class=''

	export let user = 'wacky cat'
	export let userImg = ''
	export let text = ''
	export let replayCount = 0
	export let date = new Date
	export let enableReplay = false
	export let parent = false

	export let showThread = false

	export let id: number
	export let comments: any[] = []
	let newComment = ''
	let loading = false

	export let groupId: number

	const postDate = getFormattedDate(date)

	if (userImg.length == 0)
		userImg = placeholder

	async function loadComments() {
		try {
			const response = await fetch(`http://localhost:8080/groups/${groupId}/posts/${id}/comments`, {
				credentials: 'include'
			})
			if (response.ok) {
				const data = await response.json()
				comments = data
				console.log('Loaded comments:', comments)
			}
		} catch (error) {
			console.error('Failed to load comments:', error)
		}
	}

	async function handleSubmitComment() {
		if (!newComment.trim()) return
		
		const commentData = {
			content: newComment
		}
		
		try {
			loading = true
			const response = await fetch(`http://localhost:8080/groups/${groupId}/posts/${id}/comments`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify(commentData)
			})

			const data = await response.json()
			console.log('Comment response:', data)

			if (!response.ok) {
				throw new Error(data.error || 'Failed to create comment')
			}

			newComment = ''
			await loadComments()
		} catch (error) {
			console.error('Failed to post comment:', error)
		} finally {
			loading = false
		}
	}

	onMount(() => {
		loadComments()
	})
</script>

<div class='{Class}'>
	<footer class="flex items-center gap-2">
		<img class="w-7 h-7 rounded-full" src={userImg} alt="{user} avatar" />
		<span class="font-semibold text-sm text-black dark:text-white">{user}</span>
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

	<div>
		<button on:click="{() => {}}" class="text-sm hover:underline">show</button>
		<button on:click="{() => {enableReplay = !enableReplay}}" class="text-sm mr-2 hover:underline">replay</button>

		{#if parent && replayCount > 0}
				<button on:click={() => showThread = !showThread} class="mt-3 inline-flex items-center font-medium text-sm text-primary-600 dark:text-primary-500 hover:underline mb-2">
					show {replayCount == 1 ? 'reply' : 'replies'}
					{#if !showThread}
						<AngleDownOutline size="sm" class="ml-1 mt-auto" />
					{:else}
						<AngleUpOutline size="sm" class="ml-1 mt-auto" />
					{/if}
				</button>
		{/if}
	</div>

	<div class="pl-8">
		{#if enableReplay}
			<form on:submit|preventDefault={handleSubmitComment}>
				<label for="chat" class="sr-only">Reply</label>
				<div class="flex items-center px-3 py-2 rounded-lg bg-gray-50 dark:bg-gray-700">
					<ToolbarButton color="dark" class="text-gray-500 dark:text-gray-400" on:click={() => {enableReplay = false}}>
						<CloseOutline class="" />
						<span class="sr-only">Close</span>
					</ToolbarButton>
					<ToolbarButton color="dark" class="text-gray-500 dark:text-gray-400">
						<ImageOutline class="w-6 h-6" />
						<span class="sr-only">Upload image</span>
					</ToolbarButton>
					<ToolbarButton color="dark" class="text-gray-500 dark:text-gray-400">
						<FaceGrinOutline class="w-6 h-6" />
						<span class="sr-only">Add emoji</span>
					</ToolbarButton>
					<Textarea 
						id="chat" 
						class="bg-white dark:bg-gray-800 ml-1" 
						rows="1" 
						placeholder="Your reply..."
						bind:value={newComment}
						disabled={loading}
					/>
					<ToolbarButton 
						type="submit" 
						color="blue" 
						class="rounded-full text-primary-600 dark:text-primary-500 ml-4"
						disabled={loading}
					>
						<PaperPlaneSolid class="w-6 h-6 rotate-90" />
					</ToolbarButton>
				</div>
			</form>
		{/if}

		{#if showThread}
			{#each comments as comment (comment.id)}
				<div class="mt-4">
					<div class="flex items-start gap-2">
						<img 
							class="w-6 h-6 rounded-full" 
							src={comment.avatar || placeholder} 
							alt="{comment.author} avatar" 
						/>
						<div class="flex-1">
							<p class="font-semibold text-sm">{comment.author}</p>
							<p class="text-sm">{comment.content}</p>
							<p class="text-xs text-gray-500 mt-1">
								{getFormattedDate(new Date(comment.created_at))}
							</p>
						</div>
					</div>
				</div>
			{/each}
		{/if}
	</div>
</div>
