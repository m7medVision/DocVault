import { initializePushNotifications } from '../notifications';

export class NotificationService {
  async register(): Promise<void> {
    await initializePushNotifications();
  }
}

export const notificationService = new NotificationService();
