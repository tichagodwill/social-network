<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { Avatar, Input, Button } from 'flowbite-svelte';

  interface User {
    id: number;
    username: string;
    avatar: string | null;
    is_private: boolean;
  }

  let users: User[] = [];
  let searchQuery = '';
  let isLoading = false;

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
        users = await response.json();
      }
    } catch (error) {
      console.error('Failed to fetch users:', error);
    } finally {
      isLoading = false;
    }
  }

  const debouncedSearch = debounce((searchValue: string) => {
    users = [];
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

<div class="max-w-4xl mx-auto px-4 py-8">
  <div class="mb-8">
    <Input
      type="search"
      placeholder="Search users..."
      value={searchQuery}
      on:input={handleSearch}
      class="w-full"
    >
      <svg slot="left" class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </Input>
  </div>

  <div class="bg-white rounded-lg shadow">
    {#each users as user (user.id)}
      <div class="border-b border-gray-200 last:border-0">
        <div class="flex items-center p-4 hover:bg-gray-50 transition-colors">
          <div class="flex-shrink-0">
            <Avatar
              src={user.avatar || generateAvatar(user.username)}
              size="md"
              alt={`${user.username}'s avatar`}
              class="ring-2 ring-blue-500 ring-offset-2"
            />
          </div>
          <div class="ml-4 flex-grow">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <h3 class="text-lg font-semibold text-gray-900">{user.username}</h3>
                <div class="relative group">
                  {#if user.is_private}
                    <!-- Closed Eye Icon for Private -->
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  {:else}
                    <!-- Open Eye Icon for Public -->
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.522 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.478 0-8.268-2.943-9.542-7z" />
                    </svg>
                  {/if}
                  <!-- Tooltip -->
                  <div class="absolute left-1/2 -translate-x-1/2 bottom-full mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                    {user.is_private ? 'Private Account' : 'Public Account'}
                  </div>
                </div>
              </div>
              <Button
                color="light"
                size="sm"
                class="ml-4"
                on:click={() => goToProfile(user.id)}
              >
                View Profile
                <svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </Button>
            </div>
          </div>
        </div>
      </div>
    {/each}

    {#if isLoading}
      <div class="flex justify-center items-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      </div>
    {/if}
    {users}
    {#if (users == null || users.length === 0) && !isLoading}
      <div class="text-center py-12 px-4">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">No users found</h3>
        {#if searchQuery}
          <p class="mt-1 text-sm text-gray-500">Try adjusting your search terms.</p>
        {:else}
          <p class="mt-1 text-sm text-gray-500">No users are available at the moment.</p>
        {/if}
      </div>
    {/if}
  </div>
</div>