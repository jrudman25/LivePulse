import { useEffect, useRef, useState, useCallback } from "react";
import { useAuth } from "@clerk/nextjs";

export type ChatMessage = {
  id: string;
  user_id: string;
  session_id: string;
  text: string;
  author_name: string;
  timestamp: string;
};

export type WSEvent = 
  | { type: "chat"; message: ChatMessage }
  | { type: "reaction"; user_id: string; reaction_type: string; timestamp: string }
  | { type: "milestone_achieved"; milestone: any; achieved_at: string };

export function useWebSocket(sessionId: string) {
  const { getToken } = useAuth();
  const wsRef = useRef<WebSocket | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!sessionId) return;
    
    let isMounted = true;

    const connect = async () => {
      try {
        const token = await getToken();
        if (!token) return;

        // Ensure this points to the external Go server address (localhost:8080 during dev)
        const url = `ws://localhost:8080/ws?session_id=${sessionId}&token=${token}`;
        const ws = new WebSocket(url);

        ws.onopen = () => {
          if (isMounted) setIsConnected(true);
        };

        ws.onclose = () => {
          if (isMounted) setIsConnected(false);
          // Future feature: exponential backoff reconnect logic
        };

        ws.onmessage = (event) => {
          try {
            const data: WSEvent = JSON.parse(event.data);
            if (data.type === "chat") {
              // Append newly broadcasted messages, preventing React Strict Mode duplicates
              setMessages((prev) => {
                if (prev.some(m => m.id === data.message.id)) return prev;
                return [...prev, data.message];
              });
            }
          } catch (e) {
            console.error("Failed to parse WS message", e);
          }
        };

        wsRef.current = ws;
      } catch (err) {
        console.error("WS connect error:", err);
      }
    };

    connect();

    return () => {
      isMounted = false;
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [sessionId, getToken]);

  const sendMessage = useCallback((text: string, authorName: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "chat", text, author_name: authorName }));
    }
  }, []);

  const sendReaction = useCallback((reactionType: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "reaction", reaction_type: reactionType }));
    }
  }, []);

  return { messages, isConnected, sendMessage, sendReaction, setMessages };
}
