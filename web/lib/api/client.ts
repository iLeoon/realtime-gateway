import axios, { type AxiosError, type AxiosResponse } from 'axios';

import { parseAppError } from '@/lib/api/errors';
import { triggerLogout } from '@/lib/api/auth-events';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:7000/api/v1.0';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  // Backend auth guard reads JWT from cookie; cross-origin XHR must include credentials.
  withCredentials: true
});

apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      triggerLogout();
    }

    return Promise.reject(parseAppError(error));
  }
);
