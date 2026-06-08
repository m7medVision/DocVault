import { useCallback, useEffect, useState } from 'react';
import { Directory, File, Paths } from 'expo-file-system';

interface CachedFile {
  localUri: string;
  bytes: number;
}

interface UseCachedFileArgs {
  documentId: string;
  remoteUrl: string | null;
  extension: string;
}

interface UseCachedFileResult {
  localUri: string | null;
  bytes: number | null;
  loading: boolean;
  error: string | null;
  reload: () => void;
}

const CACHE_DIR_NAME = 'docvault-previews';

function destinationDir(): Directory {
  return new Directory(Paths.cache, CACHE_DIR_NAME);
}

function destinationFile(documentId: string, extension: string): File {
  const safeExt = extension.replace(/[^a-z0-9]/gi, '').toLowerCase() || 'bin';
  return new File(destinationDir(), `${documentId}.${safeExt}`);
}

function readMeta(file: File): CachedFile | null {
  try {
    if (!file.exists) return null;
    return { localUri: file.uri, bytes: file.size ?? 0 };
  } catch {
    return null;
  }
}

async function downloadTo(remoteUrl: string, destination: File): Promise<File> {
  try {
    return await File.downloadFileAsync(remoteUrl, destination, { idempotent: true });
  } catch {
    // idempotent may not be supported in older runtime; fall back to plain call
    return await File.downloadFileAsync(remoteUrl, destination);
  }
}

export function useCachedFile({
  documentId,
  remoteUrl,
  extension,
}: UseCachedFileArgs): UseCachedFileResult {
  const [state, setState] = useState<Omit<UseCachedFileResult, 'reload'>>({
    localUri: null,
    bytes: null,
    loading: false,
    error: null,
  });
  const [nonce, setNonce] = useState(0);

  const reload = useCallback(() => setNonce((n) => n + 1), []);

  useEffect(() => {
    let cancelled = false;

    if (!documentId || !remoteUrl) {
      setState({ localUri: null, bytes: null, loading: false, error: null });
      return;
    }

    const file = destinationFile(documentId, extension);

    setState((prev) => ({ ...prev, loading: true, error: null }));

    (async () => {
      try {
        const cached = readMeta(file);
        if (cached) {
          if (!cancelled) {
            setState({
              localUri: cached.localUri,
              bytes: cached.bytes,
              loading: false,
              error: null,
            });
          }
          return;
        }

        destinationDir().create();
        const downloaded = await downloadTo(remoteUrl, file);

        if (cancelled) return;

        if (downloaded.exists) {
          setState({
            localUri: downloaded.uri,
            bytes: downloaded.size ?? 0,
            loading: false,
            error: null,
          });
        } else {
          setState({
            localUri: null,
            bytes: null,
            loading: false,
            error: 'Downloaded file not found in cache',
          });
        }
      } catch (err) {
        if (cancelled) return;
        setState({
          localUri: null,
          bytes: null,
          loading: false,
          error: err instanceof Error ? err.message : 'Failed to cache file',
        });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [documentId, remoteUrl, extension, nonce]);

  return { ...state, reload };
}

export function clearCachedFile(documentId: string, extension: string) {
  try {
    destinationFile(documentId, extension).delete();
  } catch {
    // best-effort
  }
}
