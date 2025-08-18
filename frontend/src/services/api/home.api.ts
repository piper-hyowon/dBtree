import {apiClient, handleApiError} from './client';

export interface GlobalStatsResponse {
    totalHarvestedLemons: number;
    totalCreatedInstances: number;
    totalUsers: number;
}

export interface MiniLeaderboard {
    lemonRichUsers: UserRank[] | null;
    quizMasters: UserRank[] | null;
}

export interface UserRank {
    maskedEmail: string;
    score: number;
    rank: number;
}

export interface TreeStatusResponse {
    availablePositions: number[];
    totalHarvested: number;
    nextRegrowthTime?: string;
}

export interface SystemResourceSpec {
    cpu: number;
    memory: number;
}

export interface SystemResources {
    total: SystemResourceSpec;
    reserved: SystemResourceSpec;
    available: SystemResourceSpec;
    used: SystemResourceSpec;
}

export interface InstanceResources {
    instanceId: string;
    instanceName: string;
    resources: SystemResourceSpec;
    status: string;
}

export interface SystemResourceStatusResponse {
    system: SystemResources;
    instances: InstanceResources[] | null;
    activeCount: number;
    canCreateTiny: boolean;
    canCreateSmall: boolean;
    canCreateMedium: boolean;
}

export const getGlobalStats = async (): Promise<GlobalStatsResponse> => {
    try {
        const response = await apiClient.get<GlobalStatsResponse>('/stats/global');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export const getLeaderboardMini = async (): Promise<MiniLeaderboard> => {
    try {
        const response = await apiClient.get<MiniLeaderboard>('/leaderboard/mini');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
}

export const getLemonTreeStatus = async (): Promise<TreeStatusResponse> => {
    try {
        const response = await apiClient.get<TreeStatusResponse>('/lemon/global-status');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
}

export const getSystemResources = async (): Promise<SystemResourceStatusResponse> => {
    try {
        const response = await apiClient.get<SystemResourceStatusResponse>('/system/resources');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export default {
    getGlobalStats,
    getLeaderboardMini,
    getLemonTreeStatus,
    getSystemResources,
};