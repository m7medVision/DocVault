import { fetch as expoFetch } from 'expo/fetch';

import { API_BASE_URL } from '@/lib/config';
import { getAccessToken, refreshToken } from '@/lib/auth/token-bridge';

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

/**
 * Streams a chat response for a document over Server-Sent Events.
 *
 * React Native's global fetch does not expose a readable stream and there is no
 * EventSource, so we use `expo/fetch` which returns a real streaming
 * `response.body`. The backend emits `data: {json}\n\n` chunks where a chunk of
 * type `TEXT_MESSAGE_CONTENT` carries an incremental `delta` string.
 *
 * onToken is called for each delta. Resolves when the stream ends.
 */
export async function streamChat(
  documentId: string,
  messages: ChatMessage[],
  onToken: (delta: string) => void,
  signal?: AbortSignal,
): Promise<void> {
  let token = getAccessToken();
  if (!token) {
    token = await refreshToken();
  }

  const response = await expoFetch(`${API_BASE_URL}/documents/${documentId}/chat`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ messages }),
    signal,
  });

  if (!response.ok) {
    const body = await response.text().catch(() => '');
    let message = `Chat error ${response.status}`;
    try {
      const parsed = JSON.parse(body);
      if (parsed?.error) message = parsed.error;
    } catch {
      /* keep default */
    }
    throw new Error(message);
  }

  const body = response.body;
  if (!body) {
    // Streaming unavailable — fall back to buffered text.
    const text = await response.text();
    parseSseBuffer(text, onToken);
    return;
  }

  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    const events = buffer.split('\n\n');
    buffer = events.pop() ?? '';
    for (const event of events) {
      handleEvent(event, onToken);
    }
  }

  if (buffer.trim()) {
    handleEvent(buffer, onToken);
  }
}

function handleEvent(rawEvent: string, onToken: (delta: string) => void): void {
  const line = rawEvent.split('\n').find((l) => l.startsWith('data: '));
  if (!line) return;
  const data = line.slice('data: '.length).trim();
  if (!data || data === '[DONE]') return;
  try {
    const chunk = JSON.parse(data);
    if (chunk.type === 'TEXT_MESSAGE_CONTENT' && typeof chunk.delta === 'string') {
      onToken(chunk.delta);
    } else if (chunk.type === 'RUN_ERROR') {
      throw new Error(chunk?.error?.message || 'Chat failed');
    }
  } catch (e) {
    if (e instanceof Error && e.message !== 'Chat failed') return; // ignore malformed chunks
    throw e;
  }
}

function parseSseBuffer(text: string, onToken: (delta: string) => void): void {
  for (const event of text.split('\n\n')) {
    handleEvent(event, onToken);
  }
}
