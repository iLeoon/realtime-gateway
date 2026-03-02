'use client';

import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';

import { formatMessageTime, getInitials } from '@/components/chat/chat-utils';

type MessageBubbleProps = {
  content: string;
  createdAt: string;
  authorName: string;
  isOwn: boolean;
};

export const MessageBubble = ({ content, createdAt, authorName, isOwn }: MessageBubbleProps) => (
  <article className={cn('group flex w-full items-end gap-2 animate-message-in', isOwn ? 'justify-end' : 'justify-start')}>
    {!isOwn ? (
      <Avatar className="h-8 w-8 transition-transform duration-200 group-hover:scale-105">
        <AvatarFallback className="text-[10px]">{getInitials(authorName)}</AvatarFallback>
      </Avatar>
    ) : null}

    <div className={cn('max-w-[65%] space-y-1', isOwn ? 'items-end' : 'items-start')}>
      <div
        className={cn(
          'rounded-2xl px-4 py-2.5 text-sm shadow-sm transition-all duration-200',
          isOwn
            ? 'bg-gradient-to-br from-primary to-primary/80 text-primary-foreground'
            : 'border border-border/70 bg-muted text-foreground'
        )}
      >
        <p className="whitespace-pre-wrap break-words">{content}</p>
      </div>

      <p className={cn('px-1 text-[11px] text-muted-foreground/90', isOwn ? 'text-right' : 'text-left')}>
        {formatMessageTime(createdAt)}
      </p>
    </div>
  </article>
);
