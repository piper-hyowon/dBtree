import {apiClient, handleApiError} from './client';

export interface SupportContext {
    userEmail?: string;
    lemonBalance?: number;
    timestamp: string;
    userAgent: string;
    currentPage: string;
    instanceCount?: number;
}

export interface SendInquiryRequest {
    category: 'bug' | 'feature' | 'billing' | 'other';
    subject: string;
    message: string;
    context: SupportContext;
}

export interface SendInquiryResponse {
    success: boolean;
    ticketId?: string;
    message: string;
}

export const sendInquiry = async (data: SendInquiryRequest): Promise<SendInquiryResponse> => {
    try {
        const response = await apiClient.post<SendInquiryResponse>('/support/inquiry', data);
        return response.data;
    } catch (error) {
        handleApiError(error);
        throw error;
    }
};

export default {
    sendInquiry
};