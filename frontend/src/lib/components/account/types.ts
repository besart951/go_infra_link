export type AccountTab = 'information' | 'notifications' | 'password' | 'preferences';

export interface AccountTabItem {
  value: AccountTab;
  label: string;
}
