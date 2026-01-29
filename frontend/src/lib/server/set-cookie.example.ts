/**
 * Server-side utility to parse Set-Cookie headers
 * Copy this file to set-cookie.ts and configure as needed
 */

export function getSetCookieValues(headers: Headers): string[] {
	const setCookies: string[] = [];
	headers.forEach((value, key) => {
		if (key.toLowerCase() === 'set-cookie') {
			setCookies.push(value);
		}
	});
	return setCookies;
}
