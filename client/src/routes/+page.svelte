<script lang="ts">
	import Post from '$lib/components/post.svelte';

	const prop = {
		text: 'Here are the biggest enterprise technology acquisitions of 2021 so far, in reverse chronological order.',
		replayCount: 2,
		date: new Date("2024-09-24 12:45:23Z"),
	}
</script>
<div class="m-5">
	<Post {...prop} />
</div>