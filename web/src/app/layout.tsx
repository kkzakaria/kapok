import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Kapok Console",
  description: "Kapok BaaS Admin Console",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="font-sans">{children}</body>
    </html>
  );
}
