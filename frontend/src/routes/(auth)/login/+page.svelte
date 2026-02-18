<script lang="ts">
	import { Eye, EyeOff } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as InputGroup from '$lib/components/ui/input-group/index.js';
	import { goto, invalidateAll } from '$app/navigation';
	import { api } from '$lib/api/client';
	import {
		Field,
		FieldGroup,
		FieldLabel,
		FieldContent,
		FieldError
	} from '$lib/components/ui/field/index.js';
	import { dev } from '$app/environment';
	import { createTranslator } from '$lib/i18n/translator';

	let showPassword = $state(false);
	let email = $state(dev ? 'besart_morina@hotmail.com' : '');
	let password = $state(dev ? 'password' : '');
	let error = $state<string | null>(null);
	let isLoading = $state(false);
	const t = createTranslator();

	const toogleShowPassword = () => {
		showPassword = !showPassword;
	};

	const errorMessage = (errCode: string | null) => {
		switch (errCode) {
			case 'missing_fields':
				return $t('auth.missing_fields');
			case 'service_unavailable':
				return $t('auth.service_unavailable');
			case 'invalid_credentials':
				return $t('auth.invalid_credentials');
			case 'account_disabled':
				return $t('auth.account_disabled');
			case 'account_locked':
				return $t('auth.account_locked');
			case 'missing_auth_cookies':
				return $t('auth.missing_auth_cookies');
			case 'login_failed':
				return $t('auth.login_failed');
			default:
				if (errCode) return errCode;
				return null;
		}
	};

	async function handleLogin(e: Event) {
		e.preventDefault();
		error = null;
		isLoading = true;

		if (!email || !password) {
			error = 'missing_fields';
			isLoading = false;
			return;
		}

		try {
			// Using backend relative path via proxy
			await api('/auth/login', {
				method: 'POST',
				body: JSON.stringify({ email, password })
			});

			await invalidateAll();
			await goto('/');
		} catch (err: any) {
			error = err.error || 'login_failed';
			console.error('Login error:', err);
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="space-y-6">
	<div class="space-y-2">
		<h1 class="text-2xl font-semibold tracking-tight">{$t('auth.sign_in')}</h1>
		<p class="text-sm text-muted-foreground">{$t('auth.login')}</p>
	</div>

	{#if error}
		<div
			class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
		>
			{errorMessage(error)}
		</div>
	{/if}

	<form onsubmit={handleLogin} class="space-y-6">
		<FieldGroup>
			<Field>
				<FieldLabel for="email" class="text-sm font-medium">{$t('auth.email')}</FieldLabel>
				<FieldContent>
					<Input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						bind:value={email}
						disabled={isLoading}
					/>
				</FieldContent>
			</Field>

			<Field>
				<FieldLabel for="password" class="text-sm font-medium">{$t('auth.password')}</FieldLabel>
				<FieldContent>
					<InputGroup.Root>
						<InputGroup.Input
							id="password"
							name="password"
							type={showPassword ? 'text' : 'password'}
							autocomplete="current-password"
							required
							bind:value={password}
							disabled={isLoading}
						/>
						<InputGroup.Addon align="inline-end">
							<InputGroup.Button
								type="button"
								variant="ghost"
								size="icon-xs"
								onclick={toogleShowPassword}
								aria-pressed={showPassword}
								aria-label={showPassword ? 'Hide password' : 'Show password'}
								disabled={isLoading}
							>
								{#if showPassword}
									<EyeOff class="h-4 w-4" />
								{:else}
									<Eye class="h-4 w-4" />
								{/if}
							</InputGroup.Button>
						</InputGroup.Addon>
					</InputGroup.Root>
				</FieldContent>
			</Field>
		</FieldGroup>

		<Button type="submit" class="w-full" disabled={isLoading}>
			{isLoading ? $t('auth.working') : $t('auth.login')}
		</Button>
	</form>
</div>
