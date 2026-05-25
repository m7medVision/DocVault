import { useState } from 'react';
import { router } from 'expo-router';
import { StyleSheet, View, KeyboardAvoidingView, Platform } from 'react-native';
import { Button, Card, Input , useThemeColor } from 'heroui-native';

import { useAuth } from '@/lib/auth/auth-context';
import { Spacing } from '@/constants/theme';
import { ThemedText } from '@/components/themed-text';

export default function LoginScreen() {
  const { login } = useAuth();
  const [background, danger] = useThemeColor(['background', 'danger']);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit() {
    setError(null);

    if (!email.trim() || !email.includes('@')) {
      setError('Enter a valid email address');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    try {
      setSubmitting(true);
      await login(email, password);
      router.replace('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <KeyboardAvoidingView
      style={[styles.root, { backgroundColor: background }]}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <View style={styles.container}>
        <View style={styles.header}>
          <ThemedText type="subtitle">Sign in</ThemedText>
          <ThemedText type="small" themeColor="muted">
            Use your DocVault credentials to continue.
          </ThemedText>
        </View>

        <Card className="rounded-3xl border border-divider bg-content1 p-5">
          <View style={styles.form}>
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
                autoComplete="current-password"
                editable={!submitting}
              />
              <ThemedText type="small" themeColor="muted">
                At least 8 characters with uppercase, lowercase, a number, and a symbol.
              </ThemedText>
            </View>

            {error && (
              <View style={[styles.errorBox, { backgroundColor: `${danger}15` }]}>
                <ThemedText type="small" themeColor="danger">{error}</ThemedText>
              </View>
            )}

            <Button variant="primary" onPress={() => void handleSubmit()} isDisabled={submitting}>
              <Button.Label>{submitting ? 'Signing in...' : 'Sign in'}</Button.Label>
            </Button>
          </View>
        </Card>

        <View style={styles.switchRow}>
          <ThemedText type="small" themeColor="muted">Need an account?</ThemedText>
          <Button variant="secondary" size="sm" onPress={() => router.push('/register')}>
            <Button.Label>Create account</Button.Label>
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
  },
  switchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: Spacing.three,
  },
});
