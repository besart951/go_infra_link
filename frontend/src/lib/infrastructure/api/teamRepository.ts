import {
  addTeamMember,
  createTeam,
  deleteTeam,
  getTeam,
  listTeamMembers,
  listTeams,
  removeTeamMember,
  updateTeam
} from '$lib/api/teams.js';

export type {
  AddTeamMemberRequest,
  CreateTeamRequest,
  Team,
  TeamListResponse,
  TeamMember,
  TeamMemberListResponse
} from '$lib/api/teams.js';

export const teamRepository = {
  list: listTeams,
  get: getTeam,
  create: createTeam,
  update: updateTeam,
  delete: deleteTeam,
  listMembers: listTeamMembers,
  addMember: addTeamMember,
  removeMember: removeTeamMember
};
