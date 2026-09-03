"use client";

import { useWebSocket } from "@/hooks/useWebSocket";
import { useState, useRef, useEffect } from "react";
import { Send, X } from "lucide-react";
import { useUser } from "@clerk/nextjs";

export default function ChatRoom({ sessionId }: { sessionId: string }) {
  const { user } = useUser();
  const [inputBox, setInputBox] = useState("");
  const endOfMessagesRef = useRef<HTMLDivElement>(null);
  const { messages, isConnected, sendMessage, errorMsg, clearError } =
    useWebSocket(sessionId);

  const maxChars = 500;
  const charsRemaining = maxChars - inputBox.length;
  const isOverLimit = charsRemaining < 0;

  // Auto-scroll to bottom of chat feed upon new messages
  useEffect(() => {
    endOfMessagesRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputBox.trim() || isOverLimit) {return;}

    // Fallback hierarchy for OAuth users without usernames
    const authorName = user?.username || user?.firstName || "Anonymous Guest";

    // Relay to the Go backend via WebSocket!
    sendMessage(inputBox, authorName);

    setInputBox("");
  };

  return (
    <div className="flex h-full min-h-[620px] flex-col">
      <div className="flex h-14 shrink-0 items-center justify-between border-b border-[#45413c] px-4 sm:px-6">
        <div className="flex items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#f2efe8]">
          <span className={`h-2 w-2 ${isConnected ? "animate-pulse bg-[#ed2f24]" : "bg-[#67625b]"}`} />
          {isConnected ? "Connected" : "Connecting"}
        </div>
        <span className="frame-number font-mono text-[9px] uppercase tracking-[0.12em] text-[#67625b]">{messages.length} messages / {sessionId.slice(0, 8)}</span>
      </div>

      {/* Messages Feed */}
      <div className="flex-1 overflow-y-auto scroll-smooth">
        {!isConnected && (
          <div className="grid h-full min-h-[400px] place-items-center">
            <div className="border border-[#45413c] px-6 py-4 font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#aaa49b]">
              Connecting to the room
            </div>
          </div>
        )}

        {isConnected && messages.length === 0 && (
          <div className="grid h-full min-h-[400px] place-items-center p-8 text-center">
            <div>
              <p className="font-heading text-4xl font-bold uppercase tracking-[-0.03em] text-[#f2efe8]">You are on the wire</p>
              <p className="mt-2 text-sm text-[#aaa49b]">No messages yet. Start the room.</p>
            </div>
          </div>
        )}

        {messages.map((msg, i) => {
          // We determine if this is "my" message to color coordinate bubble tails
          const isMe = msg.user_id === user?.id;

          return (
            <div key={i} className={`grid border-b border-[#302e2a] transition-colors hover:bg-[#1d1b19] sm:grid-cols-[150px_minmax(0,1fr)] ${isMe ? "bg-[#1b1715]" : ""}`}>
              <div className="flex items-center justify-between gap-3 border-b border-[#302e2a] px-4 py-3 sm:block sm:border-r sm:border-b-0 sm:px-5 sm:py-4">
                <span className={`block truncate font-mono text-[10px] font-semibold uppercase tracking-[0.1em] ${isMe ? "text-[#ed2f24]" : "text-[#aaa49b]"}`}>
                  {isMe ? "You" : msg.author_name || `User ${msg.user_id.slice(-5)}`}
                </span>
                <span className="frame-number mt-1 block font-mono text-[9px] text-[#67625b]">#{String(i + 1).padStart(3, "0")}</span>
              </div>
              <div className="px-4 py-4 text-sm leading-relaxed text-[#e2ddd4] sm:px-6 sm:py-5">
                {msg.text}
              </div>
            </div>
          );
        })}
        {/* Invisible div forces scroll tracking */}
        <div ref={endOfMessagesRef} />
      </div>

      {/* Text Input Footer */}
      <div className="shrink-0 border-t border-[#67625b] bg-[#11100f] p-3 sm:p-4">
        {/* Toast Error Banner */}
        {errorMsg && (
          <div className="mb-3 flex items-center justify-between border border-[#ed2f24] bg-[#2c1513] px-4 py-3 font-mono text-[10px] font-semibold uppercase tracking-[0.1em] text-[#ff8d86]">
            <span>{errorMsg}</span>
            <button onClick={clearError} className="p-1 transition-colors hover:text-white" aria-label="Dismiss error"><X className="h-4 w-4" aria-hidden="true" /></button>
          </div>
        )}

        <form onSubmit={handleSend} className="grid grid-cols-[minmax(0,1fr)_auto] border border-[#67625b] focus-within:border-[#f2efe8]">
          <input
            type="text"
            value={inputBox}
            onChange={(e) => setInputBox(e.target.value)}
            disabled={!isConnected}
            placeholder={isConnected ? "Write to the room" : "Waiting to connect"}
            className={`min-w-0 bg-transparent px-4 py-4 text-sm text-[#f2efe8] placeholder:text-[#67625b] focus:outline-none disabled:opacity-50 sm:px-5 ${isOverLimit ? "text-[#ff8d86]" : ""}`}
          />
          <button type="submit" disabled={!inputBox.trim() || !isConnected || isOverLimit} className="flex items-center gap-3 bg-[#ed2f24] px-4 font-mono text-[10px] font-semibold uppercase tracking-[0.14em] text-[#fffaf2] transition-colors hover:bg-[#f2efe8] hover:text-[#11100f] disabled:cursor-not-allowed disabled:bg-[#302e2a] disabled:text-[#67625b] sm:px-6">
            <span className="hidden sm:inline">Send</span>
            <Send className="h-4 w-4" aria-hidden="true" />
          </button>
        </form>

        {/* Proactive Character Limit UX */}
        <div className="mt-2 flex items-center justify-between px-1 font-mono text-[9px] uppercase tracking-[0.12em]">
          <span className="text-[#67625b]">Messages are visible to everyone in this room</span>
          <span className={isOverLimit ? "text-[#ff5148]" : charsRemaining < 50 ? "text-[#f2efe8]" : "text-[#67625b]"}>{inputBox.length} / 500</span>
        </div>
      </div>
    </div>
  );
}
