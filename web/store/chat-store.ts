'use client';

import { create } from 'zustand';

import type { Conversation, Message } from '@/lib/validation';

type ChatState = {
  conversations: Conversation[];
  activeConversationId: string | null;
  unreadByConversationId: Record<string, number>;
  optimisticMessagesByConversationId: Record<string, Message[]>;
  setActiveConversation: (conversationId: string) => void;
  incrementUnread: (conversationId: string) => void;
  addMessage: (message: Message) => void;
  addOptimisticMessage: (message: Message) => void;
  reconcileOptimisticMessage: (message: Message) => void;
  pruneOptimisticMessages: (conversationId: string, persistedMessages: Message[]) => void;
  setConversations: (conversations: Conversation[]) => void;
};

export const useChatStore = create<ChatState>((set, get) => ({
  conversations: [],
  activeConversationId: null,
  unreadByConversationId: {},
  optimisticMessagesByConversationId: {},
  setActiveConversation: (conversationId) => {
    set((state) => ({
      activeConversationId: conversationId,
      unreadByConversationId: {
        ...state.unreadByConversationId,
        [conversationId]: 0
      }
    }));
  },
  incrementUnread: (conversationId) => {
    set((state) => ({
      unreadByConversationId: {
        ...state.unreadByConversationId,
        [conversationId]: (state.unreadByConversationId[conversationId] ?? 0) + 1
      }
    }));
  },
  addMessage: (message) => {
    const current = get().conversations;
    const exists = current.some((conversation) => conversation.id === message.conversationId);

    if (!exists) {
      return;
    }

    const reordered = [
      ...current.filter((conversation) => conversation.id === message.conversationId),
      ...current.filter((conversation) => conversation.id !== message.conversationId)
    ];

    set({ conversations: reordered });
  },
  addOptimisticMessage: (message) => {
    set((state) => {
      const current = state.optimisticMessagesByConversationId[message.conversationId] ?? [];
      const incomingTs = new Date(message.createdAt).getTime();
      const hasRecentDuplicate = current.some((item) => {
        if (item.creatorId !== message.creatorId || item.content !== message.content) {
          return false;
        }

        const delta = Math.abs(new Date(item.createdAt).getTime() - incomingTs);
        return delta <= 3000;
      });

      if (hasRecentDuplicate) {
        return state;
      }

      return {
        optimisticMessagesByConversationId: {
          ...state.optimisticMessagesByConversationId,
          [message.conversationId]: [...current, message]
        }
      };
    });
  },
  reconcileOptimisticMessage: (message) => {
    set((state) => {
      const current = state.optimisticMessagesByConversationId[message.conversationId] ?? [];
      const targetTimestamp = new Date(message.createdAt).getTime();
      const matchIndex = current.findIndex((item) => {
        if (item.creatorId !== message.creatorId || item.content !== message.content) {
          return false;
        }

        const delta = Math.abs(new Date(item.createdAt).getTime() - targetTimestamp);
        return delta <= 15000;
      });

      if (matchIndex < 0) {
        return state;
      }

      const updated = [...current.slice(0, matchIndex), ...current.slice(matchIndex + 1)];
      return {
        optimisticMessagesByConversationId: {
          ...state.optimisticMessagesByConversationId,
          [message.conversationId]: updated
        }
      };
    });
  },
  pruneOptimisticMessages: (conversationId, persistedMessages) => {
    set((state) => {
      const optimistic = state.optimisticMessagesByConversationId[conversationId] ?? [];
      if (optimistic.length === 0) {
        return state;
      }

      const filtered = optimistic.filter((optimisticMessage) => {
        const optimisticTs = new Date(optimisticMessage.createdAt).getTime();
        return !persistedMessages.some((persisted) => {
          if (persisted.creatorId !== optimisticMessage.creatorId || persisted.content !== optimisticMessage.content) {
            return false;
          }

          const delta = Math.abs(new Date(persisted.createdAt).getTime() - optimisticTs);
          return delta <= 15000;
        });
      });

      if (filtered.length === optimistic.length) {
        return state;
      }

      return {
        optimisticMessagesByConversationId: {
          ...state.optimisticMessagesByConversationId,
          [conversationId]: filtered
        }
      };
    });
  },
  setConversations: (conversations) => {
    set((state) => {
      const unreadByConversationId = Object.fromEntries(
        conversations.map((conversation) => [conversation.id, state.unreadByConversationId[conversation.id] ?? 0])
      );

      return {
        conversations,
        activeConversationId: state.activeConversationId,
        unreadByConversationId
      };
    });
  }
}));
