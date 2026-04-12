// Reminder API calls to the backend

import { CONFIG } from '../config';
import { authorizedFetch } from '../auth';
import { handleResponse } from './client';
import {
  Reminder,
  ReminderStatus,
  CreateReminderRequest,
  SnoozeReminderRequest,
  RegisterPushTokenRequest,
} from './types';

export {
  Reminder,
  ReminderStatus,
  CreateReminderRequest,
  SnoozeReminderRequest,
  RegisterPushTokenRequest,
};

// GET /api/v1/reminders - Get all reminders for current user
export async function getReminders(accessToken: string): Promise<Reminder[]> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders`, {
    method: 'GET',
  });
  return handleResponse<Reminder[]>(response);
}

// GET /api/v1/reminders/:id - Get single reminder
export async function getReminder(
  accessToken: string,
  reminderId: string
): Promise<Reminder> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders/${reminderId}`, {
    method: 'GET',
  });
  return handleResponse<Reminder>(response);
}

// POST /api/v1/reminders - Create a new reminder
export async function createReminder(
  accessToken: string,
  request: CreateReminderRequest
): Promise<Reminder> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
  return handleResponse<Reminder>(response);
}

// PATCH /api/v1/reminders/:id/snooze - Snooze a reminder
export async function snoozeReminder(
  accessToken: string,
  reminderId: string,
  request: SnoozeReminderRequest
): Promise<Reminder> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders/${reminderId}/snooze`, {
    method: 'PATCH',
    body: JSON.stringify(request),
  });
  return handleResponse<Reminder>(response);
}

// PATCH /api/v1/reminders/:id/dismiss - Dismiss a reminder
export async function dismissReminder(
  accessToken: string,
  reminderId: string
): Promise<Reminder> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders/${reminderId}/dismiss`, {
    method: 'PATCH',
  });
  return handleResponse<Reminder>(response);
}

// PATCH /api/v1/reminders/:id/complete - Mark reminder as completed
export async function completeReminder(
  accessToken: string,
  reminderId: string
): Promise<Reminder> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders/${reminderId}/complete`, {
    method: 'PATCH',
  });
  return handleResponse<Reminder>(response);
}

// DELETE /api/v1/reminders/:id - Delete a reminder
export async function deleteReminder(
  accessToken: string,
  reminderId: string
): Promise<{ message: string }> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders/${reminderId}`, {
    method: 'DELETE',
  });
  return handleResponse<{ message: string }>(response);
}

// POST /api/v1/reminders/push-token - Register push token with backend
export async function registerPushToken(
  accessToken: string,
  request: RegisterPushTokenRequest
): Promise<{ message: string }> {
  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/reminders/push-token`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
  return handleResponse<{ message: string }>(response);
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper functions for display
// ─────────────────────────────────────────────────────────────────────────────

export function formatDueDate(dueDate: string): string {
  const date = new Date(dueDate);
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays < 0) {
    const overdueDays = Math.abs(diffDays);
    if (overdueDays === 0) return 'Due today';
    if (overdueDays === 1) return '1 day overdue';
    return `${overdueDays} days overdue`;
  }

  if (diffDays === 0) {
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    if (diffHours <= 0) return 'Due now';
    if (diffHours === 1) return 'Due in 1 hour';
    return `Due in ${diffHours} hours`;
  }

  if (diffDays === 1) return 'Due tomorrow';
  if (diffDays < 7) return `Due in ${diffDays} days`;

  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
  });
}

export function getReminderStatusInfo(status: ReminderStatus): {
  label: string;
  color: string;
  bgColor: string;
} {
  switch (status) {
    case 'pending':
      return { label: 'Pending', color: '#f59e0b', bgColor: 'rgba(245, 158, 11, 0.1)' };
    case 'sent':
      return { label: 'Sent', color: '#3b82f6', bgColor: 'rgba(59, 130, 246, 0.1)' };
    case 'dismissed':
      return { label: 'Dismissed', color: '#6b7280', bgColor: 'rgba(107, 114, 128, 0.1)' };
    case 'completed':
      return { label: 'Completed', color: '#10b981', bgColor: 'rgba(16, 185, 129, 0.1)' };
    default:
      return { label: 'Unknown', color: '#6b7280', bgColor: 'rgba(107, 114, 128, 0.1)' };
  }
}

export const SNOOZE_OPTIONS = [
  { label: '15 minutes', value: 15 },
  { label: '30 minutes', value: 30 },
  { label: '1 hour', value: 60 },
  { label: '4 hours', value: 240 },
  { label: '1 day', value: 1440 },
  { label: '1 week', value: 10080 },
] as const;
