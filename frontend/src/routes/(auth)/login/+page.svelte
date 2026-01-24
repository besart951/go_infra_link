<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import type { ActionData } from './$types.js';

	let { form }: { form?: ActionData } = $props();

	const errorMessage = (error?: string) => {
		switch (error) {
			case 'missing_fields':
				return 'Email and password are required.';
			case 'service_unavailable':
				return 'Service is currently unavailable. Please try again in a moment.';
			case 'invalid_credentials':
				return 'Invalid email or password.';
			case 'account_disabled':
				return 'Your account is disabled.';
			case 'account_locked':
				return 'Your account is locked.';
			case 'missing_auth_cookies':
				return 'Login succeeded but tokens were missing.';
			default:
				return 'Login failed.';
		}
	};
</script>

<div class="space-y-6">
	<div class="space-y-2">
		<h1 class="text-2xl font-semibold tracking-tight">Sign in</h1>
		<p class="text-sm text-muted-foreground">Use your account credentials.</p>
	</div>

	{#if form?.error}
		<div
			class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
		>
			{errorMessage(form.error)}
		</div>
	{/if}

	<form method="POST" class="space-y-4">
		<div class="space-y-2">
			<label class="text-sm font-medium" for="email">Email</label>
			<Input id="email" name="email" type="email" autocomplete="email" required />
		</div>
		<div class="space-y-2">
			<label class="text-sm font-medium" for="password">Password</label>
			<Input
				id="password"
				name="password"
				type="password"
				autocomplete="current-password"
				required
			/>
		</div>
		<Button type="submit" class="w-full">Sign in</Button>
	</form>
</div>
