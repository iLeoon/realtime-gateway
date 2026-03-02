'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

import { useAuthStore } from '@/store/auth-store';

const getTokenFromLocation = (): string | null => {
  const url = new URL(window.location.href);
  const queryToken = url.searchParams.get('token');
  if (queryToken) {
    return queryToken;
  }

  const hash = url.hash.startsWith('#') ? url.hash.slice(1) : url.hash;
  const hashParams = new URLSearchParams(hash);
  const hashToken = hashParams.get('token');
  return hashToken;
};

const AuthCallbackPage = () => {
  const router = useRouter();
  const setToken = useAuthStore((state) => state.setToken);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const run = async () => {
      const token = getTokenFromLocation();

      if (!token) {
        setError('Missing auth token in callback URL.');
        return;
      }

      await setToken(token);
      router.replace('/chat');
    };

    void run();
  }, [router, setToken]);

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center justify-center px-6">
      <div className="w-full rounded-xl border bg-card p-6 shadow-sm">
        <h1 className="mb-2 text-xl font-semibold">Completing sign-in</h1>
        {error ? (
          <p className="text-sm text-red-600">{error}</p>
        ) : (
          <p className="text-sm text-muted-foreground">Please wait while we finish authentication.</p>
        )}
      </div>
    </main>
  );
};

export default AuthCallbackPage;
