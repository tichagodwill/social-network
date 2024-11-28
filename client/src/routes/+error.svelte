<script>
    import { page } from '$app/stores';
</script>

<div class="flex flex-col items-center justify-center min-h-[60vh]">
    <h1 class="text-4xl font-bold mb-4">{$page.status}</h1>
    <p class="text-xl text-gray-600">{$page.error?.message}</p>
</div> 