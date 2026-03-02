import { apiClient } from '@/lib/api/client';
import { parseOrThrow } from '@/lib/api/validation';
import {
  ConversationsListSchema,
  ConversationRequestSchema,
  ConversationSchema,
  MessagesListSchema,
  ParticipantsListSchema,
  UserSchema,
  WsTicketSchema,
  type Conversation,
  type ConversationRequest,
  type Message,
  type Participant,
  type User,
  type WsTicket
} from '@/lib/validation';

export const chatApi = {
  getConversations: async (): Promise<Conversation[]> => {
    const response = await apiClient.get('/conversations');
    return parseOrThrow(ConversationsListSchema, response.data, 'conversations').value;
  },

  createConversation: async (payload: ConversationRequest): Promise<Conversation> => {
    const body = parseOrThrow(ConversationRequestSchema, payload, 'conversation request');
    const response = await apiClient.post('/conversations', body);
    return parseOrThrow(ConversationSchema, response.data, 'conversation');
  },

  getConversation: async (id: string): Promise<Conversation> => {
    const response = await apiClient.get(`/conversations/${id}`);
    return parseOrThrow(ConversationSchema, response.data, 'conversation');
  },

  getMessages: async (conversationId: string): Promise<Message[]> => {
    const response = await apiClient.get(`/conversations/${conversationId}/messages`);
    return parseOrThrow(MessagesListSchema, response.data, 'messages').value;
  },

  getParticipants: async (conversationId: string): Promise<Participant[]> => {
    const response = await apiClient.get(`/conversations/${conversationId}/participants`);
    return parseOrThrow(ParticipantsListSchema, response.data, 'participants').value;
  },

  getUser: async (id: string): Promise<User> => {
    const response = await apiClient.get(`/users/${id}`);
    return parseOrThrow(UserSchema, response.data, 'user');
  },

  getWebSocketTicket: async (): Promise<WsTicket> => {
    const response = await apiClient.get('/ws/token');
    return parseOrThrow(WsTicketSchema, response.data, 'websocket ticket');
  }
};
