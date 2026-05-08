import {
  createUser,
  deleteUser,
  disableUser,
  enableUser,
  getAllowedRoles,
  getCurrentUser,
  getUserRegistration,
  inviteUser,
  listUserDirectory,
  listUsers,
  resendUserRegistration,
  setUserRole,
  updateCurrentUser,
  updateCurrentUserPassword
} from '$lib/api/users.js';

export type {
  AllowedRole,
  AllowedRolesResponse,
  CreateUserRequest,
  CreateUserInvitationRequest,
  CreateUserInvitationResponse,
  ListUsersParams,
  PaginatedUserResponse,
  RegistrationProcess,
  RegistrationProcessStep,
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
  invite: inviteUser,
  getRegistration: getUserRegistration,
  resendRegistration: resendUserRegistration,
  setRole: setUserRole,
  disable: disableUser,
  enable: enableUser,
  delete: deleteUser,
  updateCurrent: updateCurrentUser,
  updateCurrentPassword: updateCurrentUserPassword
};
