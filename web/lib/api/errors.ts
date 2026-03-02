import axios from 'axios';

import { ErrorResponseSchema } from '@/lib/validation';
import type { AppError } from '@/types/app-error';

export const parseAppError = (error: unknown): AppError => {
  if (!axios.isAxiosError(error)) {
    return {
      code: 'Unknown',
      message: error instanceof Error ? error.message : 'Unknown error'
    };
  }

  const status = error.response?.status;
  const parsed = ErrorResponseSchema.safeParse(error.response?.data);

  if (parsed.success) {
    return {
      code: parsed.data.error.code,
      message: parsed.data.error.message,
      target: parsed.data.error.target,
      innerCode: parsed.data.error.innererror?.innererror?.code ?? parsed.data.error.innererror?.code,
      status
    };
  }

  return {
    code: `HTTP_${status ?? 0}`,
    message: error.message,
    status
  };
};
