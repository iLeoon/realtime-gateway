'use client';

import { useQuery } from '@tanstack/react-query';

import { chatApi } from '@/lib/api/chat';
import { queryKeys } from '@/hooks/query-keys';
import { useAuthStore } from '@/store/auth-store';

export const useConversations = () => {
  const token = useAuthStore((state) => state.token);

  return useQuery({
    queryKey: queryKeys.conversations,
    queryFn: () => chatApi.getConversations(),
    enabled: Boolean(token)
  });
};
