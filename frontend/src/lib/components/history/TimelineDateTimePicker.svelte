<script lang="ts">
  import { Calendar } from '$lib/components/ui/calendar/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { cn } from '$lib/utils.js';
  import { getLocalTimeZone, parseDate, type DateValue } from '@internationalized/date';
  import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';

  interface Props {
    id: string;
    label: string;
    timeLabel: string;
    defaultTime: string;
    date?: string;
    time?: string;
    placeholder?: string;
    disabled?: boolean;
    class?: string;
  }

  let {
    id,
    label,
    timeLabel,
    defaultTime,
    date = $bindable(''),
    time = $bindable(''),
    placeholder = 'Datum wählen',
    disabled = false,
    class: className
  }: Props = $props();

  let open = $state(false);
  let calendarValue = $state<DateValue | undefined>(parseCalendarDate(date));

  const formattedDate = $derived(
    calendarValue
      ? new Intl.DateTimeFormat('de-CH').format(calendarValue.toDate(getLocalTimeZone()))
      : placeholder
  );

  $effect(() => {
    const nextValue = parseCalendarDate(date);
    if (calendarValue?.toString() !== nextValue?.toString()) {
      calendarValue = nextValue;
    }
  });

  function parseCalendarDate(value: string): DateValue | undefined {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return undefined;

    try {
      return parseDate(value);
    } catch {
      return undefined;
    }
  }

  function selectDate(value: DateValue | undefined): void {
    calendarValue = value;
    date = value?.toString() ?? '';

    if (!date) {
      time = '';
      return;
    }

    if (!time) {
      time = defaultTime;
    }

    open = false;
  }
</script>

<div class={cn('min-w-0 space-y-1.5', className)}>
  <Label for={id}>{label}</Label>
  <div class="grid grid-cols-[minmax(8.75rem,1fr)_7.5rem] gap-2">
    <Popover.Root bind:open>
      <Popover.Trigger>
        {#snippet child({ props })}
          <Button
            {...props}
            {id}
            variant="outline"
            {disabled}
            aria-expanded={open}
            class="h-10 w-full justify-between overflow-hidden px-3 font-normal"
          >
            <span class={cn('min-w-0 truncate', !date && 'text-muted-foreground')}>
              {formattedDate}
            </span>
            <ChevronDownIcon class="size-4 opacity-50" />
          </Button>
        {/snippet}
      </Popover.Trigger>
      <Popover.Content class="w-auto overflow-hidden p-0" align="start">
        <Calendar
          type="single"
          bind:value={calendarValue}
          onValueChange={selectDate}
          captionLayout="dropdown"
          locale="de-CH"
        />
      </Popover.Content>
    </Popover.Root>

    <Input
      id={`${id}_time`}
      type="time"
      bind:value={time}
      step="60"
      disabled={disabled || !date}
      aria-label={timeLabel}
      class="h-10 appearance-none bg-background text-sm tabular-nums [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
    />
  </div>
</div>
