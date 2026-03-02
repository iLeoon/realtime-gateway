import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

export const requireServerAuth = async (): Promise<string> => {
  const cookieStore = await cookies();
  const token = cookieStore.get('token')?.value ?? cookieStore.get('auth_token')?.value;

  if (!token) {
    redirect('/login');
  }

  return token;
};
