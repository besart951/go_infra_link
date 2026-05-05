import { browser } from '$app/environment';
import { goto } from '$app/navigation';

export async function navigateBack(fallbackHref = '/'): Promise<void> {
  if (browser && window.history.length > 1) {
    window.history.back();
    return;
  }

  await goto(fallbackHref);
}
