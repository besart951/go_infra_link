export type DashboardProjectStatus = 'planned' | 'ongoing' | 'completed';

export interface DashboardProject {
  id: string;
  name: string;
  status: DashboardProjectStatus;
  phase: string;
  updated_at: string;
}

export interface DashboardUserPresence {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  last_login_at?: string;
  is_online: boolean;
}

export interface DashboardTeamMember {
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: string;
  last_login_at?: string;
  is_online: boolean;
}

export interface DashboardTeam {
  id: string;
  name: string;
  role: string;
  members: DashboardTeamMember[];
}

export interface DashboardTeamSummary {
  id: string;
  name: string;
  role: string;
  joined_at: string;
}

export interface DashboardSnapshot {
  last_project?: DashboardProject;
  primary_team?: DashboardTeam;
  teams: DashboardTeamSummary[];
  online_users: DashboardUserPresence[];
}

