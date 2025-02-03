<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { get } from 'svelte/store';
  import { fade, fly, slide } from 'svelte/transition';
  import type { Post, Comment } from '$lib/types';

  let post: Post | null = null;
  let comments: Comment[] = [];
  let loading = true;
  let error: string | null = null;
  let newComment = '';
  let submitting = false;

  function generateAvatar(username: string): string {
    const firstLetter = username ? username.charAt(0).toUpperCase() : 'U';
    return `https://ui-avatars.com/api/?name=${firstLetter}&background=0ea5e9&color=fff&size=128`;
  }

  const getAvatarSrc = (comment: Comment): string => {
    return comment.avatar || generateAvatar(comment.author);
  };

  const submitComment = async () => {
    if (!newComment.trim()) return;

    const postId = get(page).params.id;
    submitting = true;

    try {
      const response = await fetch(`http://localhost:8080/posts/addComment`, {
        credentials: 'include',
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          content: newComment,
          postId: postId
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to submit comment');
      }

      const newCommentData = await response.json();
      comments = [...comments, newCommentData];
      newComment = '';
    } catch (err: any) {
      error = err.message;
    } finally {
      submitting = false;
    }
  };

  onMount(async () => {
    const postId = get(page).params.id;
    try {
      const response = await fetch(`http://localhost:8080/posts/${postId}/details`, {
        credentials: 'include'
      });
      if (!response.ok) {
        throw new Error('Failed to fetch data');
      }

      const data = await response.json();
      post = data.post;
      if(data.comments){
        comments = data.comments;
      }
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <div class="flex items-center justify-center min-h-screen" in:fade>
    <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500" />
  </div>
{:else if error}
  <div class="min-h-screen flex items-center justify-center bg-gray-50"
       in:fly="{{ y: 20, duration: 500 }}">
    <div class="text-center p-8 max-w-lg w-full bg-white rounded-lg shadow-lg">
      <div class="mb-6">
        <svg class="mx-auto h-16 w-16 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      </div>
      <h2 class="text-2xl font-bold text-gray-900 mb-3">Oops! Something went wrong</h2>
      <p class="text-gray-600 mb-6">{error}</p>
      <button
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transform hover:scale-105 transition-all"
        on:click={() => window.location.reload()}
      >
        <svg class="w-4 h-4 mr-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Try Again
      </button>
    </div>
  </div>
{:else if post}
  <div class="container mx-auto px-4 py-8 max-w-4xl" in:fly="{{ y: 20, duration: 500 }}">
    <div class="bg-white rounded-lg shadow-md mb-8 hover:shadow-lg transition-shadow duration-300">
      <div class="p-6">
        <h1 class="text-3xl font-bold mb-4 text-gray-900">{post.title}</h1>
        <div class="flex items-center space-x-4 text-gray-600 mb-4">
          <div class="flex items-center hover:text-blue-600 transition-colors">
            <svg class="w-4 h-4 mr-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
            <span>{post.authorName}</span>
          </div>
          <div class="flex items-center">
            <svg class="w-4 h-4 mr-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
              <line x1="16" y1="2" x2="16" y2="6" />
              <line x1="8" y1="2" x2="8" y2="6" />
              <line x1="3" y1="10" x2="21" y2="10" />
            </svg>
            <span>{new Date(post.created_at).toLocaleDateString()}</span>
          </div>
        </div>
        <p class="text-lg leading-relaxed whitespace-pre-wrap text-gray-700">{post.content}</p>
      </div>
    </div>

    <div class="mt-8" transition:slide>
      <h2 class="text-2xl font-bold mb-4 flex items-center text-gray-900">
        <svg class="w-6 h-6 mr-2 text-blue-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
        </svg>
        Comments ({comments.length})
      </h2>

      {#if !comments}
        <p class="text-gray-500 italic" in:fade>Loading comments...</p>
      {:else if comments.length === 0}
        <p class="text-gray-500 italic" in:fade>No comments yet.</p>
      {:else}
        <div class="space-y-4">
          {#each comments as comment, i}
            <div class="bg-gray-50 rounded-lg shadow p-4 hover:shadow-md transition-shadow duration-300"
                 in:fly="{{ y: 20, duration: 300, delay: i * 100 }}">
              <div class="flex items-start space-x-4">
                <img
                  src={getAvatarSrc(comment)}
                  alt={comment.author}
                  class="w-10 h-10 rounded-full ring-2 ring-blue-500 ring-opacity-50"
                />
                <div class="flex-1">
                  <div class="flex items-center justify-between mb-2">
                    <span class="font-semibold text-gray-900">{comment.author_name}</span>
                    <span class="text-sm text-gray-500">
                      {new Date(comment.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <p class="text-gray-700">{comment.content}</p>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      <div class="mt-8" in:slide>
        <form
          on:submit|preventDefault={submitComment}
          class="bg-white rounded-lg shadow-sm p-4 hover:shadow-md transition-shadow duration-300"
        >
          <textarea
            bind:value={newComment}
            placeholder="Write a comment..."
            class="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none min-h-[100px] transition-all duration-300"
            disabled={submitting}
          />
          <div class="mt-3 flex justify-end">
            <button
              type="submit"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center transform hover:scale-105 transition-all duration-300"
              disabled={submitting || !newComment.trim()}
            >
              {#if submitting}
                <div class="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-white mr-2" />
              {/if}
              Post Comment
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{:else}
  <div class="min-h-screen flex items-center justify-center bg-gray-50" in:fade>
    <div class="text-center p-8">
      <h2 class="text-2xl font-bold text-gray-900 mb-3">Post Not Found</h2>
      <p class="text-gray-600">The requested post could not be found.</p>
    </div>
  </div>
{/if}