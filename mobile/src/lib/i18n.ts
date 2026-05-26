import { useEffect, useState } from 'react';
import { I18n, type TranslateOptions } from 'i18n-js';
import { getLocales } from 'expo-localization';
import { useAuth } from '@/lib/auth/auth-context';

export type Locale = 'en' | 'ar';

const translations = {
  en: {
    home: {
      greeting: 'Welcome back, %{name}',
      subtitle: 'Scan, upload, and find every document from your phone.',
      totalDocuments: 'Total',
      pendingDocuments: 'Pending',
      completedThisWeek: 'This Week',
      storageUsed: 'Storage',
      recentDocuments: 'Recent Documents',
      upcomingReminders: 'Upcoming',
      viewAll: 'View all',
      scan: 'Scan',
      upload: 'Upload',
      noDocuments: 'No documents yet',
      noReminders: 'All clear',
    },
    status: {
      pending: 'Pending',
      processing: 'Processing',
      processed: 'Processed',
      failed: 'Failed',
    },
    common: {
      loading: 'Loading...',
      error: 'Something went wrong',
      retry: 'Retry',
      back: 'Back',
      cancel: 'Cancel',
      save: 'Save',
      delete: 'Delete',
    },
    notifications: {
      title: 'Notifications',
      unread: '%{count} new',
      allCaughtUp: 'All caught up',
      markAsRead: 'Mark as read',
    },
    reminders: {
      daysLeft: '%{days}d',
      dueToday: 'Today',
      overdue: '%{days}d overdue',
      dismiss: 'Dismiss',
      snooze: 'Snooze',
    },
    settings: {
      title: 'Settings',
      profileSection: 'Profile',
      emailSection: 'Email',
      passwordSection: 'Password',
      displayName: 'Display Name',
      displayNamePlaceholder: 'Your name',
      locale: 'Language',
      email: 'New Email',
      emailPlaceholder: 'your@email.com',
      currentPassword: 'Current Password',
      newPassword: 'New Password',
      profileUpdated: 'Profile updated',
      emailUpdated: 'Email updated',
      passwordUpdated: 'Password updated',
      logout: 'Log Out',
      logoutConfirm: 'Are you sure you want to log out?',
    },
  },
  ar: {
    home: {
      greeting: 'مرحباً بعودتك، %{name}',
      subtitle: 'امسح ضوئياً، ارفع، وابحث عن جميع مستنداتك من هاتفك.',
      totalDocuments: 'الإجمالي',
      pendingDocuments: 'معلق',
      completedThisWeek: 'هذا الأسبوع',
      storageUsed: 'المساحة',
      recentDocuments: 'أحدث المستندات',
      upcomingReminders: 'التنبيهات القادمة',
      viewAll: 'عرض الكل',
      scan: 'مسح ضوئي',
      upload: 'رفع ملف',
      noDocuments: 'لا توجد مستندات بعد',
      noReminders: 'لا توجد تنبيهات',
    },
    status: {
      pending: 'معلق',
      processing: 'قيد المعالجة',
      processed: 'مكتمل',
      failed: 'فشل',
    },
    common: {
      loading: 'جارٍ التحميل...',
      error: 'حدث خطأ ما',
      retry: 'إعادة المحاولة',
      back: 'رجوع',
      cancel: 'إلغاء',
      save: 'حفظ',
      delete: 'حذف',
    },
    notifications: {
      title: 'الإشعارات',
      unread: '%{count} جديد',
      allCaughtUp: 'لا توجد إشعارات',
      markAsRead: 'تحديد كمقروء',
    },
    reminders: {
      daysLeft: '%{days} يوم',
      dueToday: 'اليوم',
      overdue: 'متأخر بـ %{days} يوم',
      dismiss: 'تجاهل',
      snooze: 'تأجيل',
    },
    settings: {
      title: 'الإعدادات',
      profileSection: 'الملف الشخصي',
      emailSection: 'البريد الإلكتروني',
      passwordSection: 'كلمة المرور',
      displayName: 'اسم العرض',
      displayNamePlaceholder: 'اسمك',
      locale: 'اللغة',
      email: 'البريد الإلكتروني الجديد',
      emailPlaceholder: 'بريدك@الإلكتروني.com',
      currentPassword: 'كلمة المرور الحالية',
      newPassword: 'كلمة المرور الجديدة',
      profileUpdated: 'تم تحديث الملف',
      emailUpdated: 'تم تحديث البريد الإلكتروني',
      passwordUpdated: 'تم تحديث كلمة المرور',
      logout: 'تسجيل الخروج',
      logoutConfirm: 'هل أنت متأكد من تسجيل الخروج؟',
    },
  },
};

function detectLocale(): Locale {
  try {
    const locales = getLocales();
    const languageCode = locales[0]?.languageCode;
    if (languageCode === 'ar') return 'ar';
    return 'en';
  } catch {
    return 'en';
  }
}

const i18n = new I18n(translations);
i18n.enableFallback = true;
i18n.defaultLocale = 'en';
i18n.locale = detectLocale();

export function useTranslation() {
  const { user } = useAuth();
  const [locale, setLocale] = useState<Locale>(detectLocale);

  useEffect(() => {
    const preferredLocale = user?.locale;
    if (preferredLocale === 'ar' || preferredLocale === 'en') {
      setLocale(preferredLocale);
      i18n.locale = preferredLocale;
    }
  }, [user?.locale]);

  return {
    t: (scope: string, options?: TranslateOptions) => i18n.t(scope, options),
    locale,
    isRTL: locale === 'ar',
  };
}

export { i18n };
