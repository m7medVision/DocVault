import { useEffect, useRef, useState } from 'react';

import { API_BASE_URL } from '@/lib/config';
import { getAccessToken } from '@/lib/auth/token-bridge';

export interface ProgressState {
  status: 'idle' | 'processing' | 'completed' | 'failed';
  progress: number;
  message?: string;
}

/**
 * Subscribes to the backend document-processing WebSocket and reports live
 * status. React Native ships a global WebSocket (unlike EventSource), so we use
 * it directly. `enabled` lets callers connect only while a document is still
 * being processed.
 */
export function useUploadProgress(documentId: string | undefined, enabled: boolean): ProgressState {
  const [state, setState] = useState<ProgressState>({ status: 'idle', progress: 0 });
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!documentId || !enabled) return;

    const token = getAccessToken();
    if (!token) return;

    const wsBase = API_BASE_URL.replace(/^http/, 'ws');
    const url = `${wsBase}/documents/${documentId}/progress/ws?token=${token}`;

    let closed = false;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data as string);
        if (data.error) return;
        setState({
          status:
            data.status === 'failed'
              ? 'failed'
              : data.status === 'completed'
                ? 'completed'
                : 'processing',
          progress:
            data.status === 'completed'
              ? 100
              : data.status === 'failed'
                ? 0
                : (data.progress ?? 0),
          message: data.message,
        });
        if (data.status === 'completed' || data.status === 'failed') {
          ws.close();
        }
      } catch {
        /* ignore malformed frames */
      }
    };

    ws.onerror = () => {
      if (!closed) setState((s) => ({ ...s, status: s.status === 'idle' ? 'idle' : s.status }));
    };

    return () => {
      closed = true;
      ws.close();
      wsRef.current = null;
    };
  }, [documentId, enabled]);

  return state;
}
