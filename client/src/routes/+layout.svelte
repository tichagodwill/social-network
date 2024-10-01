<script lang="ts">
  import '../app.css'
  import { Footer, FooterCopyright, Navbar, NavBrand, NavLi, NavUl, NavHamburger } from 'flowbite-svelte'

  import logo from '$lib/assets/angy.png'
  import type { LayoutData } from './$types'

  export let data: LayoutData
</script>

  <Navbar>
    <NavBrand href="/">
      <img src={logo} class="h-6 sm:h-9" alt="😾" />
      <span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white ml-3">Stalky</span>
    </NavBrand>
    <NavHamburger />
    <NavUl>
      <NavLi href="/">Home</NavLi>
      <NavLi href="/about">Topics</NavLi>
      <NavLi href="/chat">Chat</NavLi>
      <NavLi href="/more">More</NavLi>
      {#if data.user.loggedIn}
        <NavLi href="/profile">Profile</NavLi>
      {:else}
        <NavLi href="/login">Log In</NavLi>
      {/if}
    </NavUl>
  </Navbar>

  <div class="flex flex-col h-[calc(100vh-9rem)] justify-between m-9">
    <slot></slot>

    <Footer>
      <div class="sm:flex sm:items-center sm:justify-center my-6">
        <FooterCopyright href="/" by="Stalky Team™" year={2024} />
      </div>
    </Footer>
  </div>
