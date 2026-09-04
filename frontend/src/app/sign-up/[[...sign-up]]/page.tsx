import { SignUp } from "@clerk/nextjs";

export default function Page() {
  return (
    <div className="mx-auto grid min-h-[calc(100vh-72px)] w-full max-w-[1600px] border-x border-[#45413c] lg:grid-cols-[minmax(0,1fr)_560px]">
      <div className="broadcast-grid hidden flex-col justify-between border-r border-[#45413c] p-12 lg:flex">
        <span className="font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#aaa49b]">Access control / 02</span>
        <div>
          <p className="mb-4 flex items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#ed2f24]"><span className="h-2.5 w-2.5 bg-[#ed2f24]" />New operator</p>
          <h1 className="font-heading text-8xl font-black uppercase leading-[0.82] tracking-[-0.045em] text-[#f2efe8]">Join an<br />event<br />room.</h1>
        </div>
        <p className="max-w-md text-sm leading-relaxed text-[#aaa49b]">Create an account to save events and take part in their conversations.</p>
      </div>
      <div className="flex items-center justify-center bg-[#f2efe8] p-5 text-[#11100f] sm:p-10">
        <SignUp
          appearance={{
            elements: {
              rootBox: "w-full max-w-md mx-auto",
              card: "w-full rounded-none border border-[#11100f] bg-[#f2efe8] shadow-none",
              headerTitle: "font-heading text-4xl uppercase tracking-[-0.035em] text-[#11100f]",
              headerSubtitle: "text-[#67625b]",
              socialButtonsBlockButton: "rounded-none border-[#8c877f] bg-transparent text-[#11100f] hover:bg-[#dedad1]",
              socialButtonsBlockButtonText: "font-medium text-[#11100f]",
              dividerLine: "bg-[#8c877f]",
              dividerText: "font-mono text-[10px] uppercase tracking-[0.12em] text-[#67625b]",
              formFieldLabel: "font-mono text-[10px] uppercase tracking-[0.12em] text-[#11100f]",
              formFieldInput: "rounded-none border-[#8c877f] bg-transparent text-[#11100f] focus:border-[#11100f] focus:ring-0",
              formButtonPrimary: "rounded-none bg-[#ed2f24] font-mono text-[11px] uppercase tracking-[0.14em] text-white hover:bg-[#11100f]",
              footerActionText: "text-[#67625b]",
              footerActionLink: "font-semibold text-[#11100f] hover:text-[#ed2f24]"
            }
          }}
        />
      </div>
    </div>
  );
}
