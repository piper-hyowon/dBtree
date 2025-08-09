import {apiClient, handleApiError} from './client';
import {User} from "./auth.api";

export const getUserProfile = async (): Promise<User> => {
    try {
        const response = await apiClient.get<User>('/user');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export default {
    getUserProfile,
};