import type {
    User,
    UserRole,
    CreateUserRequest,
    UpdateUserRequest
} from '$lib/domain/user/index.js';
import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export interface AllowedRolesResponse {
    roles: UserRole[];
}

export interface UserRepository extends ListRepository<User> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<User>>;
    get(id: string, signal?: AbortSignal): Promise<User>;
    getCurrentUser(signal?: AbortSignal): Promise<User>;
    create(data: CreateUserRequest, signal?: AbortSignal): Promise<User>;
    update(id: string, data: UpdateUserRequest, signal?: AbortSignal): Promise<User>;
    delete(id: string, signal?: AbortSignal): Promise<void>;

    // Admin operations
    getAllowedRoles(signal?: AbortSignal): Promise<AllowedRolesResponse>;
    setRole(userId: string, role: UserRole, signal?: AbortSignal): Promise<void>;
    disable(userId: string, signal?: AbortSignal): Promise<void>;
    enable(userId: string, signal?: AbortSignal): Promise<void>;
}
