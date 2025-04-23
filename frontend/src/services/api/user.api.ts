import {apiClient, handleApiError} from './client';
import {User} from "./auth.api";

export interface UserProfileResponse {
    success: boolean;
    user: User;
}


export const getUserProfile = async (): Promise<UserProfileResponse> => {
    try {
        const response = await apiClient.get<UserProfileResponse>('/profile');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export default {
    getUserProfile,
};