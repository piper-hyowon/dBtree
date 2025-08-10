import React, {createContext, useContext, useState, useEffect, ReactNode} from 'react';
import {
    logout as logoutApi,
    sendOTP,
    User,
    verifyOTP,
    VerifyOTPResponse
} from '../services/api/auth.api';
import api from '../services/api';

interface AuthContextType {
    isLoggedIn: boolean;
    user: User | null;
    loading: boolean;
    error: string | null;
    requestOtp: (email: string) => Promise<{
        success: boolean,
        isNewUser?: boolean,
        message?: string,
        retryAfter?: number
    }>;
    verifyOtp: (otp: string) => Promise<{
        success: boolean,
        isNewUser?: boolean,
        message?: string,
        retryAfter?: number
    }>;
    resendOtp: () => Promise<{ success: boolean, isNewUser?: boolean, message?: string, retryAfter?: number }>;
    logout: () => Promise<void>;
    clearError: () => void;
    refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

let tempEmail: string | null = null;

export const AuthProvider: React.FC<{ children: ReactNode }> = ({children}) => {
    const [user, setUser] = useState<User | null>(() => {
        const token = localStorage.getItem('token');
        const userStr = localStorage.getItem('user');

        if (token && userStr) {
            try {
                return JSON.parse(userStr);
            } catch {
                localStorage.removeItem('token');
                localStorage.removeItem('user');
                return null;
            }
        }
        return null;
    });
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    const refreshUser = async (): Promise<void> => {
        const token = localStorage.getItem('token');

        if (!token) {
            setUser(null);
            return;
        }

        try {
            // API에서 최신 사용자 정보 가져오기
            const userResponse = await api.user.getUserProfile();

            // 상태 업데이트
            setUser(userResponse);

            // localStorage도 업데이트
            localStorage.setItem('user', JSON.stringify(userResponse));
        } catch (error: any) {
            console.error('Failed to refresh user data:', error);
            // 토큰이 유효하지 않으면 로그아웃 처리
            if (error?.response?.status === 401) {
                localStorage.removeItem('token');
                localStorage.removeItem('user');
                setUser(null);
            }
        }
    };

    // 초기 로드 시 API 호출하여 최신 정보 가져오기
    useEffect(() => {
        const initAuth = async () => {
            const token = localStorage.getItem('token');

            if (token) {
                await refreshUser();
            }
            setLoading(false);
        };

        initAuth();
    }, []);

    useEffect(() => {
        const handleStorageChange = async (e: StorageEvent) => {
            if (e.key === 'user' || e.key === 'token') {
                const token = localStorage.getItem('token');
                if (token) {
                    await refreshUser();
                } else {
                    setUser(null);
                }
            }
        };

        window.addEventListener('storage', handleStorageChange);
        return () => window.removeEventListener('storage', handleStorageChange);
    }, []);

    const requestOtp = async (email: string): Promise<{
        success: boolean,
        isNewUser?: boolean,
        message?: string,
        retryAfter?: number
    }> => {
        setLoading(true);
        setError(null);

        try {
            const isNewUser = await sendOTP(email);
            tempEmail = email;
            return {success: true, isNewUser};
        } catch (err: any) {
            const errorMessage = err?.response?.data?.error || "인증 코드 발송에 실패했습니다";
            const retryAfter = err?.response?.headers?.['retry-after'];
            const waitSeconds = retryAfter ? parseInt(retryAfter, 10) : undefined;

            setError(errorMessage);
            return {
                success: false,
                message: errorMessage,
                retryAfter: waitSeconds && !isNaN(waitSeconds) ? waitSeconds : undefined
            };
        } finally {
            setLoading(false);
        }
    };

    const resendOtp = async (): Promise<{ success: boolean, message?: string, retryAfter?: number }> => {
        if (!tempEmail) {
            return {success: false, message: "이메일 정보가 없습니다."};
        }

        setLoading(true);
        setError(null);

        try {
            await sendOTP(tempEmail);
            return {success: true};
        } catch (err: any) {
            const errorMessage = err?.response?.data?.error || "인증 코드 재전송에 실패했습니다";
            const retryAfter = err?.response?.headers?.['retry-after'];
            const waitSeconds = retryAfter ? parseInt(retryAfter, 10) : undefined;

            setError(errorMessage);
            return {
                success: false,
                message: errorMessage,
                retryAfter: waitSeconds && !isNaN(waitSeconds) ? waitSeconds : undefined
            };
        } finally {
            setLoading(false);
        }
    };

    const verifyOtp = async (otp: string): Promise<{ success: boolean, message?: string }> => {
        if (!tempEmail) {
            return {success: false, message: "이메일 정보가 없습니다."};
        }

        setLoading(true);
        setError(null);

        try {
            const response: VerifyOTPResponse = await verifyOTP(tempEmail, otp);

            if (response.profile) {
                setUser(response.profile);
                localStorage.setItem('user', JSON.stringify(response.profile));
                tempEmail = null;
            }

            return {success: true};

        } catch (err: any) {
            const errorMessage = err?.response?.data?.error || "인증 과정에서 오류가 발생했습니다";
            setError(errorMessage);
            return {success: false, message: errorMessage};
        } finally {
            setLoading(false);
        }
    };

    const logout = async (): Promise<void> => {
        setLoading(true);

        try {
            await logoutApi();
            setUser(null);
        } catch (err) {
            setError("로그아웃 중 오류가 발생했습니다");
            setUser(null);
        } finally {
            setLoading(false);
        }
    };

    const clearError = () => {
        setError(null);
    };

    const value = {
        isLoggedIn: !!user,
        user,
        loading,
        error,
        requestOtp,
        verifyOtp,
        resendOtp,
        logout,
        clearError,
        refreshUser
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = (): AuthContextType => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};
