import {apiClient, handleApiError} from './client';

export interface DailyHarvest {
    date: string;
    amount: number;
}

export type DailyHarvestResponse = DailyHarvest[];

export interface DailyHarvestParams {
    days?: number;
}

export interface TransactionWithInstance {
    id: string;
    instanceId: string;
    instanceName: string | null;
    actionType: 'welcome_bonus' | 'harvest' | 'instance_create' | 'instance_maintain' | 'instance_create_refund';
    status: 'successful' | 'failed';
    amount: number;
    balance: number;
    createdAt: string;
    note: string;
}

export interface PaginationInfo {
    currentPage: number;
    totalPages: number;
    totalItems: number;
    hasNext: boolean;
    hasPrev: boolean;
}

export interface TransactionsResponse {
    data: TransactionWithInstance[];
    pagination: PaginationInfo;
}

export interface TransactionsParams {
    page?: number;
    limit?: number;
    instance_name?: string;
}

export interface UserInstanceSummary {
    id: string;
    name: string;
}

export type UserInstancesResponse = UserInstanceSummary[];

export const getDailyHarvest = async (params?: DailyHarvestParams): Promise<DailyHarvestResponse> => {
    try {
        const response = await apiClient.get<DailyHarvestResponse>('/stats/daily-harvest', {
            params
        });
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export const getTransactions = async (params?: TransactionsParams): Promise<TransactionsResponse> => {
    try {
        const response = await apiClient.get<TransactionsResponse>('/stats/transactions', {
            params
        });
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export const getInstanceNames = async (): Promise<UserInstancesResponse> => {
    try {
        const response = await apiClient.get<UserInstancesResponse>('/stats/summary/instances');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export default {
    getDailyHarvest,
    getTransactions,
    getInstanceNames
};