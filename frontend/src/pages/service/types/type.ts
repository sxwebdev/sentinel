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
      grpc: object | null;
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
    response_time_ns: number;
    created_at: string;
    updated_at: string;
  };
}
