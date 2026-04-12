import { NextRequest, NextResponse } from 'next/server';
import { type LoginInput } from '@/lib/auth';
import {
  BackendAuthError,
  applyAuthCookies,
  loginWithPassword,
} from '@/lib/server-auth';

export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as Partial<LoginInput>;
    const session = await loginWithPassword({
      email: body.email || '',
      password: body.password || '',
      rememberMe: Boolean(body.rememberMe),
    });

    const response = NextResponse.json({ session });
    applyAuthCookies(response, session, Boolean(body.rememberMe));
    return response;
  } catch (error) {
    if (error instanceof BackendAuthError) {
      return NextResponse.json(
        { error: error.message },
        { status: error.status }
      );
    }

    return NextResponse.json(
      { error: 'Unable to sign in right now' },
      { status: 500 }
    );
  }
}
