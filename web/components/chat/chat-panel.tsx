'use client';

import { useEffect, useMemo, useRef } from 'react';

import { ChatHeader } from '@/components/chat/chat-header';
import { ChatInput } from '@/components/chat/chat-input';
import { MessageBubble } from '@/components/chat/message-bubble';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useMessages } from '@/hooks/use-messages';
import { decodeJwtSub } from '@/lib/auth/jwt';
import { useAuthStore } from '@/store/auth-store';
import { useChatStore } from '@/store/chat-store';
import { useWebSocketStore } from '@/store/websocket-store';

type ChatPanelProps = {
  conversationId: string | null;
};

export const ChatPanel = ({ conversationId }: ChatPanelProps) => {
  const { data: messages, isPending } = useMessages(conversationId);
  const conversations = useChatStore((state) => state.conversations);
  const optimisticMessagesByConversationId = useChatStore((state) => state.optimisticMessagesByConversationId);
  const addOptimisticMessage = useChatStore((state) => state.addOptimisticMessage);
  const pruneOptimisticMessages = useChatStore((state) => state.pruneOptimisticMessages);
  const sendMessage = useWebSocketStore((state) => state.sendMessage);
  const currentUser = useAuthStore((state) => state.user);
  const token = useAuthStore((state) => state.token);
  const currentUserId = currentUser ? String(currentUser.userId) : decodeJwtSub(token);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const conversation = useMemo(
    () => conversations.find((item) => item.id === conversationId) ?? null,
    [conversations, conversationId]
  );

  const optimisticMessages = useMemo(
    () => (conversationId ? optimisticMessagesByConversationId[conversationId] ?? [] : []),
    [conversationId, optimisticMessagesByConversationId]
  );

  const sortedMessages = useMemo(() => {
    const persistedMessages = messages ?? [];
    const merged = [...persistedMessages];

    optimisticMessages.forEach((optimisticMessage) => {
      const optimisticTs = new Date(optimisticMessage.createdAt).getTime();
      const exists = persistedMessages.some((persisted) => {
        if (persisted.creatorId !== optimisticMessage.creatorId || persisted.content !== optimisticMessage.content) {
          return false;
        }

        const delta = Math.abs(new Date(persisted.createdAt).getTime() - optimisticTs);
        return delta <= 15000;
      });

      if (!exists) {
        merged.push(optimisticMessage);
      }
    });

    return merged.sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
  }, [messages, optimisticMessages]);

  useEffect(() => {
    if (!conversationId || !messages) {
      return;
    }

    pruneOptimisticMessages(conversationId, messages);
  }, [conversationId, messages, pruneOptimisticMessages]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [conversationId, sortedMessages.length]);

  if (!conversationId || !conversation) {
    return (
      <section className="flex h-full items-center justify-center rounded-xl border border-border/70 bg-muted/20 p-6 shadow-sm">
        <div className="max-w-sm space-y-2 text-center">
          <h3 className="text-lg font-semibold text-foreground">Select a conversation</h3>
          <p className="text-sm text-muted-foreground">Choose a chat from the sidebar to view messages and start talking.</p>
        </div>
      </section>
    );
  }

  return (
    <section className="relative flex h-full min-h-0 flex-col overflow-hidden rounded-xl border border-border/70 bg-background/40 shadow-sm backdrop-blur-md">
      <ChatHeader conversation={conversation} currentUserId={currentUser ? String(currentUser.userId) : undefined} />

      <ScrollArea className="min-h-0 flex-1 px-4 py-4 sm:px-6">
        <div className="space-y-3">
          {isPending ? <p className="text-sm text-muted-foreground">Loading messages...</p> : null}

          {sortedMessages.map((message) => {
            const isOwn = currentUserId ? message.creatorId === currentUserId : false;
            const author =
              conversation.participants.find((participant) => participant.id === message.creatorId)?.displayName ??
              `User ${message.creatorId}`;

            return (
              <MessageBubble
                key={message.id}
                content={message.content}
                createdAt={message.createdAt}
                authorName={author}
                isOwn={isOwn}
              />
            );
          })}

          <div ref={messagesEndRef} />
        </div>
      </ScrollArea>

      <ChatInput
        onSend={(body) => {
          const senderId = currentUserId ?? '';
          const recipientUserID =
            conversation.participants.find((participant) => participant.id !== senderId)?.id ??
            conversation.creatorId;

          addOptimisticMessage({
            id: `optimistic-${crypto.randomUUID()}`,
            creatorId: senderId || conversation.creatorId,
            conversationId,
            content: body,
            createdAt: new Date().toISOString()
          });

          sendMessage({
            type: 'send_message',
            payload: {
              content: body,
              conversationID: conversationId,
              recipientUserID
            }
          });
        }}
      />
    </section>
  );
};
