import { NextRequest, NextResponse } from 'next/server';
import { type RegisterInput } from '@/lib/auth';
import {
  BackendAuthError,
  applyAuthCookies,
  registerWithPassword,
} from '@/lib/server-auth';

export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as Partial<RegisterInput>;
    const session = await registerWithPassword({
      displayName: body.displayName || '',
      email: body.email || '',
      password: body.password || '',
      locale: body.locale === 'ar' ? 'ar' : 'en',
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
      { error: 'Unable to create your account right now' },
      { status: 500 }
    );
  }
}
