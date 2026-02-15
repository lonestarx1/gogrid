import Hero from "@/components/Hero";
import Problem from "@/components/Problem";
import Beliefs from "@/components/Beliefs";
import Architecture from "@/components/Architecture";
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
      <WhyGo />
      <WhoItsFor />
      <Footer />
    </main>
  );
}
