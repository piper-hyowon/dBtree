import {apiClient, handleApiError} from './client';

export interface User {
    id: string;
    email: string;
    lemonBalance: number;
    totalEarnedLemons: number;
    totalSpentLemons: number;
    lastHarvestAt: string | null;
    joinedAt: string;
}

export interface SendOTPResponse {
    isNewUser: boolean;
}

export interface VerifyOTPResponse {
    profile: {
        id: string;
        email: string;
        lemonBalance: number;
        totalEarnedLemons: number;
        totalSpentLemons: number;
        lastHarvestAt: string | null;
        joinedAt: string;
    }
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

        if (response.data.profile) {
            localStorage.setItem('user', JSON.stringify(response.data.profile));
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
        const response = await apiClient.post<LogoutResponse>('/logout');
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

/**
 * 탈퇴
 */
export const deleteAccount = async (): Promise<void> => {
    try {
        const response = await apiClient.delete('/user');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export default {
    sendOTP,
    verifyOTP,
    logout,
    isAuthenticated,
    deleteAccount,
};