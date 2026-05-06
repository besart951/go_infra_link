export type TimelineDateTimeBoundary = 'start' | 'end';

function defaultTime(boundary: TimelineDateTimeBoundary): string {
  return boundary === 'start' ? '00:00' : '23:59';
}

export function buildTimelineDateTimeISO(
  date: string,
  time: string,
  boundary: TimelineDateTimeBoundary
): string | undefined {
  if (!date) return undefined;

  const normalizedTime = time || defaultTime(boundary);
  const parsed = new Date(`${date}T${normalizedTime}`);
  if (Number.isNaN(parsed.getTime())) return undefined;

  return parsed.toISOString();
}
