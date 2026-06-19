import { apiFetch } from './core';

export type AclResourceType = 'document' | 'folder';
export type AclPrincipalType = 'user' | 'group';
export type AclPermission = 'read' | 'write' | 'delete';

// The backend serializes repository.Grant without json tags, so the list
// endpoint returns capitalized Go field names (ID/PrincipalType/...). Match
// that shape exactly, as lib/api/casbin.ts does for MemberPermission.
export interface Grant {
  ID: string;
  TenantID: string;
  OrgID: string;
  ResourceType: AclResourceType;
  ResourceID: string;
  PrincipalType: AclPrincipalType;
  PrincipalID: string;
  Permission: AclPermission;
  GrantedBy: string | null;
  CreatedAt: string;
}

export interface GrantListResponse {
  grants: Grant[];
}

export interface CreateGrantInput {
  resource_type: AclResourceType;
  resource_id: string;
  principal_type: AclPrincipalType;
  principal_id: string;
  permission?: AclPermission;
}

export interface CreateGrantResponse {
  id: string;
  resource_type: AclResourceType;
  resource_id: string;
  principal_type: AclPrincipalType;
  principal_id: string;
  permission: AclPermission;
}

export interface RestrictResponse {
  id: string;
  restricted: boolean;
}

// The backend serializes repository.Group without json tags, so groups come
// back with capitalized Go field names (ID/Name/...). Match that shape.
export interface Group {
  ID: string;
  TenantID: string;
  OrgID: string;
  Name: string;
  CreatedBy: string | null;
  CreatedAt: string;
}

export interface GroupListResponse {
  groups: Group[];
}

// --- Restrictions ---

export async function restrictDocument(id: string): Promise<RestrictResponse> {
  return apiFetch<RestrictResponse>(`/documents/${id}/restrict`, {
    method: 'PUT',
  });
}

export async function unrestrictDocument(
  id: string
): Promise<RestrictResponse> {
  return apiFetch<RestrictResponse>(`/documents/${id}/restrict`, {
    method: 'DELETE',
  });
}

export async function restrictFolder(id: string): Promise<RestrictResponse> {
  return apiFetch<RestrictResponse>(`/folders/${id}/restrict`, {
    method: 'PUT',
  });
}

export async function unrestrictFolder(id: string): Promise<RestrictResponse> {
  return apiFetch<RestrictResponse>(`/folders/${id}/restrict`, {
    method: 'DELETE',
  });
}

// --- Grants ---

export async function listGrants(
  resourceType: AclResourceType,
  resourceId: string
): Promise<GrantListResponse> {
  const params = new URLSearchParams({
    resource_type: resourceType,
    resource_id: resourceId,
  });
  const response = await apiFetch<GrantListResponse>(
    `/acl/grants?${params.toString()}`
  );
  return {
    ...response,
    grants: response?.grants ?? [],
  };
}

export async function createGrant(
  input: CreateGrantInput
): Promise<CreateGrantResponse> {
  return apiFetch<CreateGrantResponse>('/acl/grants', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function deleteGrant(id: string): Promise<void> {
  await apiFetch(`/acl/grants/${id}`, { method: 'DELETE' });
}

// --- Groups ---

export async function listGroups(): Promise<GroupListResponse> {
  const response = await apiFetch<GroupListResponse>('/groups');
  return {
    ...response,
    groups: response?.groups ?? [],
  };
}

export async function createGroup(name: string): Promise<{ group: Group }> {
  return apiFetch<{ group: Group }>('/groups', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export async function deleteGroup(id: string): Promise<void> {
  await apiFetch(`/groups/${id}`, { method: 'DELETE' });
}

export async function addGroupMember(
  groupId: string,
  userId: string
): Promise<void> {
  await apiFetch(`/groups/${groupId}/members`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId }),
  });
}

export async function removeGroupMember(
  groupId: string,
  userId: string
): Promise<void> {
  await apiFetch(`/groups/${groupId}/members/${userId}`, {
    method: 'DELETE',
  });
}
