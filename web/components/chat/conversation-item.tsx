'use client';

import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import type { Conversation } from '@/lib/validation';
import { cn } from '@/lib/utils';

import { getConversationTitle, getInitials } from '@/components/chat/chat-utils';

type ConversationItemProps = {
  conversation: Conversation;
  currentUserId?: string;
  isActive: boolean;
  unreadCount: number;
  onClick: () => void;
};

export const ConversationItem = ({ conversation, currentUserId, isActive, unreadCount, onClick }: ConversationItemProps) => {
  const title = getConversationTitle(conversation, currentUserId);
  const subtitle = `${conversation.conversationType === 'group-chat' ? 'Group' : 'Direct'} • ${conversation.participants.length} members`;

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'group flex w-full items-center gap-3 rounded-xl border border-transparent px-3 py-3 text-left transition-all duration-200',
        'hover:bg-accent/40 hover:shadow-sm',
        isActive && 'border-border/70 bg-accent/50 shadow-sm'
      )}
    >
      <Avatar className="h-10 w-10 transition-transform duration-200 group-hover:scale-105">
        <AvatarFallback>{getInitials(title)}</AvatarFallback>
      </Avatar>

      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-foreground">{title}</p>
        <p className="truncate text-xs text-muted-foreground">{subtitle}</p>
      </div>

      {unreadCount > 0 ? <Badge className="ml-auto min-w-6 justify-center rounded-full px-1.5">{unreadCount}</Badge> : null}
    </button>
  );
};
