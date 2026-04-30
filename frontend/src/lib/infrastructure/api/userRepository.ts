import {
  createUser,
  deleteUser,
  disableUser,
  enableUser,
  getAllowedRoles,
  getCurrentUser,
  listUserDirectory,
  listUsers,
  setUserRole,
  updateCurrentUser,
  updateCurrentUserPassword
} from '$lib/api/users.js';

export type {
  AllowedRole,
  AllowedRolesResponse,
  CreateUserRequest,
  ListUsersParams,
  PaginatedUserResponse,
  UpdateUserRequest,
  User,
  UserDirectoryPageCapabilities,
  UserDirectoryResponse,
  UserDirectoryTeam,
  UserDirectoryTeamFilter,
  UserDirectoryUser,
  UserRole
} from '$lib/api/users.js';

export const userRepository = {
  list: listUsers,
  listDirectory: listUserDirectory,
  getCurrent: getCurrentUser,
  getAllowedRoles,
  create: createUser,
  setRole: setUserRole,
  disable: disableUser,
  enable: enableUser,
  delete: deleteUser,
  updateCurrent: updateCurrentUser,
  updateCurrentPassword: updateCurrentUserPassword
};
