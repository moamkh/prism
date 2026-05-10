export interface Provider {
  id: string;
  name: string;
  base_url: string;
  api_token: string;
  http_proxy: string | null;
  socks5_proxy: string | null;
  enable_proxy: boolean;
  max_concurrent_requests: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Model {
  id: string;
  provider_id: string;
  model_id: string;
  display_model_id: string | null;
  is_active: boolean;
  created_at: string;
}

export interface Token {
  id: string;
  name: string;
  key_hash: string;
  max_input_tokens: number | null;
  max_output_tokens: number | null;
  requests_per_minute: number | null;
  is_active: boolean;
  created_at: string;
  model_permissions: Array<{
    model_id: string;
    max_input_tokens: number | null;
    max_output_tokens: number | null;
  }>;
}

export interface TokenModelUsage {
  model_id: string;
  model_name: string;
  max_tokens: number;
  current_usage: number;
  percentage: number;
}

export interface UsageLog {
  id: string;
  token_id: string | null;
  provider_id: string | null;
  model_id: string | null;
  request_path: string | null;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  latency_ms: number | null;
  status_code: number | null;
  created_at: string;
}

export interface ConfigItem {
  key: string;
  value: string;
  updated_at: string;
}

export interface TopModel {
  model_id: string | null;
  model_name: string | null;
  provider_name: string | null;
  total_requests: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_tokens: number;
}

export interface DashboardStats {
  total_requests: number;
  total_input_tokens: number;
  total_output_tokens: number;
  active_providers: number;
  active_tokens: number;
  top_models: TopModel[];
  recent_logs: UsageLog[];
}
