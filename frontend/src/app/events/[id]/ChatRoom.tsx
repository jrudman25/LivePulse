"use client";

import { useWebSocket } from "@/hooks/useWebSocket";
import { useState, useRef, useEffect } from "react";
import { useUser } from "@clerk/nextjs";

export default function ChatRoom({ sessionId }: { sessionId: string }) {
  const { messages, isConnected, sendMessage } = useWebSocket(sessionId);
  const { user } = useUser();
  const [inputBox, setInputBox] = useState("");
  const endOfMessagesRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom of chat feed upon new messages
  useEffect(() => {
    endOfMessagesRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputBox.trim()) return;

    // Fallback hierarchy for OAuth users without usernames
    const authorName = user?.username || user?.firstName || "Anonymous Guest";

    // Relay to the Go backend via WebSocket!
    sendMessage(inputBox, authorName);

    setInputBox("");
  }

  return (
    <div className="flex flex-col h-full">
      {/* Messages Feed */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4 scroll-smooth">
        {!isConnected && (
          <div className="flex h-full items-center justify-center">
            <div className="px-6 py-3 rounded-full bg-fuchsia-500/10 border border-fuchsia-500/20 text-fuchsia-400 text-sm animate-pulse shadow-[0_0_15px_rgba(217,70,239,0.1)]">
              Connecting to LivePulse Arena...
            </div>
          </div>
        )}

        {messages.map((msg, i) => {
          // We determine if this is "my" message to color coordinate bubble tails
          const isMe = msg.user_id === user?.id;

          return (
            <div key={i} className={`flex flex-col ${isMe ? 'items-end' : 'items-start'}`}>
              <span className="text-[10px] uppercase tracking-wider text-slate-500 mb-1 px-2">
                {msg.author_name || `User ${msg.user_id.slice(-5)}`}
              </span>
              <div
                className={`px-4 py-2.5 text-sm/relaxed max-w-[85%] md:max-w-[70%] shadow-lg ${isMe
                    ? 'bg-fuchsia-600 text-white rounded-2xl rounded-br-none shadow-fuchsia-900/50'
                    : 'bg-white/10 text-slate-100 rounded-2xl rounded-bl-none border border-white/5'
                  }`}
              >
                {msg.text}
              </div>
            </div>
          )
        })}
        {/* Invisible div forces scroll tracking */}
        <div ref={endOfMessagesRef} />
      </div>

      {/* Text Input Footer */}
      <div className="p-4 bg-white/5 border-t border-white/10">
        <form onSubmit={handleSend} className="flex gap-3">
          <input
            type="text"
            value={inputBox}
            onChange={(e) => setInputBox(e.target.value)}
            disabled={!isConnected}
            placeholder={isConnected ? "Join the conversation..." : "Waiting for connection..."}
            className="flex-1 bg-black/60 border border-white/20 rounded-full px-6 py-3 text-white placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-fuchsia-500 focus:border-transparent transition-all disabled:opacity-50"
          />
          <button
            type="submit"
            disabled={!inputBox.trim() || !isConnected}
            className="bg-gradient-to-r from-fuchsia-600 to-blue-600 hover:from-fuchsia-500 hover:to-blue-500 disabled:from-slate-800 disabled:to-slate-800 disabled:text-slate-500 disabled:cursor-not-allowed text-white rounded-full px-8 py-3 font-semibold transition-all shadow-[0_0_20px_rgba(217,70,239,0.3)] disabled:shadow-none"
          >
            Send
          </button>
        </form>
      </div>
    </div>
  );
}
