<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { setThemePreference, themePreference, type ThemePreference } from '$lib/stores/theme.js';
	import { LaptopMinimal, Moon, Sun } from '@lucide/svelte';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

	type ThemeOption = {
		value: ThemePreference;
		label: string;
		description: string;
		icon: typeof LaptopMinimal;
	};

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
</script>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-3xl font-bold tracking-tight">{$t('pages.settings')}</h1>
		<p class="mt-2 text-sm text-muted-foreground">{$t('pages.settings_desc')}</p>
	</div>

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
					class="h-auto justify-start gap-3 px-4 py-3"
					onclick={() => setThemePreference(opt.value)}
				>
					<svelte:component this={opt.icon} />
					<span class="flex flex-col items-start gap-0.5 text-left">
						<span class="leading-tight">{opt.label}</span>
						<span
							class={active
								? 'text-xs text-primary-foreground/80'
								: 'text-xs text-muted-foreground'}
						>
							{opt.description}
						</span>
					</span>
				</Button>
			{/each}
		</div>
	</div>
</div>
