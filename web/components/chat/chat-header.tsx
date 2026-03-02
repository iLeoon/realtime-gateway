'use client';

import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import type { Conversation } from '@/lib/validation';

import { getConversationTitle, getInitials } from '@/components/chat/chat-utils';

type ChatHeaderProps = {
  conversation: Conversation;
  currentUserId?: string;
};

const PhoneIcon = () => (
  <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.8">
    <path d="M6.6 10.8a15.3 15.3 0 006.6 6.6l2.2-2.2a1.3 1.3 0 011.3-.3 11 11 0 003.5.6 1.2 1.2 0 011.2 1.2v3.4a1.2 1.2 0 01-1.2 1.2A18.8 18.8 0 012.8 4.2 1.2 1.2 0 014 3h3.4A1.2 1.2 0 018.6 4.2a11 11 0 00.6 3.5 1.3 1.3 0 01-.3 1.3l-2.3 1.8z" />
  </svg>
);

const DotsIcon = () => (
  <svg viewBox="0 0 24 24" className="h-4 w-4" fill="currentColor">
    <circle cx="5" cy="12" r="1.8" />
    <circle cx="12" cy="12" r="1.8" />
    <circle cx="19" cy="12" r="1.8" />
  </svg>
);

export const ChatHeader = ({ conversation, currentUserId }: ChatHeaderProps) => {
  const title = getConversationTitle(conversation, currentUserId);

  return (
    <header className="sticky top-0 z-20 flex items-center justify-between border-b border-border/70 bg-background/80 px-5 py-3 backdrop-blur-md">
      <div className="min-w-0">
        <h1 className="truncate text-sm font-semibold text-foreground sm:text-base">{title}</h1>
        <p className="text-xs text-muted-foreground">{conversation.participants.length} participants</p>
      </div>

      <div className="ml-4 flex items-center gap-2">
        <div className="hidden items-center -space-x-2 sm:flex">
          {conversation.participants.slice(0, 3).map((participant) => (
            <Avatar key={participant.id} className="h-7 w-7 border-background transition-transform duration-200 hover:scale-110">
              <AvatarFallback className="text-[10px]">{getInitials(participant.displayName)}</AvatarFallback>
            </Avatar>
          ))}
        </div>

        <Button variant="outline" size="sm" className="h-8 w-8 rounded-full p-0 transition-all duration-200 hover:scale-105">
          <PhoneIcon />
        </Button>
        <Button variant="outline" size="sm" className="h-8 w-8 rounded-full p-0 transition-all duration-200 hover:scale-105">
          <DotsIcon />
        </Button>
      </div>
    </header>
  );
};
