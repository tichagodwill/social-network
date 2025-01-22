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

<div class="min-h-screen bg-gradient-to-b from-gray-50 to-white dark:from-gray-900 dark:to-gray-800 flex items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
	<div class="w-full max-w-md">
		<div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 space-y-8">
			<!-- Logo/Brand -->
			<div class="text-center">
				<div class="flex justify-center mb-4">
					<svg class="w-12 h-12 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
							  d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
					</svg>
				</div>
				<h2 class="text-3xl font-bold text-gray-900 dark:text-white">
					Welcome back
				</h2>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					Sign in to your account to continue
				</p>
			</div>

			<form class="space-y-6" on:submit={handleSubmit}>
				{#if error}
					<Alert color="red" class="animate-fade-in">
						<svg slot="icon" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
								  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						{error}
					</Alert>
				{/if}

				<div class="space-y-4">
					<div class="space-y-2">
						<Label for="email" class="text-sm font-medium text-gray-700 dark:text-gray-300">
							Email address
						</Label>
						<Input
								id="email"
								name="email"
								type="email"
								required
								bind:value={email}
								disabled={loading}
								class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
								placeholder="Enter your email"
						/>
					</div>

					<div class="space-y-2">
						<Label for="password" class="text-sm font-medium text-gray-700 dark:text-gray-300">
							Password
						</Label>
						<Input
								id="password"
								name="password"
								type="password"
								required
								bind:value={password}
								disabled={loading}
								class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
								placeholder="Enter your password"
						/>
					</div>
				</div>

				<div class="flex items-center justify-between text-sm">
					<div class="flex items-center">
						<input
								id="remember-me"
								name="remember-me"
								type="checkbox"
								class="h-4 w-4 text-primary-500 focus:ring-primary-500 border-gray-300 rounded"
						/>
						<label for="remember-me" class="ml-2 text-gray-600 dark:text-gray-400">
							Remember me
						</label>
					</div>
					<a href="/forgot-password" class="text-primary-500 hover:text-primary-600 font-medium">
						Forgot password?
					</a>
				</div>

				<Button
						type="submit"
						class="w-full py-3 px-4 bg-primary-500 hover:bg-primary-600 text-white font-medium rounded-lg
                           transition-colors duration-200 flex items-center justify-center space-x-2"
						disabled={loading}
				>
					{#if loading}
						<svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
						</svg>
						<span>Signing in...</span>
					{:else}
						<span>Sign in</span>
					{/if}
				</Button>
			</form>

			<div class="text-center space-y-4">
				<div class="relative">
					<div class="absolute inset-0 flex items-center">
						<div class="w-full border-t border-gray-200 dark:border-gray-700"></div>
					</div>
					<div class="relative flex justify-center">
                        <span class="px-4 bg-white dark:bg-gray-800 text-sm text-gray-500">
                            or
                        </span>
					</div>
				</div>

				<p class="text-sm text-gray-600 dark:text-gray-400">
					Don't have an account?
					<a href="/register" class="font-medium text-primary-500 hover:text-primary-600 hover:underline">
						Create one now
					</a>
				</p>
			</div>
		</div>
	</div>
</div>