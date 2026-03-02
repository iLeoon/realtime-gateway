import type { ReactNode } from 'react';

import { AuthBootstrap } from '@/app/(protected)/auth-bootstrap';
import { requireServerAuth } from '@/lib/auth/server';

type ProtectedLayoutProps = {
  children: ReactNode;
};

const ProtectedLayout = async ({ children }: ProtectedLayoutProps) => {
  const token = await requireServerAuth();

  return (
    <>
      <AuthBootstrap token={token} />
      {children}
    </>
  );
};

export default ProtectedLayout;
