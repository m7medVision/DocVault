// Push Notifications Service for DocVault Mobile
// Handles notification permissions, token management, and receiving notifications

import * as Notifications from 'expo-notifications';
import * as Device from 'expo-device';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';
import Constants from 'expo-constants';
import { registerPushToken } from './reminders';
import { TokenStorageKeys } from './types';

// Check if running in Expo Go (notifications require development build in SDK 53+)
const isExpoGo = Constants.executionEnvironment === 'storeClient';

// Configure notification handler for foreground notifications
if (!isExpoGo) {
  Notifications.setNotificationHandler({
    handleNotification: async () => ({
      shouldShowAlert: true,
      shouldPlaySound: true,
      shouldSetBadge: true,
      shouldShowBanner: true,
      shouldShowList: true,
    }),
  });
}

// Notification channel for Android
const ANDROID_CHANNEL_ID = 'docvault-reminders';

// Create notification channel for Android
export async function createNotificationChannel(): Promise<void> {
  if (isExpoGo || Platform.OS !== 'android') {
    return;
  }

  await Notifications.setNotificationChannelAsync(ANDROID_CHANNEL_ID, {
    name: 'Document Reminders',
    importance: Notifications.AndroidImportance.HIGH,
    vibrationPattern: [0, 250, 250, 250],
    lightColor: '#4f46e5',
    sound: 'default',
    enableVibrate: true,
    bypassDnd: false,
    lockscreenVisibility: Notifications.AndroidNotificationVisibility.PUBLIC,
  });
}

// Check and request notification permissions
export async function requestNotificationPermissions(): Promise<boolean> {
  if (isExpoGo) {
    console.log('Push notifications unavailable in Expo Go - use development build');
    return false;
  }

  if (!Device.isDevice) {
    console.log('Push notifications require a physical device');
    return false;
  }

  // Check existing permissions
  const { status: existingStatus } = await Notifications.getPermissionsAsync();
  
  if (existingStatus === 'granted') {
    return true;
  }

  // Request permissions
  const { status } = await Notifications.requestPermissionsAsync({
    ios: {
      allowAlert: true,
      allowBadge: true,
      allowSound: true,
    },
    android: {
      channelId: ANDROID_CHANNEL_ID,
    },
  });

  return status === 'granted';
}

// Get the Expo push token for this device
export async function getPushToken(): Promise<string | null> {
  if (isExpoGo || !Device.isDevice) {
    return null;
  }

  try {
    // Ensure channel exists for Android
    await createNotificationChannel();

    const { data: token } = await Notifications.getExpoPushTokenAsync({
      projectId: 'docvault-mobile', // From app.json
    });

    return token;
  } catch (error) {
    console.error('Failed to get push token:', error);
    return null;
  }
}

// Get stored access token
async function getAccessToken(): Promise<string | null> {
  try {
    return await SecureStore.getItemAsync(TokenStorageKeys.ACCESS_TOKEN);
  } catch {
    return null;
  }
}

// Register push token with the backend
export async function registerPushTokenWithBackend(
  pushToken: string
): Promise<boolean> {
  const accessToken = await getAccessToken();
  
  if (!accessToken) {
    console.log('Not authenticated, skipping push token registration');
    return false;
  }

  try {
    const platform: 'ios' | 'android' = Platform.OS === 'ios' ? 'ios' : 'android';
    const deviceId = `${Platform.OS}-${Date.now()}`;

    await registerPushToken(accessToken, {
      push_token: pushToken,
      platform,
      device_id: deviceId,
    });

    console.log('Push token registered with backend');
    return true;
  } catch (error) {
    console.error('Failed to register push token with backend:', error);
    return false;
  }
}

// Initialize push notifications (call on app start)
export async function initializePushNotifications(): Promise<{
  hasPermission: boolean;
  pushToken: string | null;
}> {
  if (isExpoGo) {
    return { hasPermission: false, pushToken: null };
  }

  const hasPermission = await requestNotificationPermissions();
  
  if (!hasPermission) {
    console.log('Notification permission not granted');
    return { hasPermission: false, pushToken: null };
  }

  const pushToken = await getPushToken();
  
  if (pushToken) {
    await registerPushTokenWithBackend(pushToken);
  }

  return { hasPermission, pushToken };
}

// Schedule a local notification for a reminder
export async function scheduleReminderNotification(
  _reminderId: string,
  _documentTitle: string,
  _dueDate: Date
): Promise<string | null> {
  if (isExpoGo) {
    return null;
  }

  try {
    const notificationId = await Notifications.scheduleNotificationAsync({
      content: {
        title: '📋 Document Reminder',
        body: `"${_documentTitle}" is due ${formatDueDateText(_dueDate)}`,
        data: {
          type: 'reminder',
          reminderId: _reminderId,
          documentTitle: _documentTitle,
        },
        sound: 'default',
      },
      trigger: {
        type: Notifications.SchedulableTriggerInputTypes.DATE,
        date: _dueDate,
      },
    });

    return notificationId;
  } catch (error) {
    console.error('Failed to schedule notification:', error);
    return null;
  }
}

// Cancel a scheduled notification
export async function cancelReminderNotification(
  notificationId: string
): Promise<void> {
  if (isExpoGo) {
    return;
  }

  try {
    await Notifications.cancelScheduledNotificationAsync(notificationId);
  } catch (error) {
    console.error('Failed to cancel notification:', error);
  }
}

// Cancel all scheduled notifications for a reminder
export async function cancelAllReminderNotifications(
  notificationIds: string[]
): Promise<void> {
  if (isExpoGo) {
    return;
  }

  await Promise.all(
    notificationIds.map(id => cancelReminderNotification(id))
  );
}

// Format due date text for notification
function formatDueDateText(date: Date): string {
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  const diffMins = Math.floor(diffMs / (1000 * 60));

  if (diffMins < 0) return 'now';
  if (diffMins < 60) return `in ${diffMins} minutes`;
  
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `in ${diffHours} hour${diffHours > 1 ? 's' : ''}`;
  
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays === 1) return 'tomorrow';
  if (diffDays < 7) return `in ${diffDays} days`;
  
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

// Dummy subscription for Expo Go
const dummySubscription: Notifications.EventSubscription = {
  remove: () => {},
};

// Add notification received listener
export function addNotificationReceivedListener(
  handler: (notification: Notifications.Notification) => void
): Notifications.EventSubscription {
  if (isExpoGo) {
    return dummySubscription;
  }
  return Notifications.addNotificationReceivedListener(handler);
}

// Add notification response listener (when user taps notification)
export function addNotificationResponseListener(
  handler: (response: Notifications.NotificationResponse) => void
): Notifications.EventSubscription {
  if (isExpoGo) {
    return dummySubscription;
  }
  return Notifications.addNotificationResponseReceivedListener(handler);
}

// Get all scheduled notifications
export async function getScheduledNotifications(): Promise<
  Notifications.NotificationRequest[]
> {
  if (isExpoGo) {
    return [];
  }
  return Notifications.getAllScheduledNotificationsAsync();
}

// Badge management
export async function setBadgeCount(count: number): Promise<void> {
  if (isExpoGo || Platform.OS !== 'ios') {
    return;
  }
  await Notifications.setBadgeCountAsync(count);
}

export async function getBadgeCount(): Promise<number> {
  if (isExpoGo || Platform.OS !== 'ios') {
    return 0;
  }
  return Notifications.getBadgeCountAsync();
}
