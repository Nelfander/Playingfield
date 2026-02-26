// src/config/env.ts
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:880';
export const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:880';

// Helpers (makes code cleaner)
export const apiUrl = (path: string) => `${API_BASE_URL}${path.startsWith('/') ? '' : '/'}${path}`;
export const wsUrl = (path: string) => `${WS_BASE_URL}${path.startsWith('/') ? '' : '/'}${path}`;