import { Aperture, Lock, TrendingUp } from "lucide-react";
import type { Metadata } from "next";

import { Card, CardContent } from "@/components/ui/card";

export const metadata: Metadata = {
  title: "About GAVIS",
  description: "About GSI AI Vision for Intelligent Surveillance",
};

export default function AboutPage() {
  return (
    <div className="flex flex-col gap-4 md:gap-6">
      {/* Hero Header */}
      <Card className="border-none bg-primary/5 shadow-none">
        <CardContent className="flex flex-col items-center gap-6 p-2 text-center sm:p-4">
          <div className="flex h-20 w-20 items-center justify-center rounded-3xl bg-primary/20">
            <Aperture className="h-10 w-10 text-primary" />
          </div>
          <div className="max-w-2xl space-y-2">
            <h1 className="font-bold text-3xl tracking-tight text-primary sm:text-5xl">About GAVIS</h1>
            <p className="font-medium text-base text-muted-foreground sm:text-xl">
              GSI AI Vision for Intelligent Surveillance
            </p>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:gap-6 lg:grid-cols-3">
        {/* Main Article Content */}
        <Card className="shadow-sm lg:col-span-2">
          <CardContent className="space-y-6 pt-6 text-justify text-base leading-relaxed text-muted-foreground sm:text-[17px]">
            <p>
              <strong className="font-semibold text-foreground">
                GAVIS (GSI AI Vision for Intelligent Surveillance)
              </strong>{" "}
              is an internal innovation initiative developed at Site Wolo, designed as a strategic alternative and
              benchmark to third-party AI-powered CCTV solutions. This project reflects our commitment to building a
              robust, intelligent, and fully controlled surveillance system tailored to the operational needs of
              <strong> PT Gunung Samudera Internasional</strong>.
            </p>
            <p>
              Unlike conventional third-party systems, <strong>GAVIS</strong> is developed internally, providing greater
              flexibility and adaptability. The platform is engineered to align with the specific conditions of our
              operational environments, allowing us to customize features, detection models, and system behavior based
              on real field requirements.
            </p>
            <p>
              One of the key advantages of <strong>GAVIS</strong> is our ongoing effort to learn and develop the
              capability to train and continuously improve our own AI models. This allows the system to become
              increasingly accurate and context-aware over time, especially in recognizing patterns, personnel, and
              activities unique to <strong>PT Gunung Samudera Internasional</strong>.
            </p>
            <p>
              <strong>GAVIS</strong> is not only a technology solution, but also a strategic step toward strengthening
              internal capabilities, enhancing security intelligence, and supporting data-driven decision-making within
              the company.
            </p>

            <div className="mt-8 rounded-2xl border-l-4 border-primary bg-primary/5 p-6 pl-8">
              <p className="font-medium italic text-foreground text-[16px]">
                We present this initiative as a forward-looking solution that prioritizes flexibility, security, and
                sustainability in modern surveillance systems.
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Highlight Cards */}
        <div className="flex flex-col gap-4 md:gap-6">
          <Card className="h-full shadow-sm transition-colors hover:border-primary/20">
            <CardContent className="space-y-3 pt-6">
              <div className="flex items-center gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-blue-500/10 text-blue-500">
                  <Lock className="h-6 w-6" />
                </div>
                <h3 className="font-semibold text-foreground text-lg leading-none">Data Security</h3>
              </div>
              <p className="text-justify text-[15px] leading-relaxed text-muted-foreground">
                Data security is a primary consideration in this initiative. By maintaining full ownership of the
                infrastructure and data processing pipeline, <strong> GAVIS </strong>eliminates dependency on
                third-party providers, thereby reducing exposure to external risks and ensuring that all surveillance
                data remains within the company&apos;s secure ecosystem.
              </p>
            </CardContent>
          </Card>

          <Card className="h-full shadow-sm transition-colors hover:border-primary/20">
            <CardContent className="space-y-3 pt-6">
              <div className="flex items-center gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-500">
                  <TrendingUp className="h-6 w-6" />
                </div>
                <h3 className="font-semibold text-foreground text-lg leading-none">Long-term Efficiency</h3>
              </div>
              <p className="text-justify text-[15px] leading-relaxed text-muted-foreground">
                In addition, this approach provides long-term efficiency in terms of scalability, cost control, and
                system integration. The platform can be extended and enhanced in alignment with future operational
                demands without being constrained by third-party limitations.
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
