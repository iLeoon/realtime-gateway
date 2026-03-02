import type { Metadata } from 'next';
import type { ReactNode } from 'react';

import { ReactQueryProvider } from '@/providers/react-query-provider';

import '@/app/globals.css';

export const metadata: Metadata = {
  title: 'Realtime Chat',
  description: 'Production-grade Next.js frontend for chat system'
};

type RootLayoutProps = {
  children: ReactNode;
};

const RootLayout = ({ children }: RootLayoutProps) => (
  <html lang="en" className="dark">
    <body className="min-h-screen bg-background text-foreground">
      <ReactQueryProvider>{children}</ReactQueryProvider>
    </body>
  </html>
);

export default RootLayout;
