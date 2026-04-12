import {
  initializePushNotifications,
  requestNotificationPermissions,
  getPushToken,
  scheduleReminderNotification,
  cancelReminderNotification,
  addNotificationReceivedListener,
  addNotificationResponseListener,
} from '@/lib/notifications';
import type * as Notifications from 'expo-notifications';

export class NotificationService {
  async register(): Promise<{ hasPermission: boolean; pushToken: string | null }> {
    return initializePushNotifications();
  }

  async requestPermissions(): Promise<boolean> {
    return requestNotificationPermissions();
  }

  async getToken(): Promise<string | null> {
    return getPushToken();
  }

  async scheduleReminder(
    reminderId: string,
    documentTitle: string,
    dueDate: Date
  ): Promise<string | null> {
    return scheduleReminderNotification(reminderId, documentTitle, dueDate);
  }

  async cancelReminder(notificationId: string): Promise<void> {
    return cancelReminderNotification(notificationId);
  }

  onReceived(
    handler: (notification: Notifications.Notification) => void
  ): Notifications.EventSubscription {
    return addNotificationReceivedListener(handler);
  }

  onResponse(
    handler: (response: Notifications.NotificationResponse) => void
  ): Notifications.EventSubscription {
    return addNotificationResponseListener(handler);
  }
}

export const notificationService = new NotificationService();
