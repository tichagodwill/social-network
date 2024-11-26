<script lang="ts">
	import { Input, Label, Button, Checkbox } from 'flowbite-svelte'
	import { ArrowLeftOutline } from 'flowbite-svelte-icons'
	import { auth } from '$lib/stores/auth'

	const requiredMark = `<span class="text-red-500">*</span>`

	let formData = {
		email: '',
		password: '',
		confirmPassword: '',
		first_name: '',
		last_name: '',
		date_of_birth: '',
		avatar: 'default-avatar.png',
		username: '',
		about_me: ''
	}

	let error = ''

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault()
		error = ''

		if (formData.password !== formData.confirmPassword) {
			error = 'Passwords do not match'
			return
		}

		try {
			const { confirmPassword, ...registrationData } = formData
			console.log('Submitting registration data:', registrationData)
			await auth.register(registrationData)
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed'
		}
	}

	export let Back = () => {}
</script>

<Button class="pt-1 pb-1 pl-2 pr-2 mb-3" color="alternative" on:click={Back}>
	<ArrowLeftOutline /> Back
</Button>

{#if error}
	<div class="p-4 mb-4 text-red-800 bg-red-100 rounded-lg dark:text-red-400 dark:bg-red-900">
		{error}
	</div>
{/if}

<form on:submit={handleSubmit} class="space-y-6">
	<div class="grid gap-6 mb-6 md:grid-cols-2">
		<div>
			<Label for="first_name" class="mb-2">First name{@html requiredMark}</Label>
			<Input type="text" id="first_name" bind:value={formData.first_name} required />
		</div>
		<div>
			<Label for="last_name" class="mb-2">Last name{@html requiredMark}</Label>
			<Input type="text" id="last_name" bind:value={formData.last_name} required />
		</div>
		<div>
			<Label for="username" class="mb-2">Username{@html requiredMark}</Label>
			<Input type="text" id="username" bind:value={formData.username} required />
		</div>
		<div>
			<Label for="date_of_birth" class="mb-2">Date of Birth{@html requiredMark}</Label>
			<Input type="date" id="date_of_birth" bind:value={formData.date_of_birth} required />
		</div>
	</div>

	<div class="mb-6">
		<Label for="about_me" class="mb-2">About Me{@html requiredMark}</Label>
		<Input type="text" id="about_me" bind:value={formData.about_me} required />
	</div>

	<div class="mb-6">
		<Label for="email" class="mb-2">Email address{@html requiredMark}</Label>
		<Input type="email" id="email" bind:value={formData.email} required />
	</div>

	<div class="mb-6">
		<Label for="password" class="mb-2">Password{@html requiredMark}</Label>
		<Input type="password" id="password" bind:value={formData.password} required />
	</div>

	<div class="mb-6">
		<Label for="confirm_password" class="mb-2">Confirm password{@html requiredMark}</Label>
		<Input type="password" id="confirm_password" bind:value={formData.confirmPassword} required />
	</div>

	<Checkbox class="mb-6" required>
		I agree to the terms and conditions
	</Checkbox>

	<Button type="submit" class="w-full">Register</Button>
</form>
