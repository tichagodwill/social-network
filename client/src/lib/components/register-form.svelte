<script>
	/* eslint-disable svelte/no-at-html-tags */
	import { Input, Label, Button, Checkbox, Radio } from 'flowbite-svelte'
	import { ArrowLeftOutline } from 'flowbite-svelte-icons'

	const requiredMark = `<span class="primary-text" />`
	let gender = "male";

	export let Back = () => { }
</script>

<Button class="pt-1 pb-1 pl-2 pr-2 mb-3" color="alternative" on:click={Back}>
	<ArrowLeftOutline  /> Back
</Button>

<form>
	<div class="grid gap-6 mb-6 md:grid-cols-2">
		<div>
			<Label for="first_name" class="mb-2">First name{@html requiredMark}</Label>
			<Input type="text" id="first_name" placeholder="John" required />
		</div>
		<div>
			<Label for="last_name" class="mb-2">Last name{@html requiredMark}</Label>
			<Input type="text" id="last_name" placeholder="Doe" required />
		</div>
		<div>
			<Label for="company" class="mb-2">Mood</Label>
			<Input type="text" id="mood" placeholder="delirious" />
		</div>
		<div>
			<Label for="phone" class="mb-2">Phone number</Label>
			<Input type="tel" id="phone" placeholder="33111234" pattern={"[0-9]{8}"} />
		</div>
		<div>
			<Label for="website" class="mb-2">Gender</Label>
			<div class="flex gap-4 h-10">
				<Radio bind:group={gender} color="purple" value="female">Female</Radio>
				<Radio bind:group={gender} color="teal" value="male">Male</Radio>
				<Radio bind:group={gender} color="yellow" value="male2">Other</Radio>
			</div>
		</div>
		<div>
			<Label for="visitors" class="mb-2">Age{@html requiredMark}</Label>
			<Input type="number" id="age" placeholder="21" required />
		</div>
	</div>
	<div class="mb-6">
		<Label for="email" class="mb-2">Email address{@html requiredMark}</Label>
		<Input type="email" id="email" placeholder="john.doe@company.com" required />
	</div>
	<div class="mb-6">
		<Label for="password" class="mb-2">Password{@html requiredMark}</Label>
		<Input type="password" id="password" placeholder="•••••••••" required />
	</div>
	<div class="mb-6">
		<Label for="confirm_password" class="mb-2">Confirm password{@html requiredMark}</Label>
		<Input type="password" id="confirm_password" placeholder="•••••••••" required />
	</div>
	<Checkbox class="mb-6 space-x-1 rtl:space-x-reverse hover:underline">
		I will give this team a <p class="primary-text">pass</p>.
	</Checkbox>
	<Button type="submit">Register</Button>
</form>
