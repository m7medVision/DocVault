import { AuthSessionPayload } from '../auth';

export class TokenRefresher {
  async refresh(refreshToken: string): Promise<AuthSessionPayload> {
    const { refreshWithToken } = await import('../auth');
    return refreshWithToken(refreshToken);
  }
}

export const tokenRefresher = new TokenRefresher();
