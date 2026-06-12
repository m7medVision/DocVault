import { apiFetch } from '@/lib/api/client';

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

export interface Member {
  membership_id: string;
  user_id: string;
  org_id: string;
  email: string;
  display_name: string;
  role: 'owner' | 'admin' | 'member' | 'viewer';
  created_at: string;
}

export interface MemberListResponse {
  members: Member[];
}

export interface PolicyRule {
  subject: string;
  tenant_id: string;
  object: string;
  action: string;
}

export interface PolicyListResponse {
  policies: PolicyRule[];
  bindings: { user_id: string; role: string; tenant_id: string }[];
}

export async function getAuditLog(cursor?: string): Promise<AuditLogResponse> {
  const params = new URLSearchParams();
  if (cursor) params.set('cursor', cursor);
  const search = params.toString();
  const response = await apiFetch<AuditLogResponse>(`/admin/audit${search ? `?${search}` : ''}`);
  return { events: response?.events ?? [], cursor: response?.cursor };
}

export async function listMembers(): Promise<MemberListResponse> {
  const response = await apiFetch<MemberListResponse>('/admin/members');
  return { members: response?.members ?? [] };
}

export async function updateMemberRole(
  membershipId: string,
  role: 'admin' | 'member' | 'viewer',
): Promise<void> {
  await apiFetch(`/admin/members/${membershipId}/role`, {
    method: 'PATCH',
    body: JSON.stringify({ role }),
  });
}

export async function listPolicies(): Promise<PolicyListResponse> {
  const response = await apiFetch<PolicyListResponse>('/admin/casbin/policies');
  return { policies: response?.policies ?? [], bindings: response?.bindings ?? [] };
}
