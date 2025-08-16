import axios from "axios";

// Get configuration from environment variables or server injection
declare global {
  interface Window {
    __SENTINEL_CONFIG__?: {
      baseUrl: string;
      socketUrl: string;
    };
  }
}

// Priority: env variables > server config > defaults
const getConfig = () => {
  // First try environment variables (Vite)
  const envBaseUrl = import.meta.env.VITE_API_BASE_URL;
  const envSocketUrl = import.meta.env.VITE_SOCKET_URL;

  if (envBaseUrl && envSocketUrl) {
    return {
      baseUrl: envBaseUrl,
      socketUrl: envSocketUrl,
    };
  }

  // Fallback to server config
  if (window.__SENTINEL_CONFIG__) {
    return window.__SENTINEL_CONFIG__;
  }

  // Final fallback to defaults (use current host)
  const currentHost = window.location.hostname;
  const protocol = window.location.protocol === "https:" ? "https:" : "http:";
  const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";

  return {
    baseUrl: `${protocol}//${currentHost}:8080/api/v1`,
    socketUrl: `${wsProtocol}//${currentHost}:8080/ws`,
  };
};

const config = getConfig();

const $api = axios.create({
  baseURL: config.baseUrl,
});

// $api is used in orval-generated functions for making HTTP requests
export const customFetcher = async <T>({
  url,
  method,
  data,
  params,
  headers,
}: {
  url: string;
  method: string;
  data?: T;
  params?: Record<string, unknown>;
  headers?: Record<string, string>;
}): Promise<T> => {
  const response = await $api.request<T>({
    url,
    method,
    data,
    params,
    headers,
  });

  return response.data;
};

export const socketUrl = config.socketUrl;
