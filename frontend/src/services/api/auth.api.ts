import {apiClient, handleApiError} from './client';

export interface User {
    id: string;
    email: string;
    lemonBalance: number;
    lastHarvest: string | null;
    createdAt: string;
    updatedAt: string;
}

export interface SendOTPResponse {
    success?: boolean;
    message?: string;
    isNewUser: boolean;
}

export interface VerifyOTPResponse {
    success?: boolean;
    message?: string;
    user?: User;
    token?: string;
    expiresIn?: number;
}

export interface LogoutResponse {
    success: boolean;
    message?: string;
}

/**
 * OTP 발송 요청
 * @returns isNewUser - 신규 사용자 여부
 */
export const sendOTP = async (email: string): Promise<boolean> => {
    try {
        const response = await apiClient.post<SendOTPResponse>('/send-otp?type=authentication', {email});
        return response.data.isNewUser;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * OTP 검증 및 로그인
 */
export const verifyOTP = async (email: string, otp: string): Promise<VerifyOTPResponse> => {
    try {
        const response = await apiClient.post<VerifyOTPResponse>('/verify-otp?type=authentication', {
            email,
            otp
        });

        if (response.data.token) {
            localStorage.setItem('token', response.data.token);
        }

        if (response.data.user) {
            localStorage.setItem('user', JSON.stringify(response.data.user));
        }

        if (response.data.expiresIn) {
            const expiresAt = Date.now() + (response.data.expiresIn * 1000);
            localStorage.setItem('tokenExpiresAt', expiresAt.toString());
        }

        return response.data;

    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 로그아웃
 */
export const logout = async (): Promise<LogoutResponse> => {
    try {
        const response = await apiClient.post<LogoutResponse>('/auth/logout');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    } finally {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('tokenExpiresAt');
    }
};

/**
 * 현재 로그인한 사용자 정보 가져오기
 */
export const getCurrentUser = (): User | null => {
    const token = localStorage.getItem('token');
    const userStr = localStorage.getItem('user');

    if (!token || !userStr) {
        return null;
    }

    try {
        return JSON.parse(userStr) as User;
    } catch (e) {
        localStorage.removeItem('user');
        localStorage.removeItem('token');
        return null;
    }
};

/**
 * 인증 여부 확인
 */
export const isAuthenticated = (): boolean => {
    const token = localStorage.getItem('token');

    if (!token) {
        return false;
    }

    const expiresAt = localStorage.getItem('tokenExpiresAt');
    if (expiresAt) {
        const isExpired = parseInt(expiresAt) < Date.now();
        if (isExpired) {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            localStorage.removeItem('tokenExpiresAt');
            return false;
        }
    }

    return true;
};

export default {
    sendOTP,
    verifyOTP,
    logout,
    getCurrentUser,
    isAuthenticated
};