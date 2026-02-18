import Hero from "@/components/Hero";
import Problem from "@/components/Problem";
import Beliefs from "@/components/Beliefs";
import Architecture from "@/components/Architecture";
import TerminalDemo from "@/components/TerminalDemo";
import WhyGo from "@/components/WhyGo";
import WhoItsFor from "@/components/WhoItsFor";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <main>
      <Hero />
      <Problem />
      <Beliefs />
      <Architecture />
      <TerminalDemo />
      <WhyGo />
      <WhoItsFor />
      <Footer />
    </main>
  );
}
