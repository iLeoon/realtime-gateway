'use client';

import { Button } from '@/components/ui/button';

const OAUTH_LOGIN_URL = 'http://localhost:7000/api/v1.0/auth/login';

const LoginPage = () => {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center justify-center px-6">
      <div className="w-full rounded-xl border bg-card p-6 shadow-sm">
        <h1 className="mb-2 text-xl font-semibold">Sign In</h1>
        <p className="mb-6 text-sm text-muted-foreground">
          Authenticate using your Google account.
        </p>
        <Button className="w-full" onClick={() => (window.location.href = OAUTH_LOGIN_URL)}>
          Login with Google
        </Button>
      </div>
    </main>
  );
};

export default LoginPage;
