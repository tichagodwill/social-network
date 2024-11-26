/// <reference types="@sveltejs/kit" />
/// <reference types="svelte" />

// See https://kit.svelte.dev/docs/types#app
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user?: {
				id: number;
				username: string;
			};
		}
		interface PageData {
			user?: {
				id: number;
				username: string;
				loggedIn?: boolean;
			};
		}
		interface Platform {}
	}

	interface PageData {
		user?: {
			id: number;
			username: string;
			loggedIn?: boolean;
		};
	}
}

declare module '@emoji-mart/data';

export {};
