import type { Metadata, Viewport } from "next";
import { AppShell } from "@/components/AppShell";
import "./globals.css";

export const metadata: Metadata = {
  title: "Domestic Light & Magic",
  description: "Domestic Light & Magic — wire light models and scenes",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
};

const themeInitScript = `
(function(){
  try {
    var root = document.documentElement;
    var t = localStorage.getItem('dlm-theme');
    if (t === 'dark') { root.classList.add('dark'); return; }
    if (t === 'light') { root.classList.remove('dark'); return; }
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  } catch (e) {}
})();`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body>
        <AppShell>{children}</AppShell>
      </body>
    </html>
  );
}
