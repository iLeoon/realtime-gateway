'use client';

import { useQuery } from '@tanstack/react-query';

import { chatApi } from '@/lib/api/chat';
import { queryKeys } from '@/hooks/query-keys';

export const useParticipants = (conversationId: string | null) =>
  useQuery({
    queryKey: conversationId ? queryKeys.participants(conversationId) : ['participants', 'none'],
    queryFn: () => {
      if (!conversationId) {
        throw new Error('Conversation id is required');
      }

      return chatApi.getParticipants(conversationId);
    },
    enabled: Boolean(conversationId)
  });
