/**
 * User API repository
 * Infrastructure layer - implements UserRepository port via HTTP
 * Consolidates the former user.adapter.ts and $lib/api/users.ts
 */
import type {
    UserRepository,
    AllowedRolesResponse
} from '$lib/domain/ports/user/userRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    User,
    UserRole,
    UserListResponse,
    CreateUserRequest,
    UpdateUserRequest
} from '$lib/domain/user/index.js';
import { api } from '$lib/api/client.js';

function buildQuery(params: Record<string, string | number | boolean | undefined>): string {
    const searchParams = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
        if (value !== undefined && value !== '') {
            searchParams.set(key, String(value));
        }
    }
    const query = searchParams.toString();
    return query ? `?${query}` : '';
}

export const userRepository: UserRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<User>> {
        const searchParams = new URLSearchParams();
        searchParams.set('page', String(params.pagination.page));
        searchParams.set('limit', String(params.pagination.pageSize));
        if (params.search.text) searchParams.set('search', params.search.text);

        if (params.filters) {
            Object.entries(params.filters).forEach(([key, value]) => {
                if (value !== undefined && value !== null) searchParams.set(key, value);
            });
        }

        const query = searchParams.toString();
        const response = await api<UserListResponse>(`/users${query ? `?${query}` : ''}`, { signal });

        return {
            items: response.items,
            metadata: {
                total: response.total,
                page: response.page,
                pageSize: params.pagination.pageSize,
                totalPages: response.total_pages
            }
        };
    },

    async get(id: string, signal?: AbortSignal): Promise<User> {
        return api<User>(`/users/${id}`, { signal });
    },

    async getCurrentUser(signal?: AbortSignal): Promise<User> {
        return api<User>('/auth/me', { signal });
    },

    async create(data: CreateUserRequest, signal?: AbortSignal): Promise<User> {
        return api<User>('/users', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateUserRequest, signal?: AbortSignal): Promise<User> {
        return api<User>(`/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/users/${id}`, { method: 'DELETE', signal });
    },

    // ── Admin operations ────────────────────────────────────────────────

    async getAllowedRoles(signal?: AbortSignal): Promise<AllowedRolesResponse> {
        return api<AllowedRolesResponse>('/users/allowed-roles', { signal });
    },

    async setRole(userId: string, role: UserRole, signal?: AbortSignal): Promise<void> {
        return api<void>(`/admin/users/${userId}/role`, {
            method: 'POST',
            body: JSON.stringify({ role }),
            signal
        });
    },

    async disable(userId: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/admin/users/${userId}/disable`, { method: 'POST', signal });
    },

    async enable(userId: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/admin/users/${userId}/enable`, { method: 'POST', signal });
    }
};
