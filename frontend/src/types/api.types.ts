export interface ApiResponse<T> {
    success: boolean;
    data?: T;
    error?: string;
}

export interface GlobalStatsResponse {
    totalHarvested: number;
    totalDbInstances: number;
    activeUsers: number;
}
