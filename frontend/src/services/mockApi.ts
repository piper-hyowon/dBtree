import {
  ApiResponse,
  User,
  SendOtpRequest,
  SendOtpResponse,
  VerifyOtpRequest,
  VerifyOtpResponse,
  LogoutResponse,
} from "../types/api.types";

// Mock 사용자 데이터
const MOCK_USERS: Record<string, User> = {
  "user@example.com": {
    id: "1",
    email: "user@example.com",
    name: "테스트 사용자",
    credits: 100,
  },
  "admin@example.com": {
    id: "2",
    email: "admin@example.com",
    name: "관리자",
    credits: 500,
  },
};

let currentUser: User | null = null;

export const mockApi = {
  /**
   * OTP 요청 API
   * POST /send-otp
   */
  sendOtp: async (
    req: SendOtpRequest
  ): Promise<ApiResponse<SendOtpResponse>> => {
    // 요청 검증
    if (!req.email) {
      return {
        success: false,
        error: "이메일이 필요합니다",
      };
    }

    console.log(
      `[MOCK] OTP 코드 ${Math.floor(100000 + Math.random() * 900000)}이 ${
        req.email
      }로 발송되었습니다.`
    );

    await new Promise((resolve) => setTimeout(resolve, 500));

    return {
      success: true,
      data: {
        message: `OTP가 ${req.email}로 발송되었습니다.`,
      },
    };
  },

  /**
   * OTP 검증 API
   * POST /verify-otp
   */
  verifyOtp: async (
    req: VerifyOtpRequest
  ): Promise<ApiResponse<VerifyOtpResponse>> => {
    if (!req.email || !req.otp) {
      return {
        success: false,
        error: "이메일과 OTP 코드가 필요합니다",
      };
    }

    await new Promise((resolve) => setTimeout(resolve, 700));

    if (Math.random() < 0.1) {
      return {
        success: false,
        error: "잘못된 OTP 코드입니다",
      };
    }

    const user = MOCK_USERS[req.email] || {
      id: Math.random().toString(36).substring(2, 9),
      email: req.email,
      credits: 50,
    };

    currentUser = user;

    return {
      success: true,
      data: {
        user,
      },
    };
  },

  /**
   * 현재 사용자 정보 조회 API
   * GET /api/auth/me
   */
  getCurrentUser: async (): Promise<ApiResponse<User>> => {
    await new Promise((resolve) => setTimeout(resolve, 300));
    if (!currentUser) {
      return {
        success: false,
        error: "인증되지 않은 사용자입니다",
      };
    }

    return {
      success: true,
      data: currentUser,
    };
  },

  /**
   * 로그아웃 API
   * POST /api/auth/logout
   */
  logout: async (): Promise<ApiResponse<LogoutResponse>> => {
    await new Promise((resolve) => setTimeout(resolve, 300));
    currentUser = null;

    return {
      success: true,
      data: {
        message: "로그아웃 되었습니다",
      },
    };
  },
};

export const apiService = {
  sendOtp: (email: string, type: "authentication" = "authentication") =>
    mockApi.sendOtp({ email, type }),

  verifyOtp: (
    email: string,
    otp: string,
    type: "authentication" = "authentication"
  ) => mockApi.verifyOtp({ email, otp, type }),

  getCurrentUser: () => mockApi.getCurrentUser(),

  logout: () => mockApi.logout(),
};
