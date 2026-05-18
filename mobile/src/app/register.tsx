import { useState } from 'react';
import { router } from 'expo-router';
import { StyleSheet, View, KeyboardAvoidingView, Platform } from 'react-native';
import { Button, Card, Input } from 'heroui-native';

import { useAuth } from '@/lib/auth/auth-context';
import type { RegisterParams } from '@/lib/auth/api';
import { Spacing } from '@/constants/theme';
import { ThemedText } from '@/components/themed-text';
import { useTheme } from '@/hooks/use-theme';

function validatePassword(pw: string): string | null {
  if (pw.length < 8) return 'At least 8 characters';
  if (!/[A-Z]/.test(pw)) return 'Needs an uppercase letter';
  if (!/[a-z]/.test(pw)) return 'Needs a lowercase letter';
  if (!/[0-9]/.test(pw)) return 'Needs a number';
  if (!/[^A-Za-z0-9]/.test(pw)) return 'Needs a special character';
  return null;
}

export default function RegisterScreen() {
  const { register } = useAuth();
  const theme = useTheme();
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit() {
    setError(null);

    if (!displayName.trim()) {
      setError('Full name is required');
      return;
    }

    if (!email.trim() || !email.includes('@')) {
      setError('Enter a valid email address');
      return;
    }

    const pwError = validatePassword(password);
    if (pwError) {
      setError(pwError);
      return;
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    try {
      setSubmitting(true);
      const params: RegisterParams = {
        email: email.trim().toLowerCase(),
        password,
        displayName: displayName.trim(),
        locale: 'en',
      };
      await register(params);
      router.replace('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <KeyboardAvoidingView
      style={[styles.root, { backgroundColor: theme.background }]}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <View style={styles.container}>
        <View style={styles.header}>
          <ThemedText type="subtitle">Create account</ThemedText>
          <ThemedText type="small" themeColor="textSecondary">
            We will set up your secure workspace right after registration.
          </ThemedText>
        </View>

        <Card className="rounded-3xl border border-divider bg-content1 p-5">
          <View style={styles.form}>
            <View style={styles.field}>
              <ThemedText type="smallBold">Full name</ThemedText>
              <Input
                placeholder="Your full name"
                value={displayName}
                onChangeText={setDisplayName}
                autoCapitalize="words"
                autoComplete="name"
                editable={!submitting}
              />
            </View>

            <View style={styles.field}>
              <ThemedText type="smallBold">Email</ThemedText>
              <Input
                placeholder="name@company.com"
                value={email}
                onChangeText={setEmail}
                autoCapitalize="none"
                keyboardType="email-address"
                autoComplete="email"
                editable={!submitting}
              />
            </View>

            <View style={styles.field}>
              <ThemedText type="smallBold">Password</ThemedText>
              <Input
                placeholder="••••••••"
                value={password}
                onChangeText={setPassword}
                secureTextEntry
                autoComplete="new-password"
                editable={!submitting}
              />
            </View>

            <View style={styles.field}>
              <ThemedText type="smallBold">Confirm password</ThemedText>
              <Input
                placeholder="••••••••"
                value={confirmPassword}
                onChangeText={setConfirmPassword}
                secureTextEntry
                autoComplete="new-password"
                editable={!submitting}
              />
            </View>

            {error && (
              <View style={styles.errorBox}>
                <ThemedText type="small" style={{ color: '#ef4444' }}>{error}</ThemedText>
              </View>
            )}

            <Button variant="primary" onPress={() => void handleSubmit()} isDisabled={submitting}>
              <Button.Label>{submitting ? 'Creating...' : 'Create account'}</Button.Label>
            </Button>
          </View>
        </Card>

        <View style={styles.switchRow}>
          <ThemedText type="small" themeColor="textSecondary">Already have an account?</ThemedText>
          <Button variant="secondary" size="sm" onPress={() => router.push('/login')}>
            <Button.Label>Sign in</Button.Label>
          </Button>
        </View>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1 },
  container: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: Spacing.four,
    maxWidth: 480,
    width: '100%',
    alignSelf: 'center',
    gap: Spacing.three,
  },
  header: {
    gap: Spacing.one,
  },
  form: {
    gap: Spacing.three,
  },
  field: {
    gap: Spacing.one,
  },
  errorBox: {
    borderRadius: Spacing.three,
    padding: Spacing.three,
    backgroundColor: 'rgba(239,68,68,0.1)',
  },
  switchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: Spacing.three,
  },
});
