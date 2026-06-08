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
    folders: {
      title: 'Folders',
      subtitle: 'Group contracts, invoices, and IDs by client, project, or year.',
      allDocuments: 'All documents',
      newFolder: 'New folder',
      newSubfolder: 'New subfolder',
      rename: 'Rename',
      delete: 'Delete',
      confirmDelete: 'Delete this folder?',
      deleteBody: 'Documents inside this folder will move back to the root.',
      folderNamePlaceholder: 'Folder name',
      empty: 'No folders yet',
      emptyHint: 'Create a folder to group related documents together.',
      couldNotLoad: 'Could not load folders.',
      retry: 'Tap to retry',
      uncategorized: 'Uncategorized',
      documents: 'Documents',
      childFolders: 'Subfolders',
      noDocuments: 'No documents in this folder yet.',
      noChildren: 'No subfolders here yet.',
      moveTitle: 'Move document',
      moveHere: 'Move here',
      moveTo: 'Move to',
      root: 'Root',
      alreadyHere: 'Already in this folder',
      movedTo: 'Moved to %{name}',
      movedRoot: 'Moved to root',
      created: 'Folder created',
      renamed: 'Folder renamed',
      deleted: 'Folder deleted',
      createFailed: 'Could not create folder',
      renameFailed: 'Could not rename folder',
      deleteFailed: 'Could not delete folder',
      moveFailed: 'Could not move document',
      backToFolders: 'Folders',
      breadcrumbFolders: 'Folders',
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
    folders: {
      title: 'المجلدات',
      subtitle: 'نظّم العقود والفواتير والهويات حسب العميل أو المشروع أو السنة.',
      allDocuments: 'كل المستندات',
      newFolder: 'مجلد جديد',
      newSubfolder: 'مجلد فرعي',
      rename: 'إعادة تسمية',
      delete: 'حذف',
      confirmDelete: 'حذف هذا المجلد؟',
      deleteBody: 'ستعود المستندات داخل هذا المجلد إلى الجذر.',
      folderNamePlaceholder: 'اسم المجلد',
      empty: 'لا توجد مجلدات بعد',
      emptyHint: 'أنشئ مجلداً لتجميع المستندات المرتبطة ببعضها.',
      couldNotLoad: 'تعذّر تحميل المجلدات.',
      retry: 'اضغط للمحاولة مجدداً',
      uncategorized: 'غير مصنّف',
      documents: 'المستندات',
      childFolders: 'المجلدات الفرعية',
      noDocuments: 'لا توجد مستندات في هذا المجلد بعد.',
      noChildren: 'لا توجد مجلدات فرعية هنا بعد.',
      moveTitle: 'نقل المستند',
      moveHere: 'نقل إلى هنا',
      moveTo: 'نقل إلى',
      root: 'الجذر',
      alreadyHere: 'موجود في هذا المجلد بالفعل',
      movedTo: 'نُقل إلى %{name}',
      movedRoot: 'نُقل إلى الجذر',
      created: 'تم إنشاء المجلد',
      renamed: 'تمت إعادة تسمية المجلد',
      deleted: 'تم حذف المجلد',
      createFailed: 'تعذّر إنشاء المجلد',
      renameFailed: 'تعذّر إعادة تسمية المجلد',
      deleteFailed: 'تعذّر حذف المجلد',
      moveFailed: 'تعذّر نقل المستند',
      backToFolders: 'المجلدات',
      breadcrumbFolders: 'المجلدات',
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
