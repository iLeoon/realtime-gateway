'use client';

import { create } from 'zustand';

import { registerLogoutHandler } from '@/lib/api/auth-events';
import { tokenStore } from '@/lib/api/token';
import type { User } from '@/lib/validation';

type AuthState = {
  token: string | null;
  user: User | null;
  login: (params: { token: string; user: User }) => Promise<void>;
  logout: () => Promise<void>;
  setToken: (token: string | null) => Promise<void>;
};

const SESSION_ROUTE = '/api/auth/session';

const persistSessionCookie = async (token: string): Promise<void> => {
  await fetch(SESSION_ROUTE, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token })
  });
};

const clearSessionCookie = async (): Promise<void> => {
  await fetch(SESSION_ROUTE, { method: 'DELETE' });
};

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  login: async ({ token, user }) => {
    tokenStore.setToken(token);
    await persistSessionCookie(token);
    set({ token, user });
  },
  logout: async () => {
    tokenStore.clearToken();
    await clearSessionCookie();
    set({ token: null, user: null });
  },
  setToken: async (token) => {
    if (!token) {
      tokenStore.clearToken();
      await clearSessionCookie();
      set({ token: null });
      return;
    }

    tokenStore.setToken(token);
    await persistSessionCookie(token);
    set({ token });
  }
}));

registerLogoutHandler(() => {
  void useAuthStore.getState().logout();
});
