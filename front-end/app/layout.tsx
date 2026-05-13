import type { Metadata } from "next";
import { Space_Grotesk, DM_Sans } from "next/font/google";
import "./globals.css";

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  weight: ["600", "700"],
  variable: "--font-heading",
});

const dmSans = DM_Sans({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-body",
});

export const metadata: Metadata = {
  title: "NusaRoute — Solusi Pengiriman Nusantara.",
  description: "NusaRoute: Platform logistik modern untuk pelacakan andal dan harga transparan. Pengiriman yang disederhanakan untuk Nusantara.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" className={`${spaceGrotesk.variable} ${dmSans.variable}`}>
      <body className="font-body antialiased min-h-screen bg-background text-text">
        {children}
      </body>
    </html>
  );
}
