export interface User {
  id: string;
  email: string;
  name?: string;
  credits: number;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface SendOtpRequest {
  email: string;
  type: "authentication";
}

export interface SendOtpResponse {
  message: string;
}

export interface VerifyOtpRequest {
  email: string;
  otp: string;
  type: "authentication";
}

export interface VerifyOtpResponse {
  user: User;
}

export interface LogoutResponse {
  message: string;
}

export interface GlobalStatsResponse {
  totalHarvested: number;
  totalDbInstances: number;
  activeUsers: number;
}
