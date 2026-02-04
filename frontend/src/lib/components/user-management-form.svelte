<script lang="ts">
	import type { UserRole } from '$lib/api/users';
	import { createUser } from '$lib/api/users';
	import { getAllowedRolesForCreation } from '$lib/stores/auth.svelte';
	import { getRoleLabel } from '$lib/utils/permissions';
	import { getErrorMessage, getFieldErrors } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import { Check } from '@lucide/svelte';
	import { cn } from '$lib/utils';

	interface Props {
		onSuccess?: () => void;
		onCancel?: () => void;
	}

	let { onSuccess, onCancel }: Props = $props();

	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let password = $state('');
	let isActive = $state(true);
	let selectedRole = $state<UserRole | ''>('');
	let openCombobox = $state(false);

	let isSubmitting = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	const allowedRoles = $derived(getAllowedRolesForCreation());

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!selectedRole) {
			error = 'Please select a role';
			return;
		}

		isSubmitting = true;
		error = '';
		fieldErrors = {};

		try {
			await createUser({
				first_name: firstName,
				last_name: lastName,
				email,
				password,
				is_active: isActive,
				role: selectedRole
			});

			firstName = '';
			lastName = '';
			email = '';
			password = '';
			isActive = true;
			selectedRole = '';

			if (onSuccess) onSuccess();
		} catch (err) {
			error = getErrorMessage(err);
			fieldErrors = getFieldErrors(err);
		} finally {
			isSubmitting = false;
		}
	}
</script>

<form onsubmit={handleSubmit} class="space-y-4">
	{#if error}
		<div class="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
			{error}
		</div>
	{/if}

	<div class="grid grid-cols-2 gap-4">
		<div class="space-y-2">
			<Label for="firstName">First Name</Label>
			<Input
				type="text"
				id="firstName"
				bind:value={firstName}
				required
				class={fieldErrors.first_name ? 'border-destructive' : ''}
			/>
			{#if fieldErrors.first_name}
				<p class="text-sm text-destructive">{fieldErrors.first_name}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label for="lastName">Last Name</Label>
			<Input
				type="text"
				id="lastName"
				bind:value={lastName}
				required
				class={fieldErrors.last_name ? 'border-destructive' : ''}
			/>
			{#if fieldErrors.last_name}
				<p class="text-sm text-destructive">{fieldErrors.last_name}</p>
			{/if}
		</div>
	</div>

	<div class="space-y-2">
		<Label for="email">Email</Label>
		<Input
			type="email"
			id="email"
			bind:value={email}
			required
			class={fieldErrors.email ? 'border-destructive' : ''}
		/>
		{#if fieldErrors.email}
			<p class="text-sm text-destructive">{fieldErrors.email}</p>
		{/if}
	</div>

	<div class="space-y-2">
		<Label for="password">Password</Label>
		<Input
			type="password"
			id="password"
			bind:value={password}
			required
			minlength={8}
			class={fieldErrors.password ? 'border-destructive' : ''}
		/>
		{#if fieldErrors.password}
			<p class="text-sm text-destructive">{fieldErrors.password}</p>
		{/if}
	</div>

	<div class="space-y-2">
		<Label for="role">Role</Label>
		<Popover.Root bind:open={openCombobox}>
			<Popover.Trigger asChild let:builder>
				<Button
					variant="outline"
					role="combobox"
					aria-expanded={openCombobox}
					class={cn(
						'w-full justify-between',
						!selectedRole && 'text-muted-foreground',
						fieldErrors.role && 'border-destructive'
					)}
					builders={[builder]}
				>
					{selectedRole ? getRoleLabel(selectedRole) : 'Select a role...'}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="ml-2 h-4 w-4 shrink-0 opacity-50"
					>
						<path d="m7 15 5 5 5-5" />
						<path d="m7 9 5-5 5 5" />
					</svg>
				</Button>
			</Popover.Trigger>
			<Popover.Content class="w-full p-0 bg-background text-foreground" side="bottom" align="start">
				<Command.Root>
					<Command.Input placeholder="Search roles..." />
					<Command.Empty>No role found.</Command.Empty>
					<Command.List>
						{#each allowedRoles as role (role)}
							<Command.Item
								value={role}
								onSelect={(currentValue) => {
									selectedRole = currentValue as UserRole;
									openCombobox = false;
								}}
							>
								<Check class={cn('mr-2 h-4 w-4', selectedRole !== role && 'text-transparent')} />
								{getRoleLabel(role)}
							</Command.Item>
						{/each}
					</Command.List>
				</Command.Root>
			</Popover.Content>
		</Popover.Root>
		{#if fieldErrors.role}
			<p class="text-sm text-destructive">{fieldErrors.role}</p>
		{/if}
		<p class="text-sm text-muted-foreground">
			You can only assign roles that you have permission to manage
		</p>
	</div>

	<div class="flex items-center gap-2">
		<Checkbox id="isActive" checked={isActive} onCheckedChange={(v) => (isActive = !!v)} />
		<Label for="isActive" class="text-sm font-normal">User is active</Label>
	</div>

	<div class="flex justify-end gap-2 pt-2">
		{#if onCancel}
			<Button type="button" variant="outline" onclick={onCancel} disabled={isSubmitting}>
				Cancel
			</Button>
		{/if}
		<Button type="submit" disabled={isSubmitting}>
			{isSubmitting ? 'Creating...' : 'Create User'}
		</Button>
	</div>
</form>
