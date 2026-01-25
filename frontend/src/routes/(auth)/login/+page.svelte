<script lang="ts">
	import { Eye, EyeOff } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as InputGroup from '$lib/components/ui/input-group/index.js';
	import type { ActionData } from './$types.js';
	import {
		Field,
		FieldGroup,
		FieldLabel,
		FieldContent,
		FieldError
	} from '$lib/components/ui/field/index.js';
	import { dev } from '$app/environment';

	export let form: ActionData;
	let showPassword = false;
	let email = dev ? 'besart_morina@hotmail.com' : '';
	let password = dev ? 'password' : '';

	const toogleShowPassword = () => {
		showPassword = !showPassword;
	};

	const errorMessage = (error?: string) => {
		if (form?.message && form.error !== 'missing_fields' && form.error !== 'service_unavailable') {
			return form.message;
		}
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

	<form method="POST" class="space-y-6">
		<FieldGroup>
			<Field>
				<FieldLabel for="email" class="text-sm font-medium">Email</FieldLabel>
				<FieldContent>
					<Input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						bind:value={email}
					/>
					{#if form?.error === 'missing_fields'}
						<FieldError errors={[{ message: 'Email is required.' }]} />
					{/if}
				</FieldContent>
			</Field>

			<Field>
				<FieldLabel for="password" class="text-sm font-medium">Password</FieldLabel>
				<FieldContent>
					<InputGroup.Root>
						<InputGroup.Input
							id="password"
							name="password"
							type={showPassword ? 'text' : 'password'}
							autocomplete="current-password"
							required
							bind:value={password}
						/>
						<InputGroup.Addon align="inline-end">
							<InputGroup.Button
								type="button"
								variant="ghost"
								size="icon-xs"
								onclick={toogleShowPassword}
								aria-pressed={showPassword}
								aria-label={showPassword ? 'Hide password' : 'Show password'}
								title={showPassword ? 'Hide password' : 'Show password'}
							>
								{#if showPassword}
									<EyeOff class="size-4" />
								{:else}
									<Eye class="size-4" />
								{/if}
							</InputGroup.Button>
						</InputGroup.Addon>
					</InputGroup.Root>
					{#if form?.error && form.error !== 'missing_fields'}
						<FieldError errors={[{ message: errorMessage(form.error) }]} />
					{/if}
				</FieldContent>
			</Field>
		</FieldGroup>

		<Button type="submit" class="w-full">Sign in</Button>
	</form>
</div>
