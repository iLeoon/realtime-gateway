import { z } from 'zod';

export const ConversationTypeSchema = z.enum(['private-chat', 'group-chat']);

export const ParticipantSchema = z
  .object({
    id: z.string().min(1),
    displayName: z.string().min(1),
    email: z.string().email(),
    joinedDate: z.string().datetime({ offset: true }),
    role: z.string().min(1)
  })
  .strict();

export const ConversationSchema = z
  .object({
    id: z.string().min(1),
    creatorId: z.string().min(1),
    conversationType: ConversationTypeSchema,
    createdDate: z.string().datetime({ offset: true }),
    participants: z.array(ParticipantSchema)
  })
  .strict();

export const ConversationRequestSchema = z
  .object({
    recipientID: z.string().min(1),
    conversationType: ConversationTypeSchema
  })
  .strict();

export const MessageSchema = z
  .object({
    id: z.string().min(1),
    creatorId: z.string().min(1),
    conversationId: z.string().min(1),
    content: z.string().min(1),
    createdAt: z.string().datetime({ offset: true })
  })
  .strict();

export const UserSchema = z
  .object({
    userId: z.number().int(),
    userName: z.string().min(1),
    email: z.string().email()
  })
  .strict();

export const WsTicketSchema = z
  .object({
    ticket: z.string().min(1)
  })
  .strict();

export const ConversationsListSchema = z
  .object({
    value: z.array(ConversationSchema)
  })
  .strict();

export const ParticipantsListSchema = z
  .object({
    value: z.array(ParticipantSchema)
  })
  .strict();

export const MessagesListSchema = z
  .object({
    value: z.array(MessageSchema)
  })
  .strict();

export type Conversation = z.infer<typeof ConversationSchema>;
export type ConversationRequest = z.infer<typeof ConversationRequestSchema>;
export type Participant = z.infer<typeof ParticipantSchema>;
export type Message = z.infer<typeof MessageSchema>;
export type User = z.infer<typeof UserSchema>;
export type WsTicket = z.infer<typeof WsTicketSchema>;
