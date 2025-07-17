export interface Incident {
  id: string;
  service_id: string;
  start_time: string;
  end_time: string;
  error: string;
  duration: number;
  resolved: boolean;
}

export interface ServiceStats {
  avg_response_time: number;
  period: number;
  service_id: string;
  total_downtime: number;
  total_incidents: number;
  uptime_percentage: number;
}

export interface Service {
  service: {
    id: string;
    name: string;
    protocol: string;
    interval: number;
    timeout: number;
    retries: number;
    tags: string[];
    config: {
      grpc: GRPC | null;
      http: {
        condition: string;
        endpoints: {
          expected_status: number;
          method: string;
          name: string;
          url: string;
        }[];
      } | null;
      tcp: object | null;
    };
    is_enabled: boolean;
    total_incidents: number;
    active_incidents: number;
  };
  state: {
    id: string;
    service_id: string;
    status: string;
    last_check: string;
    next_check: string;
    consecutive_fails: number;
    consecutive_success: number;
    total_checks: number;
    last_error?: string;
    response_time_ns: number;
    created_at: string;
    updated_at: string;
  };
}

export interface HTTPEndpoint {
  body?: string;
  expected_status: number; // max 599 min 100
  headers?: string | object;
  json_path?: string;
  method: "GET" | "POST" | "PUT" | "DELETE" | "HEAD" | "OPTIONS";
  name: string;
  url: string;
  username?: string;
  password?: string;
}

export interface HTTP {
  condition?: string;
  endpoints?: HTTPEndpoint[];
  timeout?: number;
}

export interface TCP {
  endpoint: string;
  expect_data?: string;
  send_data?: string;
}

export interface GRPC {
  check_type: "health" | "reflection" | "connectivity";
  endpoint: string;
  tls?: boolean;
  service_name?: string;
  insecure_tls?: boolean;
}


export interface ServiceForm {
  name: string;
  protocol: string;
  interval?: number;
  timeout?: number;
  retries?: number;
  tags?: string[];
  is_enabled?: boolean;
  config: {
    http?: HTTP | null;
    tcp?: TCP | null;
    grpc?: GRPC | null;
  };
}