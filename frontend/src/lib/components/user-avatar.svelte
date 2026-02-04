<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { cn } from '$lib/utils.js';

	interface Props {
		firstName: string;
		lastName: string;
		class?: string;
	}

	let { firstName, lastName, class: className }: Props = $props();

	function initials(first: string, last: string): string {
		return `${first?.[0] ?? ''}${last?.[0] ?? ''}`.toUpperCase();
	}
</script>

<Avatar.Root class={cn('h-8 w-8', className)}>
	<Avatar.Fallback class="text-xs">
		{initials(firstName, lastName)}
	</Avatar.Fallback>
</Avatar.Root>
