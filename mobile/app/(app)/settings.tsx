import { View, Text, TouchableOpacity, StyleSheet, Alert, Modal, Pressable } from 'react-native';
import { useTranslation } from 'react-i18next';
import { useSettings } from '../../src/features/settings/useSettings';
import { tokens, colors } from '../../src/theme/tokens';

const LANGUAGES = [
  { code: 'en', label: 'English', nativeLabel: 'English' },
  { code: 'ar', label: 'Arabic', nativeLabel: 'العربية' },
];

export default function SettingsScreen() {
  const { t } = useTranslation();
  const {
    user,
    settings,
    languageModalVisible,
    setLanguageModalVisible,
    changeLanguageSetting,
    handleLogout,
  } = useSettings();

  const getCurrentLanguageLabel = () => {
    const lang = LANGUAGES.find((l) => l.code === settings.language);
    return lang?.nativeLabel || 'English';
  };

  const onLanguageChange = (langCode: string) => {
    changeLanguageSetting(langCode);
    setLanguageModalVisible(false);
  };

  const onLogout = () => {
    Alert.alert(t('signOut'), t('signOutConfirm'), [
      { text: t('cancel'), style: 'cancel' },
      {
        text: t('signOut'),
        style: 'destructive',
        onPress: handleLogout,
      },
    ]);
  };

  return (
    <View style={styles.container}>
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>{t('account')}</Text>
        <View style={styles.card}>
          <View style={styles.userRow}>
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>
                {user?.name?.charAt(0).toUpperCase() || 'U'}
              </Text>
            </View>
            <View style={styles.userInfo}>
              <Text style={styles.userName}>{user?.name || 'User'}</Text>
              <Text style={styles.userEmail}>{user?.email || ''}</Text>
            </View>
          </View>
        </View>
      </View>

      {user?.tenantId && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t('organization')}</Text>
          <View style={styles.card}>
            <Text style={styles.tenantLabel}>{t('tenantId')}</Text>
            <Text style={styles.tenantValue}>{user.tenantId}</Text>
          </View>
        </View>
      )}

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>{t('language')}</Text>
        <View style={styles.card}>
          <TouchableOpacity
            style={styles.row}
            onPress={() => setLanguageModalVisible(true)}
          >
            <Text style={styles.rowLabel}>{t('language')}</Text>
            <View style={styles.languageValue}>
              <Text style={styles.rowValue}>{getCurrentLanguageLabel()}</Text>
              <Text style={styles.chevron}>›</Text>
            </View>
          </TouchableOpacity>
        </View>
      </View>

      <View style={styles.section}>
        <TouchableOpacity style={styles.logoutButton} onPress={onLogout}>
          <Text style={styles.logoutText}>{t('signOut')}</Text>
        </TouchableOpacity>
      </View>

      <Modal
        animationType="slide"
        transparent={true}
        visible={languageModalVisible}
        onRequestClose={() => setLanguageModalVisible(false)}
      >
        <Pressable
          style={styles.modalOverlay}
          onPress={() => setLanguageModalVisible(false)}
        >
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>{t('selectLanguage')}</Text>
              <TouchableOpacity onPress={() => setLanguageModalVisible(false)}>
                <Text style={styles.modalClose}>✕</Text>
              </TouchableOpacity>
            </View>

            <View style={styles.languageList}>
              {LANGUAGES.map((lang) => (
                <TouchableOpacity
                  key={lang.code}
                  style={[
                    styles.languageOption,
                    settings.language === lang.code && styles.languageOptionActive,
                  ]}
                  onPress={() => onLanguageChange(lang.code)}
                >
                  <View style={styles.languageInfo}>
                    <Text
                      style={[
                        styles.languageLabel,
                        settings.language === lang.code && styles.languageLabelActive,
                      ]}
                    >
                      {lang.nativeLabel}
                    </Text>
                    <Text style={styles.languageSubLabel}>{lang.label}</Text>
                  </View>
                  {settings.language === lang.code && (
                    <Text style={styles.checkmark}>✓</Text>
                  )}
                </TouchableOpacity>
              ))}
            </View>
          </View>
        </Pressable>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.secondary,
    padding: 16,
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    color: tokens.textColor.muted,
    fontSize: 13,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: 8,
    marginLeft: 4,
  },
  card: {
    backgroundColor: colors.gray[800],
    borderRadius: tokens.cardRadius,
    overflow: 'hidden',
  },
  userRow: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
  },
  avatar: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: tokens.accentColor,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  avatarText: {
    color: tokens.textColor.primary,
    fontSize: 20,
    fontWeight: 'bold',
  },
  userInfo: {
    flex: 1,
  },
  userName: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  userEmail: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    marginTop: 2,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
  },
  languageValue: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  rowLabel: {
    color: tokens.textColor.primary,
    fontSize: 16,
  },
  rowValue: {
    color: tokens.textColor.secondary,
    fontSize: 16,
  },
  chevron: {
    color: tokens.textColor.muted,
    fontSize: 20,
  },
  tenantLabel: {
    color: tokens.textColor.secondary,
    fontSize: 12,
    padding: 16,
    paddingBottom: 4,
  },
  tenantValue: {
    color: tokens.textColor.primary,
    fontSize: 14,
    fontFamily: 'monospace',
    paddingHorizontal: 16,
    paddingBottom: 16,
  },
  logoutButton: {
    backgroundColor: `${tokens.errorColor}1A`,
    borderRadius: tokens.cardRadius,
    padding: 16,
    alignItems: 'center',
  },
  logoutText: {
    color: tokens.errorColor,
    fontSize: 16,
    fontWeight: '600',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: colors.gray[900],
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: colors.gray[800],
  },
  modalTitle: {
    color: tokens.textColor.primary,
    fontSize: 20,
    fontWeight: 'bold',
  },
  modalClose: {
    color: tokens.textColor.secondary,
    fontSize: 24,
    padding: 4,
  },
  languageList: {
    padding: 8,
  },
  languageOption: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderRadius: 12,
    marginVertical: 4,
  },
  languageOptionActive: {
    backgroundColor: colors.gray[800],
  },
  languageInfo: {
    gap: 4,
  },
  languageLabel: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  languageLabelActive: {
    color: tokens.accentColor,
  },
  languageSubLabel: {
    color: tokens.textColor.secondary,
    fontSize: 13,
  },
  checkmark: {
    color: tokens.accentColor,
    fontSize: 18,
    fontWeight: 'bold',
  },
});
