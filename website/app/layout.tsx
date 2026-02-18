import type { Metadata } from "next";
import { JetBrains_Mono, Inter } from "next/font/google";
import Navbar from "@/components/Navbar";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: "GoGrid — Develop and Orchestrate AI Agents",
  description:
    "A unified system for developing and orchestrating production AI agents in Go. Built for infrastructure engineers, not notebook demos.",
  icons: {
    icon: "/favicon.svg",
  },
  openGraph: {
    title: "GoGrid — Develop and Orchestrate AI Agents",
    description:
      "A unified system for developing and orchestrating production AI agents in Go. Built for infrastructure engineers, not notebook demos.",
    url: "https://gogrid.org",
    siteName: "GoGrid",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${inter.variable} ${jetbrainsMono.variable}`}>
      <body className="antialiased">
        <Navbar />
        {children}
      </body>
    </html>
  );
}
