import { SignIn } from "@clerk/nextjs";

export default function Page() {
  return (
    <div className="flex min-h-[calc(100vh-80px)] items-center justify-center p-4">
      <SignIn 
        appearance={{ 
          elements: { 
            rootBox: "mx-auto shadow-[0_0_40px_rgba(217,70,239,0.15)] rounded-2xl",
            card: "bg-black/80 backdrop-blur-xl border border-white/10",
            headerTitle: "text-white",
            headerSubtitle: "text-slate-400",
            socialButtonsBlockButton: "border-white/10 text-white bg-white/5 hover:bg-white/10",
            socialButtonsBlockButtonText: "text-white",
            dividerLine: "bg-white/10",
            dividerText: "text-slate-500",
            formFieldLabel: "text-slate-300",
            formFieldInput: "bg-black border-white/20 text-white focus:border-fuchsia-500 focus:ring-fuchsia-500/20",
            formButtonPrimary: "bg-gradient-to-r from-fuchsia-600 to-blue-600 hover:from-fuchsia-500 hover:to-blue-500 border-none",
            footerActionText: "text-slate-400",
            footerActionLink: "text-fuchsia-400 hover:text-fuchsia-300"
          } 
        }} 
      />
    </div>
  );
}
