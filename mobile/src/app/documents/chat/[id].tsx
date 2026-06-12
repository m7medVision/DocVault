import { useState } from 'react';
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useLocalSearchParams, useRouter } from 'expo-router';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useChat } from '@/features/chat/use-chat';
import { useTranslation } from '@/lib/i18n';

export default function DocumentChatScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const theme = useTheme();
  const { t } = useTranslation();
  const { messages, streaming, error, send } = useChat(id);
  const [input, setInput] = useState('');

  const handleSend = () => {
    const text = input;
    setInput('');
    void send(text);
  };

  return (
    <SafeAreaView style={[styles.root, { backgroundColor: theme.background }]} edges={['top', 'bottom']}>
      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <View style={styles.header}>
          <Pressable onPress={() => router.back()}>
            <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
          </Pressable>
          <ThemedText type="subtitle">{t('chat.title')}</ThemedText>
        </View>

        <ScrollView style={styles.flex} contentContainerStyle={styles.messages}>
          {messages.length === 0 ? (
            <View style={styles.empty}>
              <ThemedText type="small" themeColor="muted">
                {t('chat.empty')}
              </ThemedText>
            </View>
          ) : (
            messages.map((m) => (
              <View
                key={m.id}
                style={[
                  styles.bubble,
                  m.role === 'user'
                    ? [styles.userBubble, { backgroundColor: theme.accent }]
                    : [styles.assistantBubble, { backgroundColor: theme.surface }],
                ]}
              >
                <ThemedText
                  type="small"
                  style={m.role === 'user' ? { color: '#fff' } : undefined}
                >
                  {m.content || (streaming ? t('chat.thinking') : '')}
                </ThemedText>
              </View>
            ))
          )}
          {error ? (
            <ThemedText type="small" themeColor="error">
              {error}
            </ThemedText>
          ) : null}
        </ScrollView>

        <View style={[styles.inputRow, { borderColor: theme.border }]}>
          <TextInput
            value={input}
            onChangeText={setInput}
            placeholder={t('chat.placeholder')}
            placeholderTextColor={theme.textSecondary}
            style={[styles.input, { backgroundColor: theme.surface, color: theme.text }]}
            multiline
            onSubmitEditing={handleSend}
          />
          <Pressable
            onPress={handleSend}
            disabled={streaming || !input.trim()}
            style={[
              styles.sendBtn,
              { backgroundColor: theme.accent, opacity: streaming || !input.trim() ? 0.5 : 1 },
            ]}
          >
            {streaming ? (
              <ActivityIndicator size="small" color="#fff" />
            ) : (
              <ThemedText type="smallBold" style={{ color: '#fff' }}>
                {t('chat.send')}
              </ThemedText>
            )}
          </Pressable>
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1 },
  flex: { flex: 1 },
  header: {
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
    gap: Spacing.one,
  },
  messages: {
    padding: Spacing.three,
    gap: Spacing.two,
  },
  empty: {
    alignItems: 'center',
    paddingVertical: Spacing.five,
  },
  bubble: {
    maxWidth: '85%',
    borderRadius: Spacing.three,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
  },
  userBubble: { alignSelf: 'flex-end' },
  assistantBubble: { alignSelf: 'flex-start' },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    gap: Spacing.two,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
    borderTopWidth: StyleSheet.hairlineWidth,
  },
  input: {
    flex: 1,
    maxHeight: 120,
    borderRadius: Spacing.three,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
  },
  sendBtn: {
    borderRadius: Spacing.three,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two + 2,
    justifyContent: 'center',
  },
});
