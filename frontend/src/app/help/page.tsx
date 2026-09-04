import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft, ArrowUpRight } from "lucide-react";

export const metadata: Metadata = {
  title: "Help & FAQ | LivePulse",
  description: "Answers about finding events, joining rooms, chat, and saved events on LivePulse.",
};

const questions = [
  {
    question: "What is LivePulse?",
    answer: "LivePulse organizes group chats around individual concerts, sports, and shows. Each listed event has its own room, so the conversation stays connected to what people are watching.",
  },
  {
    question: "Where do the event listings come from?",
    answer: "Event details are sourced from Ticketmaster. LivePulse focuses on events scheduled within the next 24 hours, and availability depends on the information Ticketmaster provides.",
  },
  {
    question: "Do I need an account?",
    answer: "You can browse and search events without an account. A LivePulse account is required to save an event, read its room, or send messages.",
  },
  {
    question: "Why can’t I find an event?",
    answer: "Check the spelling of the event title and try a shorter search. An event may also be outside the 24-hour window, unavailable through Ticketmaster, or not yet included in the current results.",
  },
  {
    question: "Who can see my messages?",
    answer: "Messages are visible to signed-in people in the same event room. Your selected Clerk username or first name appears with the message when available.",
  },
  {
    question: "How long are chat messages kept?",
    answer: "Room messages are scheduled for deletion one hour after the event ends. They are stored temporarily so people joining the room can follow the conversation.",
  },
  {
    question: "What does the room count mean?",
    answer: "The count reports active WebSocket connections for that event room. It can change as people connect, disconnect, or open the room on more than one device.",
  },
  {
    question: "What should I do if a room will not connect?",
    answer: "Confirm that you are signed in and that your internet connection is working, then reload the page. Network restrictions or a temporary backend interruption can prevent the WebSocket connection from opening.",
  },
];

export default function HelpPage() {
  return (
    <div className="mx-auto w-full max-w-[1600px] border-x border-[#45413c]">
      <header className="grid border-b border-[#45413c] lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="p-5 sm:p-8 lg:p-12">
          <Link href="/" className="mb-8 flex w-fit items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#aaa49b] transition-colors hover:text-[#f2efe8]">
            <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
            Home
          </Link>
          <p className="mb-4 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#aaa49b]">Product guide</p>
          <h1 className="font-heading text-[clamp(4.5rem,10vw,9rem)] font-black uppercase leading-[0.92] tracking-[-0.015em] text-[#f2efe8]">Help & FAQ</h1>
        </div>
        <div className="flex flex-col justify-end border-t border-[#45413c] bg-[#f2efe8] p-5 text-[#11100f] sm:p-8 lg:border-t-0 lg:border-l">
          <p className="font-heading text-4xl font-bold uppercase leading-none tracking-[-0.035em]">Start with an event</p>
          <p className="mt-3 text-sm leading-relaxed text-[#67625b]">Browse the upcoming schedule, choose a room, and sign in when you want to join the conversation.</p>
          <Link href="/events" className="mt-6 flex items-center justify-between border-t border-[#11100f] pt-4 font-mono text-[10px] font-semibold uppercase tracking-[0.14em]">
            Browse events
            <ArrowUpRight className="h-4 w-4" aria-hidden="true" />
          </Link>
        </div>
      </header>

      <div className="grid lg:grid-cols-[300px_minmax(0,1fr)]">
        <aside className="border-b border-[#45413c] p-5 sm:p-8 lg:border-r lg:border-b-0">
          <p className="font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#67625b]">Common questions</p>
          <p className="mt-3 max-w-xs text-sm leading-relaxed text-[#aaa49b]">Straightforward answers about how the current product works.</p>
        </aside>
        <section>
          {questions.map(({ question, answer }, index) => (
            <details key={question} className="group border-b border-[#45413c] bg-[#11100f] open:bg-[#171614]">
              <summary className="grid cursor-pointer list-none grid-cols-[48px_minmax(0,1fr)_32px] items-center gap-3 p-5 marker:hidden sm:grid-cols-[64px_minmax(0,1fr)_48px] sm:p-7 [&::-webkit-details-marker]:hidden">
                <span className="frame-number font-mono text-[10px] text-[#67625b]">{String(index + 1).padStart(2, "0")}</span>
                <h2 className="font-heading text-2xl font-bold uppercase leading-tight tracking-[-0.025em] text-[#f2efe8] sm:text-3xl">{question}</h2>
                <span className="text-right font-heading text-3xl font-light text-[#aaa49b] group-open:hidden">+</span>
                <span className="hidden text-right font-heading text-3xl font-light text-[#ed2f24] group-open:block">−</span>
              </summary>
              <div className="grid grid-cols-[48px_minmax(0,1fr)_32px] gap-3 px-5 pb-7 sm:grid-cols-[64px_minmax(0,1fr)_48px] sm:px-7 sm:pb-8">
                <span />
                <p className="max-w-2xl text-sm leading-7 text-[#aaa49b] sm:text-base">{answer}</p>
              </div>
            </details>
          ))}
        </section>
      </div>

      <section className="grid border-t border-[#45413c] bg-[#ed2f24] text-[#fffaf2] lg:grid-cols-[1fr_auto]">
        <div className="p-6 sm:p-10">
          <p className="font-heading text-4xl font-black uppercase tracking-[-0.035em]">Found a technical issue?</p>
          <p className="mt-2 max-w-xl text-sm leading-relaxed text-[#fffaf2]/80">Open an issue in the project repository with the page, browser, and steps that produced the problem. Do not include account credentials or other sensitive information.</p>
        </div>
        <a href="https://github.com/jrudman25/livepulse/issues" className="flex min-w-[260px] items-center justify-between border-t border-[#fffaf2]/40 p-6 font-mono text-[10px] font-semibold uppercase tracking-[0.14em] transition-colors hover:bg-[#11100f] sm:p-10 lg:border-t-0 lg:border-l">
          Open GitHub issues
          <ArrowUpRight className="h-4 w-4" aria-hidden="true" />
        </a>
      </section>
    </div>
  );
}
