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
  | { type: "milestone_achieved"; milestone: any; achieved_at: string }
  | { type: "error"; message: string };

export function useWebSocket(sessionId: string) {
  const { getToken } = useAuth();
  const wsRef = useRef<WebSocket | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const reconnectAttempt = useRef(0);

  useEffect(() => {
    if (!sessionId) {return;}
    
    let isMounted = true;
    let reconnectTimeoutId: NodeJS.Timeout;

    const connect = async () => {
      try {
        const token = await getToken();
        if (!token) {return;}

        // Ensure this points to the external Go server address dynamically
        const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
        // Token strictly removed from URL parameters intrinsically blocking Leakage
        const url = `${WS_URL}/ws?session_id=${sessionId}`;
        const ws = new WebSocket(url);

        ws.onopen = () => {
          if (isMounted) {
            setIsConnected(true);
            reconnectAttempt.current = 0; // Reset intrinsically on success
          }
          // Authenticate autonomously as absolutely first action over encrypted Socket
          ws.send(JSON.stringify({ type: "authenticate", token }));
        };

        ws.onclose = () => {
          if (isMounted) {
            setIsConnected(false);
            // Exponential backoff structurally clamped to 30 second maximum ceilings natively
            const delay = Math.min(1000 * Math.pow(2, reconnectAttempt.current), 30000);
            reconnectAttempt.current += 1;
            console.log(`WebSocket connection dropped gracefully. Attempting retry #${reconnectAttempt.current} natively in ${delay}ms`);
            reconnectTimeoutId = setTimeout(connect, delay);
          }
        };

        ws.onmessage = (event) => {
          try {
            const data: WSEvent = JSON.parse(event.data);
            if (data.type === "chat") {
              setMessages((prev) => {
                if (prev.some(m => m.id === data.message.id)) {return prev;}
                return [...prev, data.message];
              });
            } else if (data.type === "error") {
              // Immediately flag the React toast interface natively
              setErrorMsg(data.message);
              setTimeout(() => setErrorMsg(null), 5000); // clear gracefully
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
      clearTimeout(reconnectTimeoutId); // Clean structurally
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

  const clearError = useCallback(() => setErrorMsg(null), []);

  return { messages, isConnected, sendMessage, sendReaction, setMessages, errorMsg, clearError };
}
