<script>
	import { Input, Label, Button } from 'flowbite-svelte'

	export let RequestRegister = () => {}
</script>

<form class="pt-10">
	<div class="mb-6">
		<Label for="email" class="mb-2">Email address</Label>
		<Input type="email" id="email" placeholder="john.doe@company.com" required />
	</div>
	<div class="mb-6">
		<Label for="password" class="mb-2">Password</Label>
		<Input type="password" id="password" placeholder="•••••••••" required />
	</div>
	<div class="flex justify-center items-center space-x-2">
		<Button type="submit">Login</Button>
		<Button color="alternative" on:click={RequestRegister}>Register</Button>
	</div>
</form>
