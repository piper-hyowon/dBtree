import { useState, useEffect } from "react";
import { User } from "../types/api.types";
import { apiService } from "../services/mockApi";

interface UseAuthReturn {
  isLoggedIn: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
  requestOtp: (email: string) => Promise<boolean>;
  verifyOtp: (otp: string) => Promise<boolean>;
  resendOtp: () => Promise<boolean>;
  logout: () => Promise<void>;
  clearError: () => void;
}

let tempEmail: string | null = null;

export const useAuth = (): UseAuthReturn => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await apiService.getCurrentUser();
        if (response.success && response.data) {
          setUser(response.data);
        }
      } catch (err) {
        // 인증되지 않은 상태는 에러로 처리하지 않음
        console.log("사용자 인증이 필요합니다.");
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  const requestOtp = async (email: string): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      const response = await apiService.sendOtp(email);
      if (response.success) {
        tempEmail = email;
        return true;
      } else {
        setError(response.error || "인증 코드 발송에 실패했습니다");
        return false;
      }
    } catch (err) {
      setError("서버 오류가 발생했습니다");
      return false;
    } finally {
      setLoading(false);
    }
  };

  const resendOtp = async (): Promise<boolean> => {
    if (!tempEmail) return false;

    setLoading(true);
    setError(null);

    try {
      const response = await apiService.sendOtp(tempEmail);
      return response.success;
    } catch (err) {
      setError("인증 코드 재전송에 실패했습니다");
      return false;
    } finally {
      setLoading(false);
    }
  };

  const verifyOtp = async (otp: string): Promise<boolean> => {
    if (!tempEmail) return false;

    setLoading(true);
    setError(null);

    try {
      const response = await apiService.verifyOtp(tempEmail, otp);
      if (response.success && response.data) {
        setUser(response.data.user);
        return true;
      } else {
        setError(response.error || "인증에 실패했습니다");
        return false;
      }
    } catch (err) {
      setError("인증 과정에서 오류가 발생했습니다");
      return false;
    } finally {
      setLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    setLoading(true);

    try {
      await apiService.logout();
      setUser(null);
    } catch (err) {
      setError("로그아웃 중 오류가 발생했습니다");
    } finally {
      setLoading(false);
    }
  };

  const clearError = () => {
    setError(null);
  };

  return {
    isLoggedIn: !!user,
    user,
    loading,
    error,
    requestOtp,
    verifyOtp,
    resendOtp,
    logout,
    clearError,
  };
};
