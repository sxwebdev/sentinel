import axios from "axios";

// Get configuration from global variables injected by the server
declare global {
  interface Window {
    __SENTINEL_CONFIG__: {
      baseUrl: string;
      socketUrl: string;
    };
  }
}

// Use configuration from server or fallback to defaults
const config = window.__SENTINEL_CONFIG__ || {
  baseUrl: "http://localhost:8080/api/v1",
  socketUrl: "ws://localhost:8080/ws",
};

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

