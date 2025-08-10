import {apiClient, handleApiError} from './client';
import {
    InstanceResponse,
    CreateInstanceRequest,
    PresetResponse,
    EstimateCostRequest,
    EstimateCostResponse
} from '../../types/database.types';

/**
 * 인스턴스 목록 조회
 */
export const getInstances = async (): Promise<InstanceResponse[]> => {
    try {
        const response = await apiClient.get<InstanceResponse[]>('/db/instances');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 인스턴스 상세 조회 (동기화 포함)
 */
export const getInstance = async (instanceId: string): Promise<InstanceResponse> => {
    try {
        const response = await apiClient.get<InstanceResponse>(`/db/instances/${instanceId}`);
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 인스턴스 생성
 */
export const createInstance = async (data: CreateInstanceRequest): Promise<InstanceResponse> => {
    try {
        const response = await apiClient.post<InstanceResponse>('/db/instances', data);
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 인스턴스 삭제
 */
export const deleteInstance = async (instanceId: string): Promise<void> => {
    try {
        await apiClient.delete(`/db/instances/${instanceId}`);
        // 204 No Content - 응답 바디 없음
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 인스턴스 중지
 */
export const stopInstance = async (instanceId: string): Promise<void> => {
    try {
        await apiClient.post(`/db/instances/${instanceId}/stop`);
        // 204 No Content - 응답 바디 없음
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 인스턴스 재시작
 */
export const restartInstance = async (instanceId: string): Promise<void> => {
    try {
        await apiClient.post(`/db/instances/${instanceId}/restart`);
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 인스턴스 시작
 */
export const startInstance = async (instanceId: string): Promise<void> => {
    try {
        await apiClient.post(`/db/instances/${instanceId}/start`);
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 프리셋 목록 조회
 */
export const getPresets = async (): Promise<PresetResponse[]> => {
    try {
        const response = await apiClient.get<PresetResponse[]>('/db/presets');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

/**
 * 비용 예상
 */
export const estimateCost = async (data: EstimateCostRequest): Promise<EstimateCostResponse> => {
    try {
        const response = await apiClient.post<EstimateCostResponse>('/db/estimate-cost', data);
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};