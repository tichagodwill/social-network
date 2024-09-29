<script>
	import { Input, Label, Button, Checkbox } from 'flowbite-svelte'
	import { ArrowLeftOutline } from 'flowbite-svelte-icons'

	export let Back = () => { }
</script>

<Button class="pt-1 pb-1 pl-2 pr-2 mb-3" color="alternative" on:click={Back}>
	<ArrowLeftOutline  /> Back
</Button>

<form>
	<div class="grid gap-6 mb-6 md:grid-cols-2">
		<div>
			<Label for="first_name" class="mb-2">First name</Label>
			<Input type="text" id="first_name" placeholder="John" required />
		</div>
		<div>
			<Label for="last_name" class="mb-2">Last name</Label>
			<Input type="text" id="last_name" placeholder="Doe" required />
		</div>
		<div>
			<Label for="company" class="mb-2">Mood</Label>
			<Input type="text" id="mood" placeholder="delirious" required />
		</div>
		<div>
			<Label for="phone" class="mb-2">Phone number</Label>
			<Input type="tel" id="phone" placeholder="123-45-678" pattern={"[0-9]{3}-[0-9]{2}-[0-9]{3}"} required />
		</div>
		<div>
			<Label for="website" class="mb-2">Stuff</Label>
			<Input type="url" id="stuff" placeholder="more stuff" required />
		</div>
		<div>
			<Label for="visitors" class="mb-2">Age</Label>
			<Input type="number" id="age" placeholder="21" required />
		</div>
	</div>
	<div class="mb-6">
		<Label for="email" class="mb-2">Email address</Label>
		<Input type="email" id="email" placeholder="john.doe@company.com" required />
	</div>
	<div class="mb-6">
		<Label for="password" class="mb-2">Password</Label>
		<Input type="password" id="password" placeholder="•••••••••" required />
	</div>
	<div class="mb-6">
		<Label for="confirm_password" class="mb-2">Confirm password</Label>
		<Input type="password" id="confirm_password" placeholder="•••••••••" required />
	</div>
	<Checkbox class="mb-6 space-x-1 rtl:space-x-reverse" required>
		I will give this team a <p class="text-primary-700 dark:text-primary-600 hover:overline">pass</p>.
	</Checkbox>
	<Button type="submit">Submit</Button>
</form>
