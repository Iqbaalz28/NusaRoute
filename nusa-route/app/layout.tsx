import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "NusaRoute — Dashboard Pengiriman Paket Indonesia",
  description: "NusaRoute: Aplikasi pengiriman paket terpercaya untuk seluruh Nusantara. Lacak paket, kelola pengiriman, dan pantau operasional secara real-time.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body className={`${inter.variable}`}>
        {children}
      </body>
    </html>
  );
}
