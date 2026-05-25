import { useState } from 'react';
import { Pressable, ScrollView, StyleSheet, View } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import type { DocumentPage } from '@/features/documents/types';

interface DocumentViewerProps {
  pages: DocumentPage[];
  translated?: boolean;
}

export function DocumentViewer({ pages, translated = false }: DocumentViewerProps) {
  const theme = useTheme();
  const [currentPage, setCurrentPage] = useState(0);

  if (pages.length === 0) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        <ThemedText type="small" themeColor="textSecondary">
          No OCR text available yet.
        </ThemedText>
      </View>
    );
  }

  const page = pages[currentPage];
  const text = translated ? page.translated_text : page.ocr_text;

  return (
    <View style={styles.container}>
      <ScrollView style={styles.scroll} contentContainerStyle={styles.scrollContent}>
        {text ? (
          <View style={[styles.textCard, { backgroundColor: theme.surface }]}>
            <ThemedText type="small" style={styles.ocrText}>
              {text}
            </ThemedText>
          </View>
        ) : (
          <View style={[styles.center, { backgroundColor: theme.surface }]}>
            <ThemedText type="small" themeColor="textSecondary">
              No text available for this page.
            </ThemedText>
          </View>
        )}
      </ScrollView>

      <View style={[styles.toolbar, { backgroundColor: theme.surface }]}>
        <Pressable
          onPress={() => setCurrentPage((p) => Math.max(0, p - 1))}
          disabled={currentPage === 0}
          style={[styles.navBtn, currentPage === 0 && styles.navBtnDisabled]}
        >
          <ThemedText type="smallBold">Prev</ThemedText>
        </Pressable>

        <View style={styles.pageInfo}>
          <ThemedText type="code">
            {currentPage + 1} / {pages.length}
          </ThemedText>
          {page.confidence != null && (
            <ThemedText type="code" themeColor={page.confidence >= 80 ? 'success' : 'warning'}>
              {page.confidence.toFixed(1)}%
            </ThemedText>
          )}
        </View>

        <Pressable
          onPress={() => setCurrentPage((p) => Math.min(pages.length - 1, p + 1))}
          disabled={currentPage === pages.length - 1}
          style={[styles.navBtn, currentPage === pages.length - 1 && styles.navBtnDisabled]}
        >
          <ThemedText type="smallBold">Next</ThemedText>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  scroll: {
    flex: 1,
  },
  scrollContent: {
    padding: Spacing.three,
  },
  textCard: {
    borderRadius: Spacing.two,
    padding: Spacing.three,
  },
  ocrText: {
    lineHeight: 22,
  },
  center: {
    alignItems: 'center',
    justifyContent: 'center',
    padding: Spacing.five,
    borderRadius: Spacing.two,
  },
  toolbar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: '#E4E4E7',
  },
  navBtn: {
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one,
  },
  navBtnDisabled: {
    opacity: 0.3,
  },
  pageInfo: {
    alignItems: 'center',
    gap: Spacing.half,
  },
});
