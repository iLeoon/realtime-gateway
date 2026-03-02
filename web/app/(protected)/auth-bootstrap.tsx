'use client';

import { useEffect } from 'react';

import { tokenStore } from '@/lib/api/token';
import { useAuthStore } from '@/store/auth-store';

type AuthBootstrapProps = {
  token: string;
};

export const AuthBootstrap = ({ token }: AuthBootstrapProps) => {
  useEffect(() => {
    tokenStore.setToken(token);
    useAuthStore.setState((state) => ({ ...state, token }));
  }, [token]);

  return null;
};
