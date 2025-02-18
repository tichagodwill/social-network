<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { Avatar, Input, Button } from 'flowbite-svelte';
  import { fade, fly, scale, slide, crossfade } from 'svelte/transition';
  import { elasticOut, quintOut, cubicInOut } from 'svelte/easing';
  import { flip } from 'svelte/animate';

  interface User {
    id: number;
    username: string;
    avatar: string | null;
    is_private: boolean;
  }

  let users: User[] = [];
  let searchQuery = '';
  let isLoading = false;
  let previousUsers: User[] = [];
  let tooltipContent = '';
  let tooltipVisible = false;
  let tooltipX = 0;
  let tooltipY = 0;

  function showTooltip(event: MouseEvent, content: string) {
    const target = event.currentTarget as HTMLElement;
    const rect = target.getBoundingClientRect();
    tooltipContent = content;
    tooltipX = rect.left + (rect.width / 2);
    tooltipY = rect.top - 10;
    tooltipVisible = true;
  }

  function hideTooltip() {
    tooltipVisible = false;
  }

  const [send, receive] = crossfade({
    duration: 400,
    fallback(node, params) {
      return {
        duration: 400,
        easing: cubicInOut,
        css: t => `
          opacity: ${t};
          transform: scale(${t});
        `
      };
    }
  });

  function generateAvatar(username: string): string {
    const firstLetter = username ? username.charAt(0).toUpperCase() : 'U';
    return `https://ui-avatars.com/api/?name=${firstLetter}&background=0ea5e9&color=fff&size=128`;
  }

  function debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number
  ): (...args: Parameters<T>) => void {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;

    return (...args: Parameters<T>) => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }

      timeoutId = setTimeout(() => {
        func(...args);
        timeoutId = null;
      }, wait);
    };
  }

  async function fetchUsers(search: string = '') {
    if (isLoading) return;

    isLoading = true;
    previousUsers = [...users];

    try {
      const response = await fetch(`http://localhost:8080/explore`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({ search })
      });

      if (response.ok) {
        const newUsers = await response.json();
        users = newUsers;
      }
    } catch (error) {
      console.error('Failed to fetch users:', error);
    } finally {
      isLoading = false;
    }
  }

  const debouncedSearch = debounce((searchValue: string) => {
    fetchUsers(searchValue);
  }, 300);

  function handleSearch(event: Event) {
    const target = event.target as HTMLInputElement;
    searchQuery = target.value;
    debouncedSearch(searchQuery);
  }

  function goToProfile(userId: number) {
    goto(`/profile/${userId}`);
  }

  onMount(() => {
    fetchUsers();
  });
</script>

<style>
  @keyframes fadeSlideIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  :global(.bg-white) {
    background-color: white;
  }
</style>

<div class="max-w-4xl mx-auto px-4 py-8" in:fade={{ duration: 300 }}>
  <div class="sticky top-0 z-10 bg-white/80 backdrop-blur-lg rounded-lg shadow-lg mb-6 p-4 transform transition-all duration-300"
       in:fly={{ y: -20, duration: 400, delay: 150 }}>
    <div class="relative">
      <Input
        type="search"
        placeholder="Search users..."
        value={searchQuery}
        on:input={handleSearch}
        class="w-full !bg-gray-50/50 !border-gray-200 focus:!border-blue-400 !ring-blue-400/30 transition-all duration-300"
      >
        <div slot="left" class="flex items-center">
          <svg class="w-5 h-5 text-gray-500 transition-colors duration-300 group-focus-within:text-blue-500"
               fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </Input>
      {#if isLoading}
        <div class="absolute right-3 top-1/2 -translate-y-1/2"
             in:fade={{ duration: 200 }}
             out:fade={{ duration: 150 }}>
          <div class="animate-spin rounded-full h-5 w-5 border-2 border-blue-500 border-t-transparent"></div>
        </div>
      {/if}
    </div>
  </div>

  <div class="bg-white rounded-lg shadow overflow-hidden transition-all duration-300"
       in:fade={{ duration: 300, delay: 200 }}>
    {#if users.length > 0}
      <div class="divide-y divide-gray-200">
        {#each users as user, i (user.id)}
          <div
            class="bg-white"
            in:receive|local={{ key: user.id }}
            out:send|local={{ key: user.id }}
            animate:flip={{ duration: 300 }}
          >
            <div class="overflow-hidden">
              <div
                class="flex items-center p-4 transition-all duration-300 ease-out cursor-pointer bg-white
                hover:bg-gradient-to-r hover:from-gray-50 hover:to-blue-50/30
                hover:shadow-lg hover:shadow-blue-100/50
                relative group"
                style="animation: fadeSlideIn {300 + i * 50}ms forwards {i * 50}ms"
                on:mouseenter={(e) => {
                  e.currentTarget.style.transform = 'scale(1.01)';
                  e.currentTarget.style.backgroundColor = 'rgb(249, 250, 251)';
                }}
                on:mouseleave={(e) => {
                  e.currentTarget.style.transform = 'scale(1)';
                  e.currentTarget.style.backgroundColor = 'white';
                }}
                on:click={() => goToProfile(user.id)}
                on:keydown={(e) => e.key === 'Enter' && goToProfile(user.id)}
              >
                <div class="flex-shrink-0 transform transition-transform duration-300 group-hover:scale-105">
                  <Avatar
                    src={user.avatar || generateAvatar(user.username)}
                    size="md"
                    alt={`${user.username}'s avatar`}
                    class="ring-2 ring-blue-500 ring-offset-2 transition-all duration-300 group-hover:ring-4 group-hover:ring-blue-400"
                  />
                </div>

                <div class="ml-4 flex-grow min-w-0">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2 min-w-0">
                      <h3 class="text-lg font-semibold text-gray-900 transition-colors duration-300 group-hover:text-blue-600 truncate">{user.username}</h3>
                      <div class="relative inline-block">
                        <div
                          on:mouseenter={(e) => showTooltip(e, user.is_private ? 'Private Profile' : 'Public Profile')}
                          on:mouseleave={hideTooltip}
                        >
                          {#if user.is_private}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-500 transition-colors duration-300 hover:text-blue-500 cursor-help" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                            </svg>
                          {:else}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-500 transition-colors duration-300 hover:text-blue-500 cursor-help" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.522 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.478 0-8.268-2.943-9.542-7z" />
                            </svg>
                          {/if}
                        </div>
                      </div>
                    </div>
                    <Button
                      color="light"
                      size="sm"
                      class="ml-4 transition-all duration-300 ease-out flex-shrink-0
                      hover:scale-105 hover:shadow-md hover:shadow-blue-100/50
                      hover:bg-gradient-to-r hover:from-blue-50 hover:to-blue-100/50
                      group-hover:translate-x-1"
                    >
                      <span class="transition-all duration-300 group-hover:translate-x-0.5">View Profile</span>
                      <svg class="w-4 h-4 ml-2 transition-all duration-300 group-hover:translate-x-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}

    {#if (users == null || users.length === 0) && !isLoading}
      <div class="text-center py-12 px-4"
           in:scale={{ duration: 400, delay: 100, easing: elasticOut }}
           out:fade={{ duration: 200 }}>
        <div class="bg-gray-50/50 rounded-lg p-8 transform transition-all duration-300 hover:scale-[1.01] hover:bg-gray-50">
          <svg class="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.478 0-8.268-2.943-9.542-7z" />
          </svg>
          <h3 class="text-lg font-semibold text-gray-900 mb-2">No users found</h3>
          {#if searchQuery}
            <p class="text-gray-600">Try adjusting your search terms or try a different search.</p>
          {:else}
            <p class="text-gray-600">No users are available at the moment.</p>
          {/if}
        </div>
      </div>
    {/if}
  </div>
  {#if tooltipVisible}
    <div
      class="fixed bg-gray-900 text-white text-xs rounded-lg px-3 py-2 pointer-events-none shadow-lg transition-opacity duration-200"
      style="left: {tooltipX}px; top: {tooltipY}px; transform: translate(-50%, -100%); z-index: 99999;"
      transition:fade={{ duration: 200 }}
    >
      {tooltipContent}
      <div class="absolute left-1/2 -translate-x-1/2 top-full border-[6px] border-transparent border-t-gray-900"></div>
    </div>
  {/if}
</div>