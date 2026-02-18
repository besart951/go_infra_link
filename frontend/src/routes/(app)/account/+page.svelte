<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { addToast } from '$lib/components/toast.svelte';
	import {
		getCurrentUser,
		updateCurrentUser,
		updateCurrentUserPassword,
		type User,
		type UpdateUserRequest
	} from '$lib/api/users.js';
	import { listTeams, listTeamMembers } from '$lib/api/teams.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import { createTranslator } from '$lib/i18n/translator';
	import { setThemePreference, themePreference, type ThemePreference } from '$lib/stores/theme.js';
	import { LaptopMinimal, Moon, Sun } from '@lucide/svelte';

	type AccountTab = 'information' | 'notifications' | 'password' | 'preferences';
	type ThemeOption = {
		value: ThemePreference;
		label: string;
		description: string;
		icon: typeof LaptopMinimal;
	};

	const t = createTranslator();

	let activeTab = $state<AccountTab>('information');
	let currentUser = $state<User | null>(null);

	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');

	let newPassword = $state('');
	let confirmPassword = $state('');

	let isSavingProfile = $state(false);
	let isSavingPassword = $state(false);
	let isLoading = $state(true);

	let userTeams = $state<string[]>([]);
	let teamsError = $state<string | null>(null);

	const options: ThemeOption[] = [
		{
			value: 'system',
			label: $t('pages.settings_theme_system'),
			description: $t('pages.settings_theme_system_desc'),
			icon: LaptopMinimal
		},
		{
			value: 'light',
			label: $t('pages.settings_theme_light'),
			description: $t('pages.settings_theme_light_desc'),
			icon: Sun
		},
		{
			value: 'dark',
			label: $t('pages.settings_theme_dark'),
			description: $t('pages.settings_theme_dark_desc'),
			icon: Moon
		}
	];

	const permissions = $derived(currentUser?.permissions ?? []);

	const accountTabs: { value: AccountTab; label: string }[] = [
		{ value: 'information', label: $t('pages.account_tabs_information') },
		{ value: 'notifications', label: $t('pages.account_tabs_notifications') },
		{ value: 'password', label: $t('pages.account_tabs_password') },
		{ value: 'preferences', label: $t('pages.account_tabs_preferences') }
	];

	function applyUserToForm(user: User) {
		firstName = user.first_name;
		lastName = user.last_name;
		email = user.email;
	}

	async function loadUserTeams(userId: string) {
		teamsError = null;
		userTeams = [];

		try {
			const teamsResponse = await listTeams({ page: 1, limit: 100, search: '' });
			const memberLists = await Promise.all(
				teamsResponse.items.map(async (team) => ({
					team,
					members: await listTeamMembers(team.id, { page: 1, limit: 1000 })
				}))
			);

			userTeams = memberLists
				.filter((entry) => entry.members.items.some((m) => m.user_id === userId))
				.map((entry) => entry.team.name);
		} catch (err) {
			teamsError = getErrorMessage(err);
		}
	}

	async function loadAccount() {
		isLoading = true;
		try {
			const user = await getCurrentUser();
			currentUser = user;
			applyUserToForm(user);
			await loadUserTeams(user.id);
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		} finally {
			isLoading = false;
		}
	}

	async function handleInformationSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!currentUser) return;

		isSavingProfile = true;
		try {
			const payload: UpdateUserRequest = {
				first_name: firstName,
				last_name: lastName,
				email
			};
			const updated = await updateCurrentUser(currentUser.id, payload);
			currentUser = updated;
			applyUserToForm(updated);
			addToast($t('messages.account_info_saved'), 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		} finally {
			isSavingProfile = false;
		}
	}

	async function handlePasswordSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!currentUser) return;

		if (newPassword.length < 8) {
			addToast(
				$t('validation.password_too_short', { field: $t('auth.password'), min: 8 }),
				'error'
			);
			return;
		}

		if (newPassword !== confirmPassword) {
			addToast(
				$t('validation.must_match', {
					field1: $t('auth.new_password'),
					field2: $t('auth.confirm_password')
				}),
				'error'
			);
			return;
		}

		isSavingPassword = true;
		try {
			await updateCurrentUserPassword(currentUser.id, newPassword);
			newPassword = '';
			confirmPassword = '';
			addToast($t('messages.account_password_saved'), 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		} finally {
			isSavingPassword = false;
		}
	}

	onMount(() => {
		loadAccount();
	});
</script>

<svelte:head>
	<title>{$t('navigation.account')} | {$t('app.brand')}</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-3xl font-bold tracking-tight">{$t('navigation.account')}</h1>
		<p class="mt-1 text-muted-foreground">{$t('pages.account_desc')}</p>
	</div>

	<div class="grid gap-6 md:grid-cols-[220px_minmax(0,1fr)]">
		<aside>
			<div class="rounded-lg border bg-card p-2">
				<nav class="flex flex-col gap-1">
					{#each accountTabs as tab (tab.value)}
						<button
							type="button"
							class={activeTab === tab.value
								? 'rounded-md bg-muted px-3 py-2 text-left text-sm font-medium'
								: 'rounded-md px-3 py-2 text-left text-sm text-muted-foreground hover:bg-muted/60'}
							onclick={() => (activeTab = tab.value)}
						>
							{tab.label}
						</button>
					{/each}
				</nav>
			</div>
		</aside>

		<section>
			{#if isLoading}
				<div class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
					{$t('common.loading')}
				</div>
			{:else if !currentUser}
				<div class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
					{$t('errors.unknown_error')}
				</div>
			{:else if activeTab === 'information'}
				<div class="flex flex-col gap-4">
					<form class="rounded-lg border bg-card p-4" onsubmit={handleInformationSubmit}>
						<div class="mb-4">
							<h2 class="text-base font-semibold">{$t('pages.account_information_title')}</h2>
							<p class="text-sm text-muted-foreground">{$t('pages.account_information_desc')}</p>
						</div>

						<div class="grid gap-4 sm:grid-cols-2">
							<div class="flex flex-col gap-2">
								<label for="first_name" class="text-sm font-medium">{$t('user.firstname')}</label>
								<Input
									id="first_name"
									bind:value={firstName}
									required
									minlength={1}
									maxlength={100}
								/>
							</div>
							<div class="flex flex-col gap-2">
								<label for="last_name" class="text-sm font-medium">{$t('user.lastname')}</label>
								<Input
									id="last_name"
									bind:value={lastName}
									required
									minlength={1}
									maxlength={100}
								/>
							</div>
						</div>

						<div class="mt-4 flex flex-col gap-2">
							<label for="email" class="text-sm font-medium">{$t('user.email')}</label>
							<Input id="email" type="email" bind:value={email} required />
						</div>

						<div class="mt-4 flex justify-end">
							<Button type="submit" disabled={isSavingProfile}>
								{isSavingProfile ? $t('common.saving') : $t('common.save_changes')}
							</Button>
						</div>
					</form>

					<div class="rounded-lg border bg-card p-4">
						<h3 class="text-base font-semibold">{$t('pages.account_access_title')}</h3>
						<div class="mt-3 grid gap-4 sm:grid-cols-2">
							<div>
								<p class="text-sm text-muted-foreground">{$t('common.role')}</p>
								<p class="mt-1 font-medium">{currentUser.role}</p>
							</div>
							<div>
								<p class="text-sm text-muted-foreground">
									{$t('roles.permissions.table.permission')}
								</p>
								<div class="mt-2 flex flex-wrap gap-2">
									{#if permissions.length === 0}
										<span class="text-sm text-muted-foreground"
											>{$t('pages.account_permissions_empty')}</span
										>
									{:else}
										{#each permissions as permission (permission)}
											<Badge variant="outline">{permission}</Badge>
										{/each}
									{/if}
								</div>
							</div>
						</div>

						<div class="mt-4">
							<p class="text-sm text-muted-foreground">{$t('navigation.teams')}</p>
							<div class="mt-2 flex flex-wrap gap-2">
								{#if teamsError}
									<span class="text-sm text-muted-foreground">{teamsError}</span>
								{:else if userTeams.length === 0}
									<span class="text-sm text-muted-foreground"
										>{$t('pages.account_teams_empty')}</span
									>
								{:else}
									{#each userTeams as teamName (teamName)}
										<Badge variant="secondary">{teamName}</Badge>
									{/each}
								{/if}
							</div>
						</div>
					</div>
				</div>
			{:else if activeTab === 'notifications'}
				<div class="rounded-lg border bg-card p-6">
					<h2 class="text-base font-semibold">{$t('pages.account_notifications_title')}</h2>
					<p class="mt-2 text-sm text-muted-foreground">
						{$t('pages.account_notifications_empty')}
					</p>
				</div>
			{:else if activeTab === 'password'}
				<form class="rounded-lg border bg-card p-4" onsubmit={handlePasswordSubmit}>
					<div class="mb-4">
						<h2 class="text-base font-semibold">{$t('pages.account_password_title')}</h2>
						<p class="text-sm text-muted-foreground">{$t('pages.account_password_desc')}</p>
					</div>

					<div class="grid gap-4 sm:max-w-md">
						<div class="flex flex-col gap-2">
							<label for="new_password" class="text-sm font-medium">{$t('auth.new_password')}</label
							>
							<Input
								id="new_password"
								type="password"
								bind:value={newPassword}
								required
								minlength={8}
							/>
						</div>
						<div class="flex flex-col gap-2">
							<label for="confirm_password" class="text-sm font-medium">
								{$t('auth.confirm_password')}
							</label>
							<Input
								id="confirm_password"
								type="password"
								bind:value={confirmPassword}
								required
								minlength={8}
							/>
						</div>
					</div>

					<div class="mt-4 flex justify-end">
						<Button type="submit" disabled={isSavingPassword}>
							{isSavingPassword ? $t('common.saving') : $t('pages.account_password_save')}
						</Button>
					</div>
				</form>
			{:else if activeTab === 'preferences'}
				<div class="rounded-lg border bg-card p-4">
					<div class="flex flex-col gap-1">
						<h2 class="text-base font-semibold">{$t('pages.settings_appearance')}</h2>
						<p class="text-sm text-muted-foreground">{$t('pages.settings_appearance_desc')}</p>
					</div>

					<div class="mt-4 grid gap-2 sm:grid-cols-3">
						{#each options as opt (opt.value)}
							{@const active = $themePreference === opt.value}
							<Button
								variant={active ? 'default' : 'outline'}
								class="h-full items-start justify-start gap-3 px-4 py-3 text-left whitespace-normal"
								onclick={() => setThemePreference(opt.value)}
							>
								<opt.icon class="mt-0.5 size-4 shrink-0" />
								<span class="flex min-w-0 flex-col items-start gap-0.5 text-left">
									<span class="leading-tight">{opt.label}</span>
									<span
										class={active
											? 'text-xs leading-snug wrap-break-word text-primary-foreground/80'
											: 'text-xs leading-snug wrap-break-word text-muted-foreground'}
									>
										{opt.description}
									</span>
								</span>
							</Button>
						{/each}
					</div>
				</div>
			{/if}
		</section>
	</div>
</div>
