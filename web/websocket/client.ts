import { chatApi } from '@/lib/api/chat';
import { MessageSchema, type Message } from '@/lib/validation';

type WebSocketClientConfig = {
  ticket: string;
  onOpen: (socket: WebSocket) => void;
  onClose: () => void;
  onMessage: (event: { message: Message | null; rawText: string | null }) => void;
};

export type OutboundMessage = {
  type: 'send_message';
  payload: {
    content: string;
    conversationID: string;
    recipientUserID: string;
  };
};

type ConnectOptions = {
  reconnectAttempt: number;
};

const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_BASE_URL ?? 'ws://localhost:7000/ws';
const MAX_RECONNECT_ATTEMPTS = 5;

const parseInboundMessage = (data: unknown): Message | null => {
  const parsed = MessageSchema.safeParse(data);
  return parsed.success ? parsed.data : null;
};

export const createWebSocketClient = (config: WebSocketClientConfig) => {
  let socket: WebSocket | null = null;
  let closedManually = false;
  let reconnectAttempt = 0;
  let initialTicket = config.ticket;

  const getReconnectDelay = (attempt: number): number => {
    const baseDelay = 500;
    const jitter = Math.floor(Math.random() * 200);
    return Math.min(baseDelay * 2 ** attempt + jitter, 8000);
  };

  const fetchFreshTicket = async (): Promise<string> => {
    const ticketResponse = await chatApi.getWebSocketTicket();
    return ticketResponse.ticket;
  };

  const connectInternal = async (options: ConnectOptions): Promise<void> => {
    const ticket = options.reconnectAttempt === 0 ? initialTicket : await fetchFreshTicket();

    const url = new URL(WS_BASE_URL);
    // Backend websocket guard expects `token` query parameter.
    url.searchParams.set('token', ticket);

    socket = new WebSocket(url.toString());

    socket.onopen = () => {
      reconnectAttempt = 0;
      config.onOpen(socket as WebSocket);
    };

    socket.onmessage = (event) => {
      const raw = event.data as string;
      let inbound: Message | null = null;

      try {
        const parsedJson: unknown = JSON.parse(raw);
        inbound = parseInboundMessage(parsedJson);
      } catch {
        inbound = null;
      }

      // Backend may send plain text frames for message acks/broadcasts.
      config.onMessage({ message: inbound, rawText: inbound ? null : raw });
    };

    socket.onclose = () => {
      config.onClose();

      if (closedManually) {
        return;
      }

      if (reconnectAttempt >= MAX_RECONNECT_ATTEMPTS) {
        return;
      }

      reconnectAttempt += 1;
      const delay = getReconnectDelay(reconnectAttempt);

      window.setTimeout(() => {
        void connectInternal({ reconnectAttempt });
      }, delay);
    };
  };

  return {
    connect: () => {
      closedManually = false;
      reconnectAttempt = 0;
      void connectInternal({ reconnectAttempt: 0 });
    },
    disconnect: () => {
      closedManually = true;
      socket?.close();
      socket = null;
    },
    sendMessage: (message: OutboundMessage) => {
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        return;
      }

      socket.send(JSON.stringify(message));
    }
  };
};
