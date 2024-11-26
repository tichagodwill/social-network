<script lang="ts">
	import { Input, Label, Button, Alert } from 'flowbite-svelte';
	import { auth } from '$lib/stores/auth';

	let email = '';
	let password = '';
	let error = '';
	let loading = false;

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		error = '';
		loading = true;

		try {
			await auth.login(email, password);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
			console.error('Login error:', err);
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
	<div class="w-full max-w-md space-y-8">
		<div>
			<h2 class="mt-6 text-center text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
				Sign in to your account
			</h2>
		</div>

		<form class="mt-8 space-y-6" on:submit={handleSubmit}>
			{#if error}
				<Alert color="red" class="mb-4">
					{error}
				</Alert>
			{/if}

			<div class="space-y-4 rounded-md shadow-sm">
				<div>
					<Label for="email">Email address</Label>
					<Input
						id="email"
						name="email"
						type="email"
						required
						bind:value={email}
						disabled={loading}
					/>
				</div>

				<div>
					<Label for="password">Password</Label>
					<Input
						id="password"
						name="password"
						type="password"
						required
						bind:value={password}
						disabled={loading}
					/>
				</div>
			</div>

			<div>
				<Button type="submit" class="w-full" disabled={loading}>
					{loading ? 'Signing in...' : 'Sign in'}
				</Button>
			</div>

			<div class="text-center">
				<a href="/register" class="text-primary-600 hover:underline">
					Don't have an account? Register
				</a>
			</div>
		</form>
	</div>
</div>