import { api, getErrorMessage, getFieldErrors } from './client.js';
import type { UserRole } from './users.js';

export interface PublicRegistrationResponse {
  email: string;
  role: UserRole;
  role_display_name: string;
  expires_at: string;
}

export interface CompleteRegistrationRequest {
  first_name: string;
  last_name: string;
  password: string;
  privacy_ack: boolean;
}

export async function completeRegistration(
  token: string,
  req: CompleteRegistrationRequest
): Promise<void> {
  return api<void>(`/auth/registrations/${encodeURIComponent(token)}/complete`, {
    method: 'POST',
    body: JSON.stringify(req),
    skipHttpErrorNavigation: true
  });
}

export const getRegistrationErrorMessage = getErrorMessage;
export const getRegistrationFieldErrors = getFieldErrors;
