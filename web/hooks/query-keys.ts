export const queryKeys = {
  conversations: ['conversations'] as const,
  conversation: (id: string) => ['conversation', id] as const,
  messages: (conversationId: string) => ['messages', conversationId] as const,
  participants: (conversationId: string) => ['participants', conversationId] as const,
  wsTicket: ['ws-ticket'] as const
};
