export const decodeJwtSub = (token: string | null): string | null => {
  if (!token) {
    return null;
  }

  const parts = token.split('.');
  if (parts.length < 2) {
    return null;
  }

  const payload = parts[1];
  if (!payload) {
    return null;
  }

  try {
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
    const padded = normalized + '='.repeat((4 - (normalized.length % 4)) % 4);
    const decoded = JSON.parse(atob(padded)) as { sub?: unknown };

    return typeof decoded.sub === 'string' && decoded.sub.length > 0 ? decoded.sub : null;
  } catch {
    return null;
  }
};
