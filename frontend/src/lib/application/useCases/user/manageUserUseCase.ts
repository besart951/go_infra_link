import type { UserRepository } from '$lib/domain/ports/user/userRepository.js';
import type {
    User,
    UserRole,
    CreateUserRequest,
    UpdateUserRequest
} from '$lib/domain/user/index.js';
import type { AllowedRolesResponse } from '$lib/domain/ports/user/userRepository.js';

export class ManageUserUseCase {
    constructor(private repository: UserRepository) { }

    async get(id: string, signal?: AbortSignal): Promise<User> {
        return this.repository.get(id, signal);
    }

    async getCurrentUser(signal?: AbortSignal): Promise<User> {
        return this.repository.getCurrentUser(signal);
    }

    async create(data: CreateUserRequest, signal?: AbortSignal): Promise<User> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateUserRequest, signal?: AbortSignal): Promise<User> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }

    // ── Admin operations ────────────────────────────────────────────────

    async getAllowedRoles(signal?: AbortSignal): Promise<AllowedRolesResponse> {
        return this.repository.getAllowedRoles(signal);
    }

    async setRole(userId: string, role: UserRole, signal?: AbortSignal): Promise<void> {
        return this.repository.setRole(userId, role, signal);
    }

    async disable(userId: string, signal?: AbortSignal): Promise<void> {
        return this.repository.disable(userId, signal);
    }

    async enable(userId: string, signal?: AbortSignal): Promise<void> {
        return this.repository.enable(userId, signal);
    }
}
