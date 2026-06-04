import { ApiResponse } from './types';

// Use environment variable if available, otherwise fallback to local gateway
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export class ApiError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

interface ApiOptions extends RequestInit {
  requiresAuth?: boolean;
}

export async function apiFetch<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const { requiresAuth = false, headers, ...restOptions } = options;

  const requestHeaders = new Headers(headers);
  
  if (!requestHeaders.has('Content-Type') && !(options.body instanceof FormData)) {
    requestHeaders.set('Content-Type', 'application/json');
  }

  if (requiresAuth) {
    // In a browser environment, fetch token from localStorage
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('auth_token');
      if (token) {
        requestHeaders.set('Authorization', `Bearer ${token}`);
      }
    }
  }

  const url = `${API_BASE_URL}${path}`;

  try {
    const response = await fetch(url, {
      ...restOptions,
      headers: requestHeaders,
    });

    const data: ApiResponse<T> = await response.json().catch(() => ({ 
      success: false, 
      error: 'Invalid JSON response from server' 
    }));

    if (!response.ok || !data.success) {
      throw new ApiError(data.error || data.message || 'An error occurred while fetching data');
    }

    return data.data;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError(error instanceof Error ? error.message : 'Network error');
  }
}

export function apiGet<T>(path: string, options?: ApiOptions): Promise<T> {
  return apiFetch<T>(path, { ...options, method: 'GET' });
}

export function apiPost<T>(path: string, body?: any, options?: ApiOptions): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  });
}

export function apiPut<T>(path: string, body?: any, options?: ApiOptions): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: 'PUT',
    body: body ? JSON.stringify(body) : undefined,
  });
}

export function apiDelete<T>(path: string, options?: ApiOptions): Promise<T> {
  return apiFetch<T>(path, { ...options, method: 'DELETE' });
}
