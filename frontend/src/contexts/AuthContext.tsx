import React, {createContext, useContext, useState, useEffect, ReactNode} from 'react';
import {
    logout as logoutApi,
    sendOTP,
    User,
    verifyOTP,
    VerifyOTPResponse
} from '../services/api/auth.api';

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
    refreshUser: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

let tempEmail: string | null = null;

export const AuthProvider: React.FC<{ children: ReactNode }> = ({children}) => {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    const loadUser = () => {
        const token = localStorage.getItem('token');
        const userStr = localStorage.getItem('user');

        if (token && userStr) {
            try {
                const userData = JSON.parse(userStr);
                setUser(userData);
            } catch (err) {
                setUser(null);
                localStorage.removeItem('token');
                localStorage.removeItem('user');
            }
        } else {
            setUser(null);
        }
    };

    useEffect(() => {
        const token = localStorage.getItem('token');
        const userStr = localStorage.getItem('user');

        if (token && userStr) {
            try {
                const userData = JSON.parse(userStr);
                setUser(userData);
            } catch (err) {
                localStorage.removeItem('token');
                localStorage.removeItem('user');
            }
        }

        setLoading(false);
    }, []);

    useEffect(() => {
        const handleStorageChange = (e: StorageEvent) => {
            if (e.key === 'user' || e.key === 'token') {
                loadUser();
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

            if (response.user) {
                setUser(response.user);
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

    const refreshUser = () => {
        loadUser();
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