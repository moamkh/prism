import axios from 'axios';

const API_BASE = (import.meta as any).env?.VITE_API_BASE_URL || '/api/v1';

const client = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Toast integration
let toastHandler: ((message: string, type: 'success' | 'error' | 'info') => void) | null = null;

export function registerToastHandler(
  handler: (message: string, type: 'success' | 'error' | 'info') => void
) {
  toastHandler = handler;
}

function showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
  if (toastHandler) toastHandler(message, type);
}

// Response interceptor for automatic toast notifications
client.interceptors.response.use(
  (response) => {
    // Show success toast for mutating operations
    const method = response.config.method?.toLowerCase();
    if (method === 'post' || method === 'put' || method === 'delete') {
      const url = response.config.url || '';
      let action = 'Operation';
      if (method === 'post') action = 'Created';
      if (method === 'put') action = 'Updated';
      if (method === 'delete') action = 'Deleted';
      // Extract resource name from URL
      const resource = url.split('/')[1] || 'resource';
      if (resource && resource !== 'resource') {
        showToast(`${action} successfully`, 'success');
      }
    }
    return response;
  },
  (error) => {
    const message =
      error?.response?.data?.detail ||
      error?.response?.data?.error ||
      error?.message ||
      'An error occurred';
    showToast(message, 'error');
    return Promise.reject(error);
  }
);

export default client;
export { client };

export const providersApi = {
  list: () => client.get('/providers/'),
  get: (id: string) => client.get(`/providers/${id}`),
  create: (data: Partial<Provider>) => client.post('/providers/', data),
  update: (id: string, data: Partial<Provider>) => client.put(`/providers/${id}`, data),
  remove: (id: string) => client.delete(`/providers/${id}`),
};

export const modelsApi = {
  list: (providerId?: string) => client.get('/models/', { params: { provider_id: providerId } }),
  fetchAvailable: (providerId: string) => client.get(`/models/available/${providerId}`),
  create: (data: Partial<Model>) => client.post('/models/', data),
  update: (id: string, data: Partial<Model>) => client.put(`/models/${id}`, data),
  remove: (id: string) => client.delete(`/models/${id}`),
};

export const tokensApi = {
  list: () => client.get('/tokens/'),
  create: (data: Partial<Token> & { model_permissions?: Array<{ model_id: string; max_input_tokens: number | null; max_output_tokens: number | null }> }) => client.post('/tokens/', data),
  update: (id: string, data: Partial<Token> & { model_permissions?: Array<{ model_id: string; max_input_tokens: number | null; max_output_tokens: number | null }> }) => client.put(`/tokens/${id}`, data),
  remove: (id: string) => client.delete(`/tokens/${id}`),
  usage: (id: string) => client.get(`/tokens/${id}/usage`),
};

export const usageApi = {
  logs: (params?: Record<string, unknown>) => client.get('/usage/logs', { params }),
  aggregate: (days?: number) => client.get('/usage/aggregate', { params: { days } }),
};

export const configApi = {
  list: () => client.get('/config/'),
  get: (key: string) => client.get(`/config/${key}`),
  update: (key: string, value: string) => client.put(`/config/${key}`, { value }),
};

export const dashboardApi = {
  stats: () => client.get('/dashboard/stats'),
};

export const proxyApi = {
  logs: () => client.get('/proxy/logs'),
  clearLogs: () => client.delete('/proxy/logs'),
  status: () => client.get('/proxy/status'),
  metrics: () => client.get('/proxy/metrics'),
};

import type { Provider, Model, Token } from '../types';
