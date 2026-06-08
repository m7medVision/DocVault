type RefreshFn = () => Promise<string | null>;

let inMemoryAccessToken: string | null = null;
let refreshFn: RefreshFn | null = null;
let inflightRefresh: Promise<string | null> | null = null;

export function getAccessToken(): string | null {
  return inMemoryAccessToken;
}

export function setAccessToken(token: string | null): void {
  inMemoryAccessToken = token;
}

export function setRefreshFunction(fn: RefreshFn | null): void {
  refreshFn = fn;
}

export async function refreshToken(): Promise<string | null> {
  if (!refreshFn) return null;
  if (!inflightRefresh) {
    inflightRefresh = refreshFn().finally(() => {
      inflightRefresh = null;
    });
  }
  return inflightRefresh;
}
