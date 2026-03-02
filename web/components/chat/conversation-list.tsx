'use client';

import { useEffect } from 'react';

import { ConversationItem } from '@/components/chat/conversation-item';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { useConversations } from '@/hooks/use-conversations';
import { useAuthStore } from '@/store/auth-store';
import { useChatStore } from '@/store/chat-store';
import { cn } from '@/lib/utils';

type ConversationListProps = {
  className?: string;
  onConversationSelect?: () => void;
};

export const ConversationList = ({ className, onConversationSelect }: ConversationListProps) => {
  const { data, isPending, error } = useConversations();
  const conversations = useChatStore((state) => state.conversations);
  const activeConversationId = useChatStore((state) => state.activeConversationId);
  const unreadByConversationId = useChatStore((state) => state.unreadByConversationId);
  const setConversations = useChatStore((state) => state.setConversations);
  const setActiveConversation = useChatStore((state) => state.setActiveConversation);
  const currentUser = useAuthStore((state) => state.user);

  useEffect(() => {
    if (data) {
      setConversations(data);
    }
  }, [data, setConversations]);

  return (
    <aside
      className={cn(
        'flex h-full min-h-0 flex-col rounded-xl border border-border/70 bg-muted/40 shadow-sm backdrop-blur-md',
        className
      )}
    >
      <div className="sticky top-0 z-10 rounded-t-xl bg-muted/40 px-4 py-3 backdrop-blur-md">
        <h2 className="text-sm font-semibold tracking-wide text-foreground">Conversations</h2>
      </div>
      <Separator className="opacity-70" />

      {isPending ? <p className="px-4 py-4 text-sm text-muted-foreground">Loading conversations...</p> : null}
      {error ? <p className="px-4 py-4 text-sm text-red-400">Unable to load conversations.</p> : null}

      <ScrollArea className="min-h-0 flex-1 px-3 py-3">
        <div className="space-y-1.5">
          {conversations.map((conversation) => (
            <ConversationItem
              key={conversation.id}
              conversation={conversation}
              currentUserId={currentUser ? String(currentUser.userId) : undefined}
              isActive={activeConversationId === conversation.id}
              unreadCount={unreadByConversationId[conversation.id] ?? 0}
              onClick={() => {
                setActiveConversation(conversation.id);
                onConversationSelect?.();
              }}
            />
          ))}
        </div>
      </ScrollArea>
    </aside>
  );
};
