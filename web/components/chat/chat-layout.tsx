'use client';

import { useState } from 'react';

import { ChatPanel } from '@/components/chat/chat-panel';
import { ConversationList } from '@/components/chat/conversation-list';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetClose, SheetContent, SheetHeader, SheetOverlay, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { useChatStore } from '@/store/chat-store';
import { useChatRealtime } from '@/websocket/use-chat-realtime';

const MenuIcon = () => (
  <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8">
    <path d="M4 7h16" />
    <path d="M4 12h16" />
    <path d="M4 17h16" />
  </svg>
);

const CloseIcon = () => (
  <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.8">
    <path d="M6 6l12 12" />
    <path d="M18 6L6 18" />
  </svg>
);

export const ChatLayout = () => {
  const activeConversationId = useChatStore((state) => state.activeConversationId);
  const [open, setOpen] = useState(false);

  useChatRealtime(activeConversationId);

  return (
    <main className="grid h-dvh grid-rows-[auto_1fr] bg-background px-3 py-3 text-foreground sm:px-4 sm:py-4 lg:grid-rows-1">
      <div className="mb-3 flex items-center justify-between rounded-xl border border-border/70 bg-muted/30 px-3 py-2 shadow-sm backdrop-blur-md lg:hidden">
        <div>
          <h1 className="text-sm font-semibold">Realtime Chat</h1>
          <p className="text-xs text-muted-foreground">Messages</p>
        </div>

        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger>
            <Button variant="outline" size="sm" className="h-9 w-9 rounded-full p-0 transition-all duration-200 hover:scale-105">
              <MenuIcon />
            </Button>
          </SheetTrigger>

          <SheetOverlay />
          <SheetContent side="left" className="p-0">
            <SheetHeader className="py-3">
              <SheetTitle>Conversations</SheetTitle>
              <SheetClose>
                <Button variant="outline" size="sm" className="h-8 w-8 rounded-full p-0">
                  <CloseIcon />
                </Button>
              </SheetClose>
            </SheetHeader>
            <Separator />
            <div className="h-[calc(100%-57px)] p-3">
              <ConversationList onConversationSelect={() => setOpen(false)} />
            </div>
          </SheetContent>
        </Sheet>
      </div>

      <div className="grid min-h-0 flex-1 gap-3 lg:grid-cols-[340px_1fr] lg:gap-4">
        <ConversationList className="hidden lg:flex" />
        <ChatPanel conversationId={activeConversationId} />
      </div>
    </main>
  );
};
