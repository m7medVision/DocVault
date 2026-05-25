import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';
import { Button, Card } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { HeroActionCard } from '@/components/hero-action-card';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useAuth } from '@/lib/auth/auth-context';

export default function HomeScreen() {
  const { user, logout } = useAuth();

  return (
    <DocVaultScreen>
      <Card className="rounded-[32px] bg-primary p-6">
        <View style={styles.heroCard}>
          <ThemedText type="smallBold" style={styles.heroKicker}>
            DocVault Mobile
          </ThemedText>
          <ThemedText type="subtitle" style={styles.heroTitle}>
            {user ? `Welcome back, ${user.displayName}` : 'Scan, upload, and find every document from your phone.'}
          </ThemedText>
          <ThemedText type="small" style={styles.heroBody}>
            {user ? `${user.email} · ${user.tenantId.slice(0, 8)}...` : 'Camera capture, OCR status, reminders, search, and secure workspace access in one app.'}
          </ThemedText>
          <View style={styles.heroActions}>
            <Button variant="secondary" onPress={() => router.push('/scan')}>
              <Button.Label>Start scanning</Button.Label>
            </Button>
            <Button variant="secondary" onPress={() => void logout()}>
              <Button.Label>Sign out</Button.Label>
            </Button>
          </View>
        </View>
      </Card>

      <View style={styles.actionsGrid}>
        <HeroActionCard
          title="Camera scan"
          description="Capture documents and send them to OCR."
          onPress={() => router.push('/scan')}
        />
        <HeroActionCard
          title="Upload files"
          description="Pick PDFs, images, or Word documents."
          onPress={() => router.push('/documents')}
        />
        <HeroActionCard
          title="Search"
          description="Find documents by meaning, language, and type."
          onPress={() => router.push('/search')}
        />
        <HeroActionCard
          title="Reminders"
          description="Track due dates, renewals, and expirations."
          onPress={() => router.push('/reminders')}
        />
      </View>
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  heroCard: {
    gap: Spacing.three,
  },
  heroKicker: {
    color: 'white',
    opacity: 0.8,
  },
  heroTitle: {
    color: 'white',
  },
  heroBody: {
    color: 'white',
    opacity: 0.85,
  },
  actionsGrid: {
    gap: Spacing.three,
  },
  heroActions: {
    flexDirection: 'row',
    gap: Spacing.three,
    flexWrap: 'wrap',
  },
});
