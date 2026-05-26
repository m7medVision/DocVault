import { useState } from 'react';
import { router } from 'expo-router';
import { Alert, StyleSheet, View } from 'react-native';
import { Button, Card, Input } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useAuth } from '@/lib/auth/auth-context';
import { useTranslation } from '@/lib/i18n';
import { updateProfile, updateEmail, updatePassword } from '@/features/profile/api';

export default function SettingsScreen() {
  const { user, updateUser, logout } = useAuth();
  const { t, isRTL } = useTranslation();

  const [displayName, setDisplayName] = useState(user?.displayName ?? '');
  const [locale, setLocale] = useState<'en' | 'ar'>(user?.locale ?? 'en');
  const [profileSubmitting, setProfileSubmitting] = useState(false);

  const [email, setEmail] = useState(user?.email ?? '');
  const [emailPassword, setEmailPassword] = useState('');
  const [emailSubmitting, setEmailSubmitting] = useState(false);

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [passwordSubmitting, setPasswordSubmitting] = useState(false);

  async function handleProfileUpdate() {
    if (!displayName.trim()) return;

    try {
      setProfileSubmitting(true);
      const response = await updateProfile({ display_name: displayName.trim(), locale });
      updateUser({ displayName: response.display_name, locale: response.locale });
      Alert.alert('', t('settings.profileUpdated'));
    } catch (err) {
      Alert.alert('', err instanceof Error ? err.message : t('common.error'));
    } finally {
      setProfileSubmitting(false);
    }
  }

  async function handleEmailUpdate() {
    if (!email.trim() || !emailPassword) return;

    try {
      setEmailSubmitting(true);
      const response = await updateEmail({ email: email.trim().toLowerCase(), current_password: emailPassword });
      updateUser({ email: response.email, emailVerified: response.email_verified });
      setEmail('');
      setEmailPassword('');
      Alert.alert('', t('settings.emailUpdated'));
    } catch (err) {
      Alert.alert('', err instanceof Error ? err.message : t('common.error'));
    } finally {
      setEmailSubmitting(false);
    }
  }

  async function handlePasswordUpdate() {
    if (!currentPassword || !newPassword) return;

    try {
      setPasswordSubmitting(true);
      await updatePassword({ current_password: currentPassword, new_password: newPassword });
      setCurrentPassword('');
      setNewPassword('');
      Alert.alert('', t('settings.passwordUpdated'));
    } catch (err) {
      Alert.alert('', err instanceof Error ? err.message : t('common.error'));
    } finally {
      setPasswordSubmitting(false);
    }
  }

  function handleLogout() {
    Alert.alert('', t('settings.logoutConfirm'), [
      { text: t('common.cancel'), style: 'cancel' },
      {
        text: t('settings.logout'),
        style: 'destructive',
        onPress: async () => {
          await logout();
          router.replace('/login');
        },
      },
    ]);
  }

  return (
    <DocVaultScreen>
      <View style={[styles.header, isRTL && styles.rtlRow]}>
        <ThemedText type="title">{t('settings.title')}</ThemedText>
      </View>

      <Card className="rounded-2xl border border-divider bg-content1 p-4">
        <ThemedText type="smallBold" style={styles.sectionTitle}>
          {t('settings.profileSection')}
        </ThemedText>
        <View style={styles.formGroup}>
          <ThemedText type="small" themeColor="textSecondary">
            {t('settings.displayName')}
          </ThemedText>
          <Input
            placeholder={t('settings.displayNamePlaceholder')}
            value={displayName}
            onChangeText={setDisplayName}
            editable={!profileSubmitting}
          />
        </View>
        <View style={styles.formGroup}>
          <ThemedText type="small" themeColor="textSecondary">
            {t('settings.locale')}
          </ThemedText>
          <View style={[styles.localeRow, isRTL && styles.rtlRow]}>
            <Button
              size="sm"
              variant={locale === 'en' ? 'primary' : 'secondary'}
              onPress={() => setLocale('en')}
              isDisabled={profileSubmitting}
            >
              <Button.Label>English</Button.Label>
            </Button>
            <Button
              size="sm"
              variant={locale === 'ar' ? 'primary' : 'secondary'}
              onPress={() => setLocale('ar')}
              isDisabled={profileSubmitting}
            >
              <Button.Label>العربية</Button.Label>
            </Button>
          </View>
        </View>
        <Button variant="primary" onPress={() => void handleProfileUpdate()} isDisabled={profileSubmitting}>
          <Button.Label>{profileSubmitting ? t('common.loading') : t('common.save')}</Button.Label>
        </Button>
      </Card>

      <Card className="rounded-2xl border border-divider bg-content1 p-4">
        <ThemedText type="smallBold" style={styles.sectionTitle}>
          {t('settings.emailSection')}
        </ThemedText>
        <View style={styles.formGroup}>
          <ThemedText type="small" themeColor="textSecondary">
            {t('settings.email')}
          </ThemedText>
          <Input
            placeholder={t('settings.emailPlaceholder')}
            value={email}
            onChangeText={setEmail}
            autoCapitalize="none"
            keyboardType="email-address"
            editable={!emailSubmitting}
          />
        </View>
        <View style={styles.formGroup}>
          <ThemedText type="small" themeColor="textSecondary">
            {t('settings.currentPassword')}
          </ThemedText>
          <Input
            placeholder="••••••••"
            value={emailPassword}
            onChangeText={setEmailPassword}
            secureTextEntry
            autoComplete="current-password"
            editable={!emailSubmitting}
          />
        </View>
        <Button variant="primary" onPress={() => void handleEmailUpdate()} isDisabled={emailSubmitting}>
          <Button.Label>{emailSubmitting ? t('common.loading') : t('common.save')}</Button.Label>
        </Button>
      </Card>

      <Card className="rounded-2xl border border-divider bg-content1 p-4">
        <ThemedText type="smallBold" style={styles.sectionTitle}>
          {t('settings.passwordSection')}
        </ThemedText>
        <View style={styles.formGroup}>
          <ThemedText type="small" themeColor="textSecondary">
            {t('settings.currentPassword')}
          </ThemedText>
          <Input
            placeholder="••••••••"
            value={currentPassword}
            onChangeText={setCurrentPassword}
            secureTextEntry
            autoComplete="current-password"
            editable={!passwordSubmitting}
          />
        </View>
        <View style={styles.formGroup}>
          <ThemedText type="small" themeColor="textSecondary">
            {t('settings.newPassword')}
          </ThemedText>
          <Input
            placeholder="••••••••"
            value={newPassword}
            onChangeText={setNewPassword}
            secureTextEntry
            autoComplete="new-password"
            editable={!passwordSubmitting}
          />
        </View>
        <Button variant="primary" onPress={() => void handlePasswordUpdate()} isDisabled={passwordSubmitting}>
          <Button.Label>{passwordSubmitting ? t('common.loading') : t('common.save')}</Button.Label>
        </Button>
      </Card>

      <Button variant="secondary" onPress={() => void handleLogout()}>
        <Button.Label>{t('settings.logout')}</Button.Label>
      </Button>
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: Spacing.two,
  },
  rtlRow: {
    flexDirection: 'row-reverse',
  },
  sectionTitle: {
    marginBottom: Spacing.two,
  },
  formGroup: {
    gap: Spacing.one,
    marginBottom: Spacing.two,
  },
  localeRow: {
    flexDirection: 'row',
    gap: Spacing.two,
  },
});