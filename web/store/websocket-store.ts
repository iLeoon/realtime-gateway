'use client';

import { create } from 'zustand';

import type { Message } from '@/lib/validation';
import { createWebSocketClient, type OutboundMessage } from '@/websocket/client';

type WebSocketState = {
  socket: WebSocket | null;
  connect: (ticket: string) => void;
  disconnect: () => void;
  sendMessage: (message: OutboundMessage) => void;
  setSocket: (socket: WebSocket | null) => void;
  onInboundMessage?: (event: { message: Message | null; rawText: string | null }) => void;
  setInboundHandler: (handler: (event: { message: Message | null; rawText: string | null }) => void) => void;
};

let wsClient: ReturnType<typeof createWebSocketClient> | null = null;

export const useWebSocketStore = create<WebSocketState>((set, get) => ({
  socket: null,
  connect: (ticket) => {
    if (wsClient) {
      wsClient.disconnect();
    }

    wsClient = createWebSocketClient({
      ticket,
      onOpen: (socket) => set({ socket }),
      onClose: () => set({ socket: null }),
      onMessage: (event) => {
        const handler = get().onInboundMessage;
        if (handler) {
          handler(event);
        }
      }
    });

    wsClient.connect();
  },
  disconnect: () => {
    if (wsClient) {
      wsClient.disconnect();
      wsClient = null;
    }

    set({ socket: null });
  },
  sendMessage: (message) => {
    wsClient?.sendMessage(message);
  },
  setSocket: (socket) => set({ socket }),
  onInboundMessage: undefined,
  setInboundHandler: (handler) => set({ onInboundMessage: handler })
}));
