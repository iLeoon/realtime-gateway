import type { ZodType } from 'zod';

import { SchemaValidationError } from '@/types/app-error';

export const parseOrThrow = <T>(
  schema: ZodType<T>,
  data: unknown,
  entity: string
): T => {
  const parsed = schema.safeParse(data);

  if (!parsed.success) {
    const issues = parsed.error.issues.map((issue) => `${issue.path.join('.') || '<root>'}: ${issue.message}`);
    throw new SchemaValidationError(entity, issues);
  }

  return parsed.data;
};
