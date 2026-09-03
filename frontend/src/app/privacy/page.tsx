import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft, ArrowUpRight } from "lucide-react";

export const metadata: Metadata = {
  title: "Privacy | LivePulse",
  description: "How LivePulse handles account details, saved events, and room messages.",
};

const sections = [
  {
    title: "Information handled",
    content: [
      "Account access is provided by Clerk. LivePulse receives the account identifier and profile information needed to identify you in the application, such as a username or first name when available.",
      "When you send a room message, LivePulse processes the event room identifier, your account identifier, the displayed author name, the message text, and a timestamp.",
      "When you save an event, LivePulse stores the relationship between your account identifier and that event. Event schedules and venue information come from Ticketmaster and are not personal profile data.",
    ],
  },
  {
    title: "How it is used",
    content: [
      "Account information is used to control access to event rooms, associate messages with their sender, and retrieve your saved events.",
      "Connection information is processed to operate WebSocket chat and calculate the number of active connections in a room. The current frontend does not include a separate analytics or advertising SDK.",
    ],
  },
  {
    title: "Storage and retention",
    content: [
      "Room messages are stored temporarily in Upstash Redis and are scheduled for deletion one hour after the event ends.",
      "Account-to-event favorites are stored in Neon PostgreSQL. A favorite can be removed from the event desk or event room, and favorite records are removed when their associated event record is deleted.",
      "Authentication sessions and account profile retention are managed through Clerk and its account controls.",
    ],
  },
  {
    title: "Service providers",
    content: [
      "LivePulse relies on Clerk for authentication, Ticketmaster for event information, Neon for relational data, Upstash for temporary message storage, Vercel for the frontend, and Northflank for backend hosting.",
      "These providers process information as needed to supply their part of the service and apply their own privacy terms to that processing.",
    ],
  },
  {
    title: "Cookies and sessions",
    content: [
      "Clerk uses cookies or similar browser storage to maintain authentication sessions. LivePulse does not add a separate advertising-cookie or analytics-cookie system in the current frontend.",
    ],
  },
  {
    title: "Your choices",
    content: [
      "You can browse event listings without signing in. You can remove saved events at any time, avoid sending room messages, and use the account controls provided by Clerk to manage authentication information.",
      "Do not include private or sensitive information in room messages. Messages are shared with other signed-in people who enter the same event room.",
    ],
  },
];

export default function PrivacyPage() {
  return (
    <div className="mx-auto w-full max-w-[1600px] border-x border-[#45413c]">
      <header className="grid border-b border-[#45413c] lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="p-5 sm:p-8 lg:p-12">
          <Link href="/" className="mb-8 flex w-fit items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#aaa49b] transition-colors hover:text-[#f2efe8]">
            <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
            Home
          </Link>
          <p className="mb-4 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#aaa49b]">Plain-language policy</p>
          <h1 className="font-heading text-[clamp(4.5rem,10vw,9rem)] font-black uppercase leading-[0.82] tracking-[-0.05em] text-[#f2efe8]">Privacy</h1>
        </div>
        <div className="flex flex-col justify-end border-t border-[#45413c] bg-[#f2efe8] p-5 text-[#11100f] sm:p-8 lg:border-t-0 lg:border-l">
          <span className="font-mono text-[9px] font-semibold uppercase tracking-[0.16em] text-[#67625b]">Last updated</span>
          <time dateTime="2026-08-30" className="mt-2 font-heading text-4xl font-bold uppercase tracking-[-0.035em]">August 30, 2026</time>
          <p className="mt-4 text-sm leading-relaxed text-[#67625b]">This page describes the data flows implemented in the current LivePulse application.</p>
        </div>
      </header>

      <div className="grid lg:grid-cols-[300px_minmax(0,1fr)]">
        <aside className="border-b border-[#45413c] p-5 sm:p-8 lg:border-r lg:border-b-0">
          <p className="font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#67625b]">Summary</p>
          <p className="mt-3 max-w-xs text-sm leading-relaxed text-[#aaa49b]">LivePulse uses account data to provide room access, keeps favorites in PostgreSQL, and stores messages temporarily in Redis.</p>
        </aside>

        <div>
          {sections.map(({ title, content }, index) => (
            <section key={title} className="grid border-b border-[#45413c] p-5 sm:grid-cols-[64px_minmax(0,1fr)] sm:gap-4 sm:p-8">
              <span className="frame-number mb-4 font-mono text-[10px] text-[#67625b] sm:mb-0">{String(index + 1).padStart(2, "0")}</span>
              <div>
                <h2 className="font-heading text-3xl font-bold uppercase tracking-[-0.03em] text-[#f2efe8] sm:text-4xl">{title}</h2>
                <div className="mt-5 max-w-3xl space-y-4 text-sm leading-7 text-[#aaa49b] sm:text-base">
                  {content.map(paragraph => <p key={paragraph}>{paragraph}</p>)}
                </div>
              </div>
            </section>
          ))}
        </div>
      </div>

      <section className="grid border-t border-[#45413c] bg-[#f2efe8] text-[#11100f] lg:grid-cols-[1fr_auto]">
        <div className="p-6 sm:p-10">
          <p className="font-heading text-4xl font-black uppercase tracking-[-0.035em]">Questions about this policy?</p>
          <p className="mt-2 max-w-xl text-sm leading-relaxed text-[#67625b]">Use the project repository’s contact options for a privacy question. Do not post account details, message contents, or other personal information in a public issue.</p>
        </div>
        <a href="https://github.com/jrudman25/livepulse" className="flex min-w-[260px] items-center justify-between border-t border-[#11100f] p-6 font-mono text-[10px] font-semibold uppercase tracking-[0.14em] transition-colors hover:bg-[#11100f] hover:text-[#f2efe8] sm:p-10 lg:border-t-0 lg:border-l">
          Project repository
          <ArrowUpRight className="h-4 w-4" aria-hidden="true" />
        </a>
      </section>
    </div>
  );
}
