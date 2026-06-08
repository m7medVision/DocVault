import { Hero } from "./_components/Hero";
import { Features } from "./_components/Features";
import { HowItWorks } from "./_components/HowItWorks";
import { Pricing } from "./_components/Pricing";
import { Testimonials } from "./_components/Testimonials";
import { Faq } from "./_components/Faq";

export default function MarketingPage() {
  return (
    <>
      <Hero />
      <Features />
      <HowItWorks />
      <Pricing />
      <Testimonials />
      <Faq />
    </>
  );
}
