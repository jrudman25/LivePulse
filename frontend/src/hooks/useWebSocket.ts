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
  | { type: "authenticated" }
  | { type: "chat"; message: ChatMessage }
  | { type: "reaction"; user_id: string; reaction_type: string; timestamp: string }
  | { type: "milestone_achieved"; milestone: unknown; achieved_at: string }
  | { type: "error"; message: string };

const MAX_RECONNECT_ATTEMPTS = 6;

export function useWebSocket(sessionId: string) {
  const { getToken } = useAuth();
  const wsRef = useRef<WebSocket | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [hasFailed, setHasFailed] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const reconnectAttempt = useRef(0);
  const [connectNonce, setConnectNonce] = useState(0);

  useEffect(() => {
    if (!sessionId) {return;}

    let isMounted = true;
    let reconnectTimeoutId: ReturnType<typeof setTimeout> | undefined;
    let errorTimeoutId: ReturnType<typeof setTimeout> | undefined;

    const handleEvent = (data: WSEvent) => {
      if (data.type === "chat") {
        setMessages((prev) => {
          if (prev.some(m => m.id === data.message.id)) {return prev;}
          return [...prev, data.message];
        });
      } else if (data.type === "authenticated") {
        setIsAuthenticated(true);
        reconnectAttempt.current = 0;
      } else if (data.type === "error") {
        // Immediately flag the React toast interface natively
        setErrorMsg(data.message);
        clearTimeout(errorTimeoutId);
        errorTimeoutId = setTimeout(() => setErrorMsg(null), 5000); // clear gracefully
      }
    };

    const connect = async () => {
      const scheduleReconnect = () => {
        if (!isMounted) {return;}
        if (reconnectAttempt.current >= MAX_RECONNECT_ATTEMPTS) {
          setHasFailed(true);
          return;
        }
        // Exponential backoff structurally clamped to 30 second maximum ceilings natively
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempt.current), 30000);
        reconnectAttempt.current += 1;
        console.warn(`WebSocket disconnected. Retrying in ${delay}ms (attempt ${reconnectAttempt.current}).`);
        reconnectTimeoutId = setTimeout(connect, delay);
      };

      try {
        const token = await getToken();
        if (!isMounted) {return;}
        if (!token) {
          scheduleReconnect();
          return;
        }

        // Ensure this points to the external Go server address dynamically
        const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
        // Token strictly removed from URL parameters intrinsically blocking Leakage
        const url = `${WS_URL}/ws?session_id=${encodeURIComponent(sessionId)}`;
        const ws = new WebSocket(url);

        ws.onopen = () => {
          if (!isMounted) {return;}
          setIsConnected(true);
          // Authenticate autonomously as absolutely first action over encrypted Socket
          ws.send(JSON.stringify({ type: "authenticate", token }));
        };

        ws.onclose = () => {
          if (!isMounted) {return;}
          setIsConnected(false);
          setIsAuthenticated(false);
          scheduleReconnect();
        };

        ws.onmessage = (event) => {
          if (!isMounted || typeof event.data !== "string") {return;}
          // The Go writePump can batch several queued JSON documents into one
          // frame separated by newlines, so each line is parsed independently.
          for (const line of event.data.split("\n")) {
            const trimmed = line.trim();
            if (!trimmed) {continue;}
            try {
              handleEvent(JSON.parse(trimmed) as WSEvent);
            } catch (e) {
              console.error("Failed to parse WS message", e);
            }
          }
        };

        wsRef.current = ws;
      } catch (err) {
        console.error("WS connect error:", err);
        scheduleReconnect();
      }
    };

    connect();

    return () => {
      isMounted = false;
      clearTimeout(reconnectTimeoutId); // Clean structurally
      clearTimeout(errorTimeoutId);
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [sessionId, getToken, connectNonce]);

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

  const retry = useCallback(() => {
    reconnectAttempt.current = 0;
    setHasFailed(false);
    setConnectNonce(n => n + 1);
  }, []);

  const clearError = useCallback(() => setErrorMsg(null), []);

  return { messages, isConnected, isAuthenticated, hasFailed, retry, sendMessage, sendReaction, setMessages, errorMsg, clearError };
}
