import axios, { AxiosError, AxiosInstance } from 'axios';
import type { ApiError } from '@/types/api';
import { getStoredValue } from '@/store/tokenStorage';

const BASE_URL = process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://localhost:8080/v1';

export const apiClient: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 15_000,
  headers: { 'Content-Type': 'application/json' },
});

// ── Auth token injection ──────────────────────────────────────────────────────
apiClient.interceptors.request.use(async (config) => {
  const token = await getStoredValue('auth_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// ── Uniform error shaping ─────────────────────────────────────────────────────
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiError>) => {
    const apiErr = error.response?.data;
    if (apiErr?.error_code) return Promise.reject(apiErr);
    // Network / timeout errors
    return Promise.reject({
      error_code: 'NETWORK_ERROR',
      message: error.message ?? 'Network request failed',
    } satisfies ApiError);
  },
);

export const WS_BASE_URL = process.env.EXPO_PUBLIC_WS_BASE_URL ?? 'ws://localhost:8080';

export async function getAuthToken(): Promise<string | null> {
  return getStoredValue('auth_token');
}
