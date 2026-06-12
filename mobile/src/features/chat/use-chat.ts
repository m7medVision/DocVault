import { useCallback, useRef, useState } from 'react';

import { streamChat, type ChatMessage } from '@/lib/api/sse';

export interface ChatTurn {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

export function useChat(documentId: string) {
  const [messages, setMessages] = useState<ChatTurn[]>([]);
  const [streaming, setStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const counter = useRef(0);

  const send = useCallback(
    async (text: string) => {
      const trimmed = text.trim();
      if (!trimmed || streaming) return;

      setError(null);
      const userId = `u${counter.current++}`;
      const assistantId = `a${counter.current++}`;

      const history: ChatMessage[] = [
        ...messages.map((m) => ({ role: m.role, content: m.content })),
        { role: 'user' as const, content: trimmed },
      ];

      setMessages((prev) => [
        ...prev,
        { id: userId, role: 'user', content: trimmed },
        { id: assistantId, role: 'assistant', content: '' },
      ]);
      setStreaming(true);

      try {
        await streamChat(documentId, history, (delta) => {
          setMessages((prev) =>
            prev.map((m) =>
              m.id === assistantId ? { ...m, content: m.content + delta } : m,
            ),
          );
        });
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Chat failed');
        setMessages((prev) => prev.filter((m) => m.id !== assistantId));
      } finally {
        setStreaming(false);
      }
    },
    [documentId, messages, streaming],
  );

  return { messages, streaming, error, send };
}
