<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { createTranslator } from '$lib/i18n/translator';

  interface Props {
    newPassword: string;
    confirmPassword: string;
    isSavingPassword: boolean;
    onSubmit: (event: SubmitEvent) => void | Promise<void>;
  }

  let {
    newPassword = $bindable(),
    confirmPassword = $bindable(),
    isSavingPassword,
    onSubmit
  }: Props = $props();

  const t = createTranslator();
</script>

<form class="rounded-lg border bg-card p-4" onsubmit={onSubmit}>
  <div class="mb-4">
    <h2 class="text-base font-semibold">{$t('pages.account_password_title')}</h2>
    <p class="text-sm text-muted-foreground">{$t('pages.account_password_desc')}</p>
  </div>

  <div class="grid gap-4 sm:max-w-md">
    <div class="flex flex-col gap-2">
      <label for="new_password" class="text-sm font-medium">{$t('auth.new_password')}</label>
      <Input id="new_password" type="password" bind:value={newPassword} required minlength={8} />
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
