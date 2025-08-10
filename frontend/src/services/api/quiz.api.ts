import {apiClient, handleApiError} from './client';

export interface HarvestResponse {
    harvestAmount: number;
    newBalance: number;
    nextHarvestTime: number;
    TransactionID: string;
}

export interface StartQuizResponse {
    question: string;
    options: string[];
    timeLimit: number; // 초
    attemptID: number;
}

export enum QuizStatus {
    Started = "started",  // 퀴즈 시작 / 정답 미제출
    Done = "done", // 제출 완료(정답 여부는 is_correct로 구분)
    Timeout = "timeout", // 제한 시간 초과
    HarvestNone = "none", // 아직 수확 단계 아님(Default)
    HarvestInProgress = "in_progress", // 원이 나타나서 클릭 대기 중
    HarvestSuccess = "success", // 레몬 수확 성공
    HarvestTimeout = "timeout", // 원 클릭 시간 초과
    HarvestFailure = "failure", // 수확 실패 (다른 사용자가 먼저 수확)
}

export interface SubmitAnswerResponse {
    isCorrect: boolean;
    status: QuizStatus;
    correctOption: number;
    harvestEnabled: boolean;
    harvestTimeoutAt: string;
    attemptID: number;
}

export interface HarvestAvailability {
    canHarvest: boolean;
    waitSeconds: number;
}

export const canHarvest = async (): Promise<HarvestAvailability> => {
    try {
        const response = await apiClient.get<HarvestAvailability>('/lemon/harvestable');
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
}

export const harvestLemon = async (positionId: number, attemptId: number): Promise<HarvestResponse> => {
    try {
        const response = await apiClient.post<HarvestResponse>('/lemon/harvest', {
            positionId, attemptId
        });
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export const getQuizQuestions = async (positionId: number): Promise<StartQuizResponse> => {
    try {
        const response = await apiClient.get<StartQuizResponse>(`/quiz/${positionId}`);
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
}

export const submitQuizAnswer = async (optionIdx: number, attemptID: number): Promise<SubmitAnswerResponse> => {
    try {
        const response = await apiClient.post<SubmitAnswerResponse>('/quiz/answer', {
            optionIdx, attemptID
        });
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
}

export default {
    harvestLemon,
    getQuizQuestions,
    submitQuizAnswer,
    canHarvest,
};