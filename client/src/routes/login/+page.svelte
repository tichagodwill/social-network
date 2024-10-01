<script>
	import LoginForm from '$lib/components/login-form.svelte'
	import RegisterForm from '$lib/components/register-form.svelte'

	let register = false
</script>

<div class="w-1/3 m-auto pt-10">
	{#if !register}
		<LoginForm RequestRegister="{() => {register = true}}"></LoginForm>
	{:else}
		<RegisterForm Back="{() => {register = false}}"></RegisterForm>
	{/if}
</div>