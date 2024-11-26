<script lang="ts">
    import RegisterForm from '$lib/components/register-form.svelte';
    import { goto } from '$app/navigation';

    function handleBack() {
        goto('/');
    }
</script>

<div class="container mx-auto px-4 py-8 max-w-2xl">
    <RegisterForm Back={handleBack} />
</div> 