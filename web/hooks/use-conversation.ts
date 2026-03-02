'use client';

import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '@/hooks/query-keys';
import { chatApi } from '@/lib/api/chat';

export const useConversation = (id: string | null) =>
  useQuery({
    queryKey: id ? queryKeys.conversation(id) : ['conversation', 'none'],
    queryFn: () => {
      if (!id) {
        throw new Error('Conversation id is required');
      }

      return chatApi.getConversation(id);
    },
    enabled: Boolean(id)
  });
