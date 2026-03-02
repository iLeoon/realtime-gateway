import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';

const PROTECTED_PREFIX = '/chat';

export const middleware = (request: NextRequest) => {
  const token = request.cookies.get('token')?.value ?? request.cookies.get('auth_token')?.value;
  const isProtectedRoute = request.nextUrl.pathname.startsWith(PROTECTED_PREFIX);

  if (isProtectedRoute && !token) {
    const loginUrl = new URL('/login', request.url);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
};

export const config = {
  matcher: ['/chat/:path*']
};
