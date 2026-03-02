'use client';

import { useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';

import { decodeJwtSub } from '@/lib/auth/jwt';
import { queryKeys } from '@/hooks/query-keys';
import { useWebSocketTicket } from '@/hooks/use-websocket-ticket';
import { useAuthStore } from '@/store/auth-store';
import { useChatStore } from '@/store/chat-store';
import { useWebSocketStore } from '@/store/websocket-store';

type RawRealtimePayload = {
  content: string | null;
  authorId: string | null;
  conversationId: string | null;
};

const extractRawPayload = (rawText: string): RawRealtimePayload => {
  const trimmed = rawText.trim();
  if (!trimmed) {
    return { content: null, authorId: null, conversationId: null };
  }

  // Some transports forward JSON encoded strings (e.g. "Hello").
  try {
    const parsed: unknown = JSON.parse(trimmed);
    if (typeof parsed === 'string') {
      const content = parsed.trim();
      return {
        content: content || null,
        authorId: null,
        conversationId: null
      };
    }

    if (parsed && typeof parsed === 'object') {
      const obj = parsed as Record<string, unknown>;

      const content = [obj.content, obj.resContent, obj.ResContent]
        .find((value) => typeof value === 'string')
        ?.toString()
        .trim();

      const authorRaw = [obj.authorID, obj.authorId, obj.AuthorID, obj.creatorId]
        .find((value) => typeof value === 'string' || typeof value === 'number')
        ?.toString();

      const conversationRaw = [obj.conversationID, obj.conversationId, obj.ConversationID]
        .find((value) => typeof value === 'string' || typeof value === 'number')
        ?.toString();

      return {
        content: content || null,
        authorId: authorRaw && authorRaw !== '0' ? authorRaw : null,
        conversationId: conversationRaw && conversationRaw !== '0' ? conversationRaw : null
      };
    }
  } catch {
    // Fall through to packet/plain-text parsing.
  }

  // Handle packet-like string payloads, e.g.
  // ResponseMessagePacket{ToConnectionID: 1, AuthorID: 2, ConversationID: 3, ResContent: "Hello"}
  const content = trimmed.match(/ResContent:\s*"([\s\S]*)"\s*\}?$/)?.[1] ?? trimmed;
  const authorId = trimmed.match(/AuthorID:\s*(\d+)/)?.[1] ?? null;
  const conversationId = trimmed.match(/ConversationID:\s*(\d+)/)?.[1] ?? null;

  return {
    content: content.trim() || null,
    authorId,
    conversationId
  };
};

export const useChatRealtime = (activeConversationId: string | null) => {
  const queryClient = useQueryClient();
  const { data: wsTicket } = useWebSocketTicket({ enabled: Boolean(activeConversationId) });
  const addMessage = useChatStore((state) => state.addMessage);
  const addOptimisticMessage = useChatStore((state) => state.addOptimisticMessage);
  const incrementUnread = useChatStore((state) => state.incrementUnread);
  const reconcileOptimisticMessage = useChatStore((state) => state.reconcileOptimisticMessage);
  const currentUser = useAuthStore((state) => state.user);
  const token = useAuthStore((state) => state.token);
  const connect = useWebSocketStore((state) => state.connect);
  const disconnect = useWebSocketStore((state) => state.disconnect);
  const setInboundHandler = useWebSocketStore((state) => state.setInboundHandler);

  useEffect(() => {
    setInboundHandler(({ message, rawText }) => {
      const currentUserId = currentUser ? String(currentUser.userId) : decodeJwtSub(token);

      if (!message && rawText) {
        const { content, authorId, conversationId } = extractRawPayload(rawText);
        if (!content) {
          return;
        }

        if (authorId && conversationId) {
          addOptimisticMessage({
            id: `ws-packet-${crypto.randomUUID()}`,
            creatorId: authorId,
            conversationId,
            content,
            createdAt: new Date().toISOString()
          });

          const isOwnMessage = currentUserId !== null && currentUserId === authorId;
          if (!isOwnMessage && conversationId !== activeConversationId) {
            incrementUnread(conversationId);
          }
          void queryClient.invalidateQueries({ queryKey: queryKeys.messages(conversationId) });
        }

        void queryClient.invalidateQueries({ queryKey: queryKeys.conversations });

        if (!conversationId && activeConversationId) {
          void queryClient.invalidateQueries({ queryKey: queryKeys.messages(activeConversationId) });
        }
        return;
      }

      if (!message) {
        return;
      }

      addMessage(message);
      reconcileOptimisticMessage(message);

      const isOwnMessage = currentUserId !== null && currentUserId === message.creatorId;

      if (!isOwnMessage && message.conversationId !== activeConversationId) {
        incrementUnread(message.conversationId);
      }

      void queryClient.invalidateQueries({ queryKey: queryKeys.conversations });
      void queryClient.invalidateQueries({ queryKey: queryKeys.messages(message.conversationId) });
    });
  }, [
    activeConversationId,
    addMessage,
    addOptimisticMessage,
    currentUser,
    incrementUnread,
    queryClient,
    reconcileOptimisticMessage,
    setInboundHandler,
    token
  ]);

  useEffect(() => {
    if (!wsTicket?.ticket) {
      return;
    }

    connect(wsTicket.ticket);

    return () => {
      disconnect();
    };
  }, [connect, disconnect, wsTicket?.ticket]);

  useEffect(() => {
    if (!activeConversationId) {
      return;
    }

    void queryClient.invalidateQueries({ queryKey: queryKeys.messages(activeConversationId) });
  }, [activeConversationId, queryClient]);
};
