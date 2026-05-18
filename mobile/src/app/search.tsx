import { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { Button, Card, Input } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';

export default function SearchScreen() {
  const [query, setQuery] = useState('');

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <ThemedText type="subtitle">Search</ThemedText>
        <ThemedText type="small" themeColor="textSecondary">
          Find documents by content, OCR text, type, and language.
        </ThemedText>
      </View>

      <Card className="rounded-3xl border border-divider bg-content1 p-4">
        <View style={styles.searchCard}>
          <Input placeholder="Search documents..." value={query} onChangeText={setQuery} />
          <Button variant="primary" isDisabled={!query.trim()}>
            <Button.Label>Search</Button.Label>
          </Button>
        </View>
      </Card>
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: {
    gap: Spacing.one,
  },
  searchCard: {
    gap: Spacing.three,
  },
});
