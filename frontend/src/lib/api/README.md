/\*\*

- Example: How to use the centralized API client
-
- Import the `api` function and use it directly.
- CSRF tokens are handled automatically!
  \*/

import { api, ApiException, getErrorMessage, isApiError } from '$lib/api/client';

// ============================================================================
// FETCH DATA
// ============================================================================

// Simple GET request
async function fetchUser() {
try {
const user = await api('/users/me');
console.log('User:', user);
} catch (err) {
if (isApiError(err, 'unauthorized')) {
console.log('User is not authenticated');
} else {
console.error('Error:', getErrorMessage(err));
}
}
}

// ============================================================================
// POST DATA
// ============================================================================

// Create a new team (CSRF token included automatically!)
async function createTeam(name: string, description?: string) {
try {
const team = await api('/teams', {
method: 'POST',
body: JSON.stringify({ name, description })
});
console.log('Team created:', team);
return team;
} catch (err) {
if (err instanceof ApiException) {
console.error(`Error ${err.status}: ${err.error}`);
console.error('Message:', err.message);
}
}
}

// ============================================================================
// UPDATE DATA
// ============================================================================

async function updateTeam(teamId: string, name: string) {
try {
const updated = await api(`/teams/${teamId}`, {
method: 'PATCH',
body: JSON.stringify({ name })
});
return updated;
} catch (err) {
console.error('Update failed:', getErrorMessage(err));
}
}

// ============================================================================
// DELETE DATA (returns no content)
// ============================================================================

async function deleteTeam(teamId: string) {
try {
await api(`/teams/${teamId}`, { method: 'DELETE' });
console.log('Team deleted');
} catch (err) {
if (isApiError(err, 'not_found')) {
console.log('Team not found');
} else {
console.error('Delete failed:', getErrorMessage(err));
}
}
}

// ============================================================================
// IN SVELTE COMPONENTS
// ============================================================================

// Example: In a +page.ts or +layout.ts (server-side)
// You can still use `fetch()` with headers manually, or import api()

// Example: In a .svelte component (client-side)
/\*

<script>
	import { api, getErrorMessage } from '$lib/api/client';
	
	let loading = false;
	let error = '';
	
	async function handleCreate() {
		loading = true;
		error = '';
		try {
			const result = await api('/teams', {
				method: 'POST',
				body: JSON.stringify({ name: 'My Team' })
			});
			console.log('Created:', result);
		} catch (err) {
			error = getErrorMessage(err);
		} finally {
			loading = false;
		}
	}
</script>

<button on:click={handleCreate} disabled={loading}>
	{loading ? 'Creating...' : 'Create Team'}
</button>
{#if error}<p class="error">{error}</p>{/if}
*/
