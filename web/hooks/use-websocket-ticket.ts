'use client';

import { useQuery } from '@tanstack/react-query';

import { chatApi } from '@/lib/api/chat';
import { queryKeys } from '@/hooks/query-keys';
import { useAuthStore } from '@/store/auth-store';

type UseWebSocketTicketOptions = {
  enabled?: boolean;
};

export const useWebSocketTicket = (options?: UseWebSocketTicketOptions) => {
  const token = useAuthStore((state) => state.token);
  const enabled = options?.enabled ?? true;

  return useQuery({
    queryKey: queryKeys.wsTicket,
    queryFn: () => chatApi.getWebSocketTicket(),
    staleTime: 0,
    gcTime: 0,
    enabled: Boolean(token) && enabled
  });
};
