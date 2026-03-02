import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

const COOKIE_NAMES = ['token', 'auth_token'] as const;

export const POST = async (request: Request) => {
  const body = (await request.json()) as { token?: string };
  const token = body.token;

  if (!token) {
    return NextResponse.json({ error: 'token is required' }, { status: 400 });
  }

  const cookieStore = await cookies();

  COOKIE_NAMES.forEach((name) => {
    cookieStore.set({
      name,
      value: token,
      httpOnly: true,
      sameSite: 'lax',
      secure: process.env.NODE_ENV === 'production',
      path: '/'
    });
  });

  return NextResponse.json({ ok: true });
};

export const DELETE = async () => {
  const cookieStore = await cookies();

  COOKIE_NAMES.forEach((name) => {
    cookieStore.set({
      name,
      value: '',
      expires: new Date(0),
      httpOnly: true,
      sameSite: 'lax',
      secure: process.env.NODE_ENV === 'production',
      path: '/'
    });
  });

  return NextResponse.json({ ok: true });
};
