'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';

import { chatApi } from '@/lib/api/chat';
import { queryKeys } from '@/hooks/query-keys';
import type { ConversationRequest } from '@/lib/validation';

export const useCreateConversation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: ConversationRequest) => chatApi.createConversation(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.conversations });
    }
  });
};
