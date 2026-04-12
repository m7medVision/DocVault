// i18n configuration for mobile app
// Supports Arabic and English with RTL support, language persistence

import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import * as Localization from 'expo-localization';
import { I18nManager } from 'react-native';

// Translation resources for English and Arabic
const resources = {
  en: {
    translation: {
      // Auth
      signIn: 'Sign In',
      createAccount: 'Create Account',
      signingIn: 'Signing you in...',
      authenticationFailed: 'Authentication failed',
      signOut: 'Sign Out',
      signOutConfirm: 'Are you sure you want to sign out?',
      displayName: 'Full Name',
      displayNamePlaceholder: 'Mona Elshazly',
      displayNameRequired: 'Please enter your full name',
      email: 'Email',
      password: 'Password',
      confirmPassword: 'Confirm password',
      rememberMe: 'Remember me',
      rememberMeHint: 'Keep this device signed in for up to 30 days.',
      passwordHint: 'Use at least 8 characters with uppercase, lowercase, a number, and a symbol.',
      invalidEmail: 'Enter a valid email address',
      passwordLength: 'Password must be at least 8 characters long',
      passwordUppercase: 'Password must contain at least one uppercase letter',
      passwordLowercase: 'Password must contain at least one lowercase letter',
      passwordNumber: 'Password must contain at least one number',
      passwordSpecial: 'Password must contain at least one special character',
      passwordMismatch: 'Passwords do not match',
      needAccount: 'Need an account?',
      haveAccount: 'Already have an account?',
      platformLabel: 'Internal JWT Access',
      mobileLoginTitle: 'Pick up your documents right where you left them.',
      mobileLoginBody: 'Sign in to scan, search, and manage reminders with secure token refresh running behind the scenes.',
      mobileRegisterTitle: 'Create a secure workspace for every important record.',
      mobileRegisterBody: 'Your account sets up a tenant-aware workspace instantly so contracts, receipts, and IDs stay organized from day one.',
      spotlightSync: 'Sessions refresh before expiry while you use the app.',
      spotlightSearch: 'OCR search stays ready across Arabic and English documents.',
      spotlightShare: 'Tenant-scoped access keeps personal and team records separated.',

      // Navigation
      documents: 'Documents',
      search: 'Search',
      reminders: 'Reminders',
      settings: 'Settings',

      // Documents
      noDocuments: 'No documents yet',
      uploadFirst: 'Upload your first document to get started',
      searchPlaceholder: 'Search documents...',
      noDocumentsFound: 'No documents found',
      tryAdjustingFilters: 'Try adjusting your filters',
      uploadYourFirst: 'Upload your first document to get started',
      welcomeBack: 'Welcome back,',
      upload: 'Upload',
      camera: 'Camera',
      filters: 'Filters',
      filterDocuments: 'Filter Documents',
      clearAll: 'Clear All',
      applyFilters: 'Apply Filters',
      documentType: 'Document Type',
      folder: 'Folder',
      status: 'Status',
      allTypes: 'All Types',
      allFolders: 'All Folders',
      allStatuses: 'All Statuses',
      activeFilters: 'Active filters:',
      retry: 'Retry',

      // Document types
      contract: 'Contract',
      invoice: 'Invoice',
      warranty: 'Warranty',
      identity: 'Identity',
      receipt: 'Receipt',
      other: 'Other',

      // Document status
      pending: 'Pending',
      processing: 'Processing',
      processed: 'Processed',
      failed: 'Failed',

      // Folders
      personal: 'Personal',
      work: 'Work',
      financial: 'Financial',
      legal: 'Legal',

      // Search
      searchDocuments: 'Search Documents',
      searchPlaceholderAlt: 'Search by content, title, or metadata...',
      noResults: 'No results found',
      searching: 'Searching...',
      resultsCount: '{{count}} result found',
      resultsCount_other: '{{count}} results found',
      searchTime: 'Search completed in {{time}}ms',
      allLanguages: 'All Languages',
      dateRange: 'Date Range',
      startDate: 'Start Date',
      endDate: 'End Date',
      relevance: 'Relevance',
      date: 'Date',
      showFilters: 'Show Filters',
      hideFilters: 'Hide Filters',
      clearSearch: 'Clear',

      // Reminders
      noReminders: 'No reminders',
      noRemindersSubtext: 'You\'ll see reminders for your documents here',
      noOverdueReminders: 'No overdue reminders',
      noPendingReminders: 'No pending reminders',
      setRemindersForDocuments: 'Set reminders for important documents',
      allCaughtUp: 'All caught up!',
      all: 'All',
      overdue: 'Overdue',
      snooze: 'Snooze',
      dismiss: 'Dismiss',
      markComplete: 'Mark Complete',
      snoozeReminder: 'Snooze Reminder',
      selectDuration: 'Select duration',
      minutes: '{{count}} minute',
      minutes_other: '{{count}} minutes',
      hours: '{{count}} hour',
      hours_other: '{{count}} hours',
      tomorrow: 'Tomorrow',
      nextWeek: 'Next Week',
      dueToday: 'Due Today',
      dueTomorrow: 'Due Tomorrow',
      overdueReminder: 'Overdue',
      reminderCompleted: 'Completed',
      reminderSnoozed: 'Snoozed',
      reminderDismissed: 'Dismissed',

      // Settings
      account: 'Account',
      organization: 'Organization',
      language: 'Language',
      tenantId: 'Tenant ID',
      selectLanguage: 'Select Language',
      english: 'English',
      arabic: 'العربية / Arabic',

      // Upload
      uploadDocument: 'Upload Document',
      selectFile: 'Select File',
      takePhoto: 'Take Photo',
      chooseFromLibrary: 'Choose from Library',
      uploading: 'Uploading...',
      uploadSuccess: 'Document uploaded successfully',
      uploadFailed: 'Upload failed',
      documentTitle: 'Document Title',
      documentTypeLabel: 'Document Type',
      folderLabel: 'Folder',
      saveDocument: 'Save Document',

      // Camera
      documentScanner: 'Document Scanner',
      capture: 'Capture',
      retake: 'Retake',
      usePhoto: 'Use Photo',
      alignDocument: 'Align document within frame',

      // General
      cancel: 'Cancel',
      save: 'Save',
      delete: 'Delete',
      edit: 'Edit',
      close: 'Close',
      done: 'Done',
      next: 'Next',
      back: 'Back',
      loading: 'Loading...',
      error: 'Error',
      success: 'Success',
      organizationId: 'Organization ID',
    },
  },
  ar: {
    translation: {
      // Auth
      signIn: 'تسجيل الدخول',
      createAccount: 'إنشاء حساب',
      signingIn: 'جار تسجيل الدخول...',
      authenticationFailed: 'فشل المصادقة',
      signOut: 'تسجيل الخروج',
      signOutConfirm: 'هل أنت متأكد من رغبتك في تسجيل الخروج؟',
      displayName: 'الاسم الكامل',
      displayNamePlaceholder: 'منى الشاذلي',
      displayNameRequired: 'يرجى إدخال الاسم الكامل',
      email: 'البريد الإلكتروني',
      password: 'كلمة المرور',
      confirmPassword: 'تأكيد كلمة المرور',
      rememberMe: 'تذكرني',
      rememberMeHint: 'أبق هذا الجهاز مسجل الدخول لمدة تصل إلى 30 يوما.',
      passwordHint: 'استخدم 8 أحرف على الأقل مع حرف كبير وحرف صغير ورقم ورمز خاص.',
      invalidEmail: 'أدخل بريدا إلكترونيا صالحا',
      passwordLength: 'يجب أن تكون كلمة المرور 8 أحرف على الأقل',
      passwordUppercase: 'يجب أن تحتوي كلمة المرور على حرف كبير واحد على الأقل',
      passwordLowercase: 'يجب أن تحتوي كلمة المرور على حرف صغير واحد على الأقل',
      passwordNumber: 'يجب أن تحتوي كلمة المرور على رقم واحد على الأقل',
      passwordSpecial: 'يجب أن تحتوي كلمة المرور على رمز خاص واحد على الأقل',
      passwordMismatch: 'كلمتا المرور غير متطابقتين',
      needAccount: 'ليس لديك حساب؟',
      haveAccount: 'لديك حساب بالفعل؟',
      platformLabel: 'وصول داخلي عبر JWT',
      mobileLoginTitle: 'ارجع إلى وثائقك من حيث توقفت تماما.',
      mobileLoginBody: 'سجل الدخول لمسح الوثائق والبحث فيها وإدارة التذكيرات مع تجديد آمن للرموز في الخلفية.',
      mobileRegisterTitle: 'أنشئ مساحة آمنة لكل سجل مهم.',
      mobileRegisterBody: 'يؤسس حسابك مساحة عمل مرتبطة بالمستأجر فوريا حتى تبقى العقود والإيصالات والهويات منظمة من اليوم الأول.',
      spotlightSync: 'تتجدد الجلسات قبل انتهاء صلاحيتها أثناء استخدام التطبيق.',
      spotlightSearch: 'يبقى بحث OCR جاهزا عبر الوثائق العربية والإنجليزية.',
      spotlightShare: 'يحافظ الوصول المرتبط بالمستأجر على فصل السجلات الشخصية وسجلات الفريق.',

      // Navigation
      documents: 'المستندات',
      search: 'البحث',
      reminders: 'التذكيرات',
      settings: 'الإعدادات',

      // Documents
      noDocuments: 'لا توجد مستندات بعد',
      uploadFirst: 'قم بتحميل أول مستند للبدء',
      searchPlaceholder: 'البحث في المستندات...',
      noDocumentsFound: 'لم يتم العثور على مستندات',
      tryAdjustingFilters: 'حاول تعديل الفلاتر',
      uploadYourFirst: 'قم بتحميل أول مستند للبدء',
      welcomeBack: 'مرحباً بعودتك،',
      upload: 'تحميل',
      camera: 'الكاميرا',
      filters: 'الفلاتر',
      filterDocuments: 'تصفية المستندات',
      clearAll: 'مسح الكل',
      applyFilters: 'تطبيق الفلاتر',
      documentType: 'نوع المستند',
      folder: 'المجلد',
      status: 'الحالة',
      allTypes: 'جميع الأنواع',
      allFolders: 'جميع المجلدات',
      allStatuses: 'جميع الحالات',
      activeFilters: 'الفلاتر النشطة:',
      retry: 'إعادة المحاولة',

      // Document types
      contract: 'عقد',
      invoice: 'فاتورة',
      warranty: 'ضمان',
      identity: 'هوية',
      receipt: 'إيصال',
      other: 'أخرى',

      // Document status
      pending: 'قيد الانتظار',
      processing: 'قيد المعالجة',
      processed: 'تمت المعالجة',
      failed: 'فشل',

      // Folders
      personal: 'شخصي',
      work: 'عمل',
      financial: 'مالي',
      legal: 'قانوني',

      // Search
      searchDocuments: 'البحث في المستندات',
      searchPlaceholderAlt: 'البحث حسب المحتوى أو العنوان أو البيانات...',
      noResults: 'لم يتم العثور على نتائج',
      searching: 'جاري البحث...',
      resultsCount: 'تم العثور على {{count}} نتيجة',
      resultsCount_other: 'تم العثور على {{count}} نتائج',
      searchTime: 'البحث اكتمل في {{time}} مللي ثانية',
      allLanguages: 'جميع اللغات',
      dateRange: 'نطاق التاريخ',
      startDate: 'تاريخ البدء',
      endDate: 'تاريخ الانتهاء',
      relevance: 'الصلة',
      date: 'التاريخ',
      showFilters: 'إظهار الفلاتر',
      hideFilters: 'إخفاء الفلاتر',
      clearSearch: 'مسح',

      // Reminders
      noReminders: 'لا توجد تذكيرات',
      noRemindersSubtext: 'ستظهر تذكيرات المستندات هنا',
      all: 'الكل',
      overdue: 'متأخر',
      snooze: 'تأجيل',
      dismiss: 'رفض',
      markComplete: 'تحديد كمكتمل',
      snoozeReminder: 'تأجيل التذكير',
      selectDuration: 'اختر المدة',
      minutes: '{{count}} دقيقة',
      minutes_other: '{{count}} دقائق',
      hours: '{{count}} ساعة',
      hours_other: '{{count}} ساعات',
      tomorrow: 'غداً',
      nextWeek: 'الأسبوع القادم',
      dueToday: 'موعد اليوم',
      dueTomorrow: 'موعد غداً',
      overdueReminder: 'متأخر',
      reminderCompleted: 'مكتمل',
      reminderSnoozed: 'مؤجل',
      reminderDismissed: 'مرفوض',

      // Settings
      account: 'الحساب',
      organization: 'المنظمة',
      language: 'اللغة',
      tenantId: 'معرف المستأجر',
      selectLanguage: 'اختر اللغة',
      english: 'English / الإنجليزية',
      arabic: 'العربية / Arabic',

      // Upload
      uploadDocument: 'تحميل المستند',
      selectFile: 'اختر ملف',
      takePhoto: 'التقط صورة',
      chooseFromLibrary: 'اختر من المكتبة',
      uploading: 'جاري التحميل...',
      uploadSuccess: 'تم تحميل المستند بنجاح',
      uploadFailed: 'فشل التحميل',
      documentTitle: 'عنوان المستند',
      documentTypeLabel: 'نوع المستند',
      folderLabel: 'المجلد',
      saveDocument: 'حفظ المستند',

      // Camera
      documentScanner: 'ماسح المستندات',
      capture: 'التقاط',
      retake: 'إعادة التصوير',
      usePhoto: 'استخدام الصورة',
      alignDocument: 'محاذاة المستند داخل الإطار',

      // General
      cancel: 'إلغاء',
      save: 'حفظ',
      delete: 'حذف',
      edit: 'تعديل',
      close: 'إغلاق',
      done: 'تم',
      next: 'التالي',
      back: 'السابق',
      loading: 'جاري التحميل...',
      error: 'خطأ',
      success: 'نجاح',
      organizationId: 'معرف المنظمة',
    },
  },
};

// Initialize i18next
const initI18n = async () => {
  // Get device locale
  const deviceLocale = Localization.getLocales()[0]?.languageCode || 'en';

  i18n.use(initReactI18next).init({
    resources,
    lng: deviceLocale, // Default to device locale
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false,
    },
    compatibilityJSON: 'v3',
    react: {
      useSuspense: false,
    },
  });

  return i18n;
};

// Change language and update RTL
export const changeLanguage = async (lang: string) => {
  const isRTL = lang === 'ar';

  // Update i18next language
  await i18n.changeLanguage(lang);

  // Update RTL manager
  if (I18nManager.isRTL !== isRTL) {
    I18nManager.allowRTL(isRTL);
    I18nManager.forceRTL(isRTL);
  }
};

// Check if current language is RTL
export const isRTL = () => i18n.language === 'ar';

// Get current language
export const getCurrentLanguage = () => i18n.language;

export { initI18n };
export default i18n;
