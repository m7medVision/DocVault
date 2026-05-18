import { Pressable, StyleSheet, View } from 'react-native';
import { Card } from 'heroui-native';

import { Spacing } from '@/constants/theme';
import { ThemedText } from './themed-text';

interface HeroActionCardProps {
  title: string;
  description: string;
  onPress?: () => void;
}

export function HeroActionCard({ title, description, onPress }: HeroActionCardProps) {
  return (
    <Pressable onPress={onPress} style={({ pressed }) => pressed && styles.pressed}>
      <Card className="rounded-3xl border border-divider bg-content1 p-4">
        <View style={styles.cardBody}>
          <ThemedText type="smallBold">{title}</ThemedText>
          <ThemedText type="small" themeColor="textSecondary">
            {description}
          </ThemedText>
        </View>
      </Card>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  pressed: {
    opacity: 0.75,
  },
  cardBody: {
    gap: Spacing.one,
  },
});
