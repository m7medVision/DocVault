import { apiFetch } from './core';
import type { Member } from './admin';

export interface PolicyRule {
  subject: string;
  tenant_id: string;
  object: string;
  action: string;
}

export interface RoleBinding {
  user_id: string;
  role: string;
  tenant_id: string;
}

export interface PolicyListResponse {
  policies: PolicyRule[];
  bindings: RoleBinding[];
}

export interface PolicyInput {
  subject: string;
  object: string;
  action: string;
}

export async function listPolicies(): Promise<PolicyListResponse> {
  const response = await apiFetch<PolicyListResponse>('/admin/casbin/policies');
  return {
    policies: response?.policies ?? [],
    bindings: response?.bindings ?? [],
  };
}

export async function addPolicy(input: PolicyInput): Promise<void> {
  await apiFetch('/admin/casbin/policies', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function deletePolicy(input: PolicyInput): Promise<void> {
  await apiFetch('/admin/casbin/policies', {
    method: 'DELETE',
    body: JSON.stringify(input),
  });
}

// The backend serializes authz.Permission without json tags, so keys are
// capitalized (Role/Object/Action). Match that shape exactly.
export interface MemberPermission {
  Role: string;
  Object: string;
  Action: string;
}

export interface MemberPermissionsResponse {
  member: Member;
  roles: string[];
  permissions: MemberPermission[];
  current_role: string;
}

export async function getMemberPermissions(
  membershipId: string
): Promise<MemberPermissionsResponse> {
  const response = await apiFetch<MemberPermissionsResponse>(
    `/admin/members/${membershipId}/permissions`
  );
  return {
    ...response,
    roles: response?.roles ?? [],
    permissions: response?.permissions ?? [],
  };
}
