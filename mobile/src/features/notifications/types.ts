export interface Notification {
  id: string;
  tenant_id: string;
  user_id: string;
  type: string;
  title: string;
  body?: string;
  link?: string;
  status: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  read_at?: string;
}

export interface NotificationListResponse {
  notifications: Notification[];
  cursor?: string;
  total: number;
  unread_count: number;
}

export interface NotificationsOptions {
  status?: string;
  cursor?: string;
  limit?: number;
}
