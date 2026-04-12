import { apiFetch } from './core';

export interface AuditEvent {
  id: string;
  tenant_id: string;
  actor_id?: string;
  entity_type: string;
  entity_id: string;
  action: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AuditLogResponse {
  events: AuditEvent[];
  cursor?: string;
}

export interface AuditLogOptions {
  entity_type?: string;
  action?: string;
}

export interface Member {
  membership_id: string;
  user_id: string;
  org_id: string;
  email: string;
  display_name: string;
  role: 'admin' | 'member' | 'viewer' | 'owner';
  created_at: string;
}

export interface MemberListResponse {
  members: Member[];
}

export async function getAuditLog(
  options: AuditLogOptions = {}
): Promise<AuditLogResponse> {
  const params = new URLSearchParams();
  if (options.entity_type && options.entity_type !== 'all') {
    params.set('entity_type', options.entity_type);
  }
  if (options.action && options.action !== 'all') {
    params.set('action', options.action);
  }

  const query = params.toString();
  const response = await apiFetch<AuditLogResponse>(
    `/admin/audit${query ? `?${query}` : ''}`
  );

  return {
    ...response,
    events: response?.events ?? [],
  };
}

export async function listMembers(): Promise<MemberListResponse> {
  const response = await apiFetch<MemberListResponse>('/admin/members');

  return {
    ...response,
    members: response?.members ?? [],
  };
}

export async function inviteMember(
  email: string,
  role: 'admin' | 'member' | 'viewer'
): Promise<void> {
  await apiFetch('/admin/members/invite', {
    method: 'POST',
    body: JSON.stringify({ email, role }),
  });
}

export async function updateMemberRole(
  membershipId: string,
  role: 'admin' | 'member' | 'viewer'
): Promise<void> {
  await apiFetch(`/admin/members/${membershipId}/role`, {
    method: 'PATCH',
    body: JSON.stringify({ role }),
  });
}