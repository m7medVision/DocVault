import { StyleSheet, View } from 'react-native';

import { ThemedText } from './themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import type { DocumentVersion } from '@/features/documents/types';

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function VersionTimeline({ versions }: { versions: DocumentVersion[] }) {
  const theme = useTheme();
  const { t } = useTranslation();

  if (!versions || versions.length === 0) {
    return (
      <ThemedText type="small" themeColor="muted">
        {t('versions.empty')}
      </ThemedText>
    );
  }

  const sorted = [...versions].sort((a, b) => b.version_number - a.version_number);
  const current = sorted[0]?.version_number;

  return (
    <View style={styles.list}>
      {sorted.map((v) => (
        <View
          key={v.id}
          style={[styles.row, { backgroundColor: theme.surface, borderColor: theme.border }]}
        >
          <View style={styles.topRow}>
            <ThemedText type="smallBold">
              {t('versions.version', { number: v.version_number })}
            </ThemedText>
            {v.version_number === current && (
              <View style={[styles.badge, { backgroundColor: theme.accent }]}>
                <ThemedText type="code" style={{ color: '#fff' }}>
                  {t('versions.current')}
                </ThemedText>
              </View>
            )}
          </View>
          <ThemedText type="code" themeColor="textSecondary">
            {v.mime_type} · {formatBytes(v.file_size)} ·{' '}
            {new Date(v.created_at).toLocaleDateString()}
          </ThemedText>
          {v.uploaded_by ? (
            <ThemedText type="code" themeColor="textSecondary">
              {t('versions.uploadedBy', { name: v.uploaded_by })}
            </ThemedText>
          ) : null}
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  list: { gap: Spacing.two },
  row: {
    borderRadius: Spacing.three,
    borderWidth: StyleSheet.hairlineWidth,
    padding: Spacing.three,
    gap: 2,
  },
  topRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  badge: {
    paddingHorizontal: Spacing.two,
    paddingVertical: Spacing.half,
    borderRadius: 999,
  },
});
