export const NOTIFICATION_RULE_RECIPIENT_TYPES = [
  'users',
  'team',
  'project_users',
  'project_role'
] as const;

export type NotificationRuleRecipientType = (typeof NOTIFICATION_RULE_RECIPIENT_TYPES)[number];

export interface NotificationRule {
  id: string;
  name: string;
  enabled: boolean;
  event_key: string;
  project_id?: string | null;
  resource_type: string;
  resource_id?: string | null;
  recipient_type: NotificationRuleRecipientType;
  recipient_user_ids?: string[];
  recipient_team_id?: string | null;
  recipient_role: string;
  created_by_id?: string | null;
  created_at: string;
  updated_at: string;
}

export interface NotificationRuleList {
  items: NotificationRule[];
}

export interface UpsertNotificationRuleRequest {
  name: string;
  enabled: boolean;
  event_key: string;
  project_id?: string | null;
  resource_type?: string;
  resource_id?: string | null;
  recipient_type: NotificationRuleRecipientType;
  recipient_user_ids?: string[];
  recipient_team_id?: string | null;
  recipient_role?: string;
}
