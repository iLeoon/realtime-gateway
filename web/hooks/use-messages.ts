'use client';

import { useQuery } from '@tanstack/react-query';

import { chatApi } from '@/lib/api/chat';
import { queryKeys } from '@/hooks/query-keys';
import { useAuthStore } from '@/store/auth-store';

export const useMessages = (conversationId: string | null) => {
  const token = useAuthStore((state) => state.token);

  return useQuery({
    queryKey: conversationId ? queryKeys.messages(conversationId) : ['messages', 'none'],
    queryFn: () => {
      if (!conversationId) {
        throw new Error('Conversation id is required');
      }

      return chatApi.getMessages(conversationId);
    },
    enabled: Boolean(conversationId) && Boolean(token)
  });
};
