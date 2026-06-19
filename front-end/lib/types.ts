export interface Hub {
  id: string;
  name: string;
  code: string;
  city: string;
  province: string;
  lat: number;
  lng: number;
  type: string;
  is_active: boolean;
}

export interface NotificationItem {
  id: string;
  user_id: string;
  channel: string;
  title: string;
  message: string;
  order_id?: string;
  awb?: string;
  status: string;
  is_read: boolean;
  created_at: string;
}

export interface ServiceInfo {
  code: string;
  name: string;
  description: string;
  price_per_km: number;
  price_per_kg: number;
  base_fee: number;
  estimated_days: string;
  is_active: boolean;
}

export interface PriceCalculationRequest {
  origin_lat: number;
  origin_lng: number;
  dest_lat: number;
  dest_lng: number;
  weight_kg: number;
  length_cm: number;
  width_cm: number;
  height_cm: number;
  service_type: string;
  is_insured: boolean;
  insured_value: number;
}

export interface PriceCalculationResponse {
  service_type: string;
  service_name: string;
  distance_km: number;
  weight_kg: number;
  volumetric_kg: number;
  chargeable_kg: number;
  base_cost: number;
  weight_cost: number;
  insurance_cost: number;
  discount: number;
  total_cost: number;
  estimated_days: string;
}

export interface TrackingEvent {
  id: string;
  awb: string;
  order_id: string;
  status: string;
  location: string;
  detail: string;
  lat?: number;
  lng?: number;
  source: string;
  timestamp: string;
}

export interface CourierGPS {
  courier_id: string;
  awb: string;
  lat: number;
  lng: number;
  timestamp: number;
}

export interface TrackingTimeline {
  awb: string;
  order_id: string;
  current_status: string;
  events: TrackingEvent[];
  live_gps?: CourierGPS;
}

export interface Order {
  id: string;
  awb: string;
  user_id: string;
  status: string;
  service_type: string;
  sender_name: string;
  sender_phone: string;
  sender_address: string;
  receiver_name: string;
  receiver_phone: string;
  receiver_address: string;
  sender_city?: string;
  receiver_city?: string;
  item_description?: string;
  weight_kg: number;
  length_cm: number;
  width_cm: number;
  height_cm: number;
  shipping_cost: number;
  total_cost: number;
  is_insured?: boolean;
  insured_value?: number;
  // Routing: nearest sortation hubs to sender (origin) and receiver (destination).
  origin_hub_code?: string;
  origin_hub_name?: string;
  dest_hub_code?: string;
  dest_hub_name?: string;
  created_at: string;
  updated_at: string;
}

export interface Assignment {
  id: string;
  order_id: string;
  awb: string;
  courier_id: string;
  courier_name: string;
  status: string; // OPEN, ASSIGNED, PICKED_UP, COMPLETED, NO_SHOW, REASSIGNED
  leg?: string; // FIRST_MILE, LAST_MILE, DIRECT
  pickup_address: string;
  pickup_lat?: number;
  pickup_lng?: number;
  dropoff_address?: string;
  dropoff_lat?: number;
  dropoff_lng?: number;
  hub_name?: string;
  assigned_at: string;
  picked_up_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface DashboardStats {
  total_orders_today: number;
  active_couriers: number;
  active_hubs: number;
  sla_percentage: number;
  total_cities: number;
}

export interface VolumeDataPoint {
  date: string;
  count: number;
}

export interface AuthUser {
  id: string;
  email: string;
  full_name?: string;
  phone?: string;
  first_name?: string;
  last_name?: string;
  role: string;
}

export interface AuthResponse {
  token: string;
  user: AuthUser;
}

export interface ApiError {
  success: boolean;
  error: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data: T;
  error?: string;
  meta?: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}
