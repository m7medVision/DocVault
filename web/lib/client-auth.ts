let accessToken: string | null = null;
let refreshHandler: (() => Promise<string | null>) | null = null;
let inflightRefresh: Promise<string | null> | null = null;

export function getClientAccessToken(): string | null {
  return accessToken;
}

export function setClientAccessToken(token: string | null): void {
  accessToken = token;
}

export function setRefreshHandler(
  handler: (() => Promise<string | null>) | null
): void {
  refreshHandler = handler;
}

export async function refreshClientAccessToken(): Promise<string | null> {
  if (!refreshHandler) {
    return null;
  }

  if (!inflightRefresh) {
    inflightRefresh = refreshHandler().finally(() => {
      inflightRefresh = null;
    });
  }

  return inflightRefresh;
}
