<script lang="ts">
	/**
	 * User Management Form
	 *
	 * Form for creating/editing users with role selection
	 * filtered by current user's permissions
	 */

	import type { UserRole } from '$lib/api/users';
	import { createUser } from '$lib/api/users';
	import { auth, getAllowedRolesForCreation } from '$lib/stores/auth.svelte';
	import { getRoleLabel } from '$lib/utils/permissions';
	import { getErrorMessage, getFieldErrors } from '$lib/api/client';

	interface Props {
		onSuccess?: () => void;
		onCancel?: () => void;
	}

	let { onSuccess, onCancel }: Props = $props();

	// Form state
	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let password = $state('');
	let isActive = $state(true);
	let selectedRole = $state<UserRole | ''>('');

	let isSubmitting = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	// Get allowed roles for the current user
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

			// Reset form
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
	<h2 class="mb-4 text-2xl font-bold">Create User</h2>

	{#if error}
		<div class="rounded-md bg-red-50 p-4 text-red-800">
			{error}
		</div>
	{/if}

	<div class="grid grid-cols-2 gap-4">
		<div>
			<label for="firstName" class="mb-1 block text-sm font-medium"> First Name </label>
			<input
				type="text"
				id="firstName"
				bind:value={firstName}
				required
				class="w-full rounded-md border px-3 py-2"
				class:border-red-500={fieldErrors.first_name}
			/>
			{#if fieldErrors.first_name}
				<p class="mt-1 text-sm text-red-600">{fieldErrors.first_name}</p>
			{/if}
		</div>

		<div>
			<label for="lastName" class="mb-1 block text-sm font-medium"> Last Name </label>
			<input
				type="text"
				id="lastName"
				bind:value={lastName}
				required
				class="w-full rounded-md border px-3 py-2"
				class:border-red-500={fieldErrors.last_name}
			/>
			{#if fieldErrors.last_name}
				<p class="mt-1 text-sm text-red-600">{fieldErrors.last_name}</p>
			{/if}
		</div>
	</div>

	<div>
		<label for="email" class="mb-1 block text-sm font-medium"> Email </label>
		<input
			type="email"
			id="email"
			bind:value={email}
			required
			class="w-full rounded-md border px-3 py-2"
			class:border-red-500={fieldErrors.email}
		/>
		{#if fieldErrors.email}
			<p class="mt-1 text-sm text-red-600">{fieldErrors.email}</p>
		{/if}
	</div>

	<div>
		<label for="password" class="mb-1 block text-sm font-medium"> Password </label>
		<input
			type="password"
			id="password"
			bind:value={password}
			required
			minlength="8"
			class="w-full rounded-md border px-3 py-2"
			class:border-red-500={fieldErrors.password}
		/>
		{#if fieldErrors.password}
			<p class="mt-1 text-sm text-red-600">{fieldErrors.password}</p>
		{/if}
	</div>

	<div>
		<label for="role" class="mb-1 block text-sm font-medium"> Role </label>
		<select
			id="role"
			bind:value={selectedRole}
			required
			class="w-full rounded-md border px-3 py-2"
			class:border-red-500={fieldErrors.role}
		>
			<option value="">Select a role</option>
			{#each allowedRoles as role}
				<option value={role}>{getRoleLabel(role)}</option>
			{/each}
		</select>
		{#if fieldErrors.role}
			<p class="mt-1 text-sm text-red-600">{fieldErrors.role}</p>
		{/if}
		<p class="mt-1 text-sm text-gray-500">
			You can only assign roles that you have permission to manage
		</p>
	</div>

	<div class="flex items-center">
		<input
			type="checkbox"
			id="isActive"
			bind:checked={isActive}
			class="h-4 w-4 rounded text-blue-600"
		/>
		<label for="isActive" class="ml-2 text-sm"> User is active </label>
	</div>

	<div class="flex justify-end gap-2 pt-4">
		{#if onCancel}
			<button
				type="button"
				onclick={onCancel}
				class="rounded-md border px-4 py-2 hover:bg-gray-50"
				disabled={isSubmitting}
			>
				Cancel
			</button>
		{/if}
		<button
			type="submit"
			disabled={isSubmitting}
			class="rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
		>
			{isSubmitting ? 'Creating...' : 'Create User'}
		</button>
	</div>
</form>
