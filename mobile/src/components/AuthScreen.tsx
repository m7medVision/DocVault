import { useEffect, useMemo, useState } from 'react';
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Switch,
  Text,
  TextInput,
  View,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../contexts/AuthContext';
import { getCurrentLanguage } from '../i18n';
import { tokens } from '../theme/tokens';

type Mode = 'login' | 'register';

interface AuthScreenProps {
  mode: Mode;
}

function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validatePassword(password: string): string | null {
  if (password.length < 8) return 'passwordLength';
  if (!/[A-Z]/.test(password)) return 'passwordUppercase';
  if (!/[a-z]/.test(password)) return 'passwordLowercase';
  if (!/[0-9]/.test(password)) return 'passwordNumber';
  if (!/[^A-Za-z0-9]/.test(password)) return 'passwordSpecial';
  return null;
}

export default function AuthScreen({ mode }: AuthScreenProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { isAuthenticated, isLoading, login, register } = useAuth();
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace('/(app)/documents');
    }
  }, [isAuthenticated, isLoading, router]);

  const headline = useMemo(
    () => (mode === 'login' ? t('mobileLoginTitle') : t('mobileRegisterTitle')),
    [mode, t]
  );

  const subhead = useMemo(
    () => (mode === 'login' ? t('mobileLoginBody') : t('mobileRegisterBody')),
    [mode, t]
  );

  const handleSubmit = async () => {
    setError(null);

    if (!isValidEmail(email.trim())) {
      setError(t('invalidEmail'));
      return;
    }

    const passwordErrorKey = validatePassword(password);
    if (passwordErrorKey) {
      setError(t(passwordErrorKey));
      return;
    }

    if (mode === 'register') {
      if (!displayName.trim()) {
        setError(t('displayNameRequired'));
        return;
      }
      if (password !== confirmPassword) {
        setError(t('passwordMismatch'));
        return;
      }
    }

    try {
      setSubmitting(true);

      if (mode === 'login') {
        await login({ email, password, rememberMe });
      } else {
        await register({
          displayName,
          email,
          password,
          rememberMe,
          locale: getCurrentLanguage() === 'ar' ? 'ar' : 'en',
        });
      }
    } catch (submitError) {
      setError(
        submitError instanceof Error ? submitError.message : t('authenticationFailed')
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      style={styles.keyboardArea}
    >
      <ScrollView contentContainerStyle={styles.scrollContent} keyboardShouldPersistTaps="handled">
        <View style={styles.heroPanel}>
          <Text style={styles.kicker}>{t('platformLabel')}</Text>
          <Text style={styles.logo}>DocVault</Text>
          <Text style={styles.headline}>{headline}</Text>
          <Text style={styles.subhead}>{subhead}</Text>

          <View style={styles.bulletList}>
            <Text style={styles.bulletItem}>{t('spotlightSync')}</Text>
            <Text style={styles.bulletItem}>{t('spotlightSearch')}</Text>
            <Text style={styles.bulletItem}>{t('spotlightShare')}</Text>
          </View>
        </View>

        <View style={styles.formCard}>
          {mode === 'register' ? (
            <View style={styles.fieldGroup}>
              <Text style={styles.label}>{t('displayName')}</Text>
              <TextInput
                autoCapitalize="words"
                autoCorrect={false}
                editable={!submitting}
                onChangeText={setDisplayName}
                placeholder={t('displayNamePlaceholder')}
                placeholderTextColor={tokens.textColor.muted}
                style={styles.input}
                value={displayName}
              />
            </View>
          ) : null}

          <View style={styles.fieldGroup}>
            <Text style={styles.label}>{t('email')}</Text>
            <TextInput
              autoCapitalize="none"
              autoCorrect={false}
              editable={!submitting}
              keyboardType="email-address"
              onChangeText={setEmail}
              placeholder="name@company.com"
              placeholderTextColor={tokens.textColor.muted}
              style={styles.input}
              textContentType="emailAddress"
              value={email}
            />
          </View>

          <View style={styles.fieldGroup}>
            <Text style={styles.label}>{t('password')}</Text>
            <TextInput
              autoCapitalize="none"
              editable={!submitting}
              onChangeText={setPassword}
              placeholder="••••••••"
              placeholderTextColor={tokens.textColor.muted}
              secureTextEntry
              style={styles.input}
              textContentType={mode === 'login' ? 'password' : 'newPassword'}
              value={password}
            />
          </View>

          {mode === 'register' ? (
            <View style={styles.fieldGroup}>
              <Text style={styles.label}>{t('confirmPassword')}</Text>
              <TextInput
                autoCapitalize="none"
                editable={!submitting}
                onChangeText={setConfirmPassword}
                placeholder="••••••••"
                placeholderTextColor={tokens.textColor.muted}
                secureTextEntry
                style={styles.input}
                textContentType="newPassword"
                value={confirmPassword}
              />
            </View>
          ) : null}

          <View style={styles.switchRow}>
            <View style={styles.switchTextBlock}>
              <Text style={styles.label}>{t('rememberMe')}</Text>
              <Text style={styles.helper}>{t('rememberMeHint')}</Text>
            </View>
            <Switch
              onValueChange={setRememberMe}
              thumbColor="#ffffff"
              trackColor={{ false: '#334155', true: tokens.accentColor }}
              value={rememberMe}
            />
          </View>

          <Text style={styles.helper}>{t('passwordHint')}</Text>
          {error ? <Text style={styles.error}>{error}</Text> : null}

          <Pressable
            disabled={submitting}
            onPress={handleSubmit}
            style={({ pressed }) => [
              styles.primaryButton,
              pressed && !submitting ? styles.primaryButtonPressed : null,
              submitting ? styles.primaryButtonDisabled : null,
            ]}
          >
            {submitting ? (
              <ActivityIndicator color={tokens.textColor.primary} />
            ) : (
              <Text style={styles.primaryButtonText}>
                {mode === 'login' ? t('signIn') : t('createAccount')}
              </Text>
            )}
          </Pressable>

          <Pressable
            onPress={() =>
              router.replace(mode === 'login' ? '/register' : '/login')
            }
            style={styles.secondaryAction}
          >
            <Text style={styles.secondaryText}>
              {mode === 'login' ? t('needAccount') : t('haveAccount')}{' '}
              <Text style={styles.secondaryLink}>
                {mode === 'login' ? t('createAccount') : t('signIn')}
              </Text>
            </Text>
          </Pressable>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  keyboardArea: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.primary,
  },
  scrollContent: {
    flexGrow: 1,
    padding: tokens.screenPadding,
    justifyContent: 'center',
    gap: 20,
  },
  heroPanel: {
    backgroundColor: tokens.backgroundColor.tertiary,
    borderRadius: 24,
    padding: 24,
    gap: 12,
  },
  kicker: {
    alignSelf: 'flex-start',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 999,
    overflow: 'hidden',
    backgroundColor: 'rgba(79, 70, 229, 0.18)',
    color: tokens.accentColor,
    fontSize: 12,
    fontWeight: '700',
    letterSpacing: 0.6,
    textTransform: 'uppercase',
  },
  logo: {
    fontSize: 36,
    fontWeight: '800',
    color: tokens.textColor.primary,
  },
  headline: {
    fontSize: 28,
    fontWeight: '700',
    color: tokens.textColor.primary,
    lineHeight: 34,
  },
  subhead: {
    fontSize: 15,
    lineHeight: 24,
    color: tokens.textColor.secondary,
  },
  bulletList: {
    gap: 10,
    marginTop: 8,
  },
  bulletItem: {
    paddingVertical: 10,
    paddingHorizontal: 14,
    borderRadius: 14,
    backgroundColor: 'rgba(255, 255, 255, 0.06)',
    color: tokens.textColor.primary,
    fontSize: 14,
  },
  formCard: {
    backgroundColor: tokens.backgroundColor.secondary,
    borderRadius: 24,
    padding: 20,
    gap: 14,
  },
  fieldGroup: {
    gap: 8,
  },
  label: {
    color: tokens.textColor.primary,
    fontSize: 14,
    fontWeight: '600',
  },
  input: {
    borderWidth: 1,
    borderColor: '#24324f',
    borderRadius: 14,
    backgroundColor: '#111827',
    color: tokens.textColor.primary,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 15,
  },
  helper: {
    color: tokens.textColor.secondary,
    fontSize: 13,
    lineHeight: 18,
  },
  switchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 12,
    paddingVertical: 6,
  },
  switchTextBlock: {
    flex: 1,
    gap: 4,
  },
  error: {
    paddingHorizontal: 14,
    paddingVertical: 12,
    borderRadius: 14,
    backgroundColor: 'rgba(239, 68, 68, 0.18)',
    color: tokens.errorColor,
    fontSize: 14,
  },
  primaryButton: {
    minHeight: 54,
    borderRadius: 16,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: tokens.accentColor,
  },
  primaryButtonPressed: {
    opacity: 0.92,
  },
  primaryButtonDisabled: {
    opacity: 0.7,
  },
  primaryButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '700',
  },
  secondaryAction: {
    paddingVertical: 8,
    alignItems: 'center',
  },
  secondaryText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
  },
  secondaryLink: {
    color: tokens.accentColor,
    fontWeight: '700',
  },
});
