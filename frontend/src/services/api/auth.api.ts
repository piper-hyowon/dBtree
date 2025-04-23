import {apiClient, handleApiError} from './client';

export interface User {
    id: string;
    email: string;
    lemonBalance: number;
    totalHarvested: number;
    joinedAt: number;
    instances: string[];
    createdAt: Date;
    updatedAt: Date;
}

export interface SendOTPResponse {
    success: boolean;
    message?: string;
    isNewUser: boolean;
}

export interface VerifyOTPResponse {
    success: boolean;
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
 * @returns isNewUser
 */
export const sendOTP = async (email: string): Promise<boolean> => {
    try {
        const response = await apiClient.post<SendOTPResponse>('/send-otp?type=authentication', {email});
        console.log(response);
        return response.data.isNewUser;
    } catch (error) {
        console.log("error: ", error)
        handleApiError(error);
        throw error;
    }
};

export const verifyOTP = async (email: string, otpCode: string): Promise<VerifyOTPResponse> => {
    try {
        const response = await apiClient.post<VerifyOTPResponse>('/verify-otp?type=authentication', {
            email,
            otpCode
        });

        if (response.data.success && response.data.token) {
            localStorage.setItem('token', response.data.token);

            if (response.data.user) {
                localStorage.setItem('user', JSON.stringify(response.data.user));
            }
        }

        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};


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
    }
};


export const getCurrentUser = (): User | null => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
        try {
            return JSON.parse(userStr) as User;
        } catch (e) {
            localStorage.removeItem('user');
            return null;
        }
    }

    return null;
};

export const isAuthenticated = (): boolean => {
    return localStorage.getItem('token') !== null;
};

export default {
    sendOTP,
    verifyOTP,
    logout,
    getCurrentUser,
    isAuthenticated
};