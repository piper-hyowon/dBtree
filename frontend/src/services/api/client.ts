import axios, {AxiosError, InternalAxiosRequestConfig} from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export const apiClient = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    }
});

apiClient.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error: AxiosError) => Promise.reject(error)
);

apiClient.interceptors.response.use(
    response => response,
    (error: AxiosError) => {
        if (error.response) {
            const headers = error.response.headers;

            if (error.response.status === 429) {
                const retryAfter = headers['retry-after'];

                if (retryAfter) {
                    const seconds = parseInt(retryAfter, 10);
                    if (!isNaN(seconds)) {
                        (error as any).retryAfter = seconds;
                    }
                }
            }
        }
        return Promise.reject(error);
    }
);

export const handleApiError = (error: unknown): void => {
    if (axios.isAxiosError(error)) {
        if (error.response) {
            const isAuthFlow = error.config?.url?.includes('/verify-otp') ||
                error.config?.url?.includes('/send-otp');

            if (error.response.status === 401 && !isAuthFlow) {
                console.log("401 Error - Session expired, logging out");
                localStorage.removeItem('token');
                localStorage.removeItem('user');
            }

            if (error.response.status === 429 && !('retryAfter' in error)) {
                const retryAfter = error.response.headers['retry-after'];
                if (retryAfter) {
                    const seconds = parseInt(retryAfter, 10);
                    if (!isNaN(seconds)) {
                        (error as any).retryAfter = seconds;
                    }
                }
            }
        }
    }
};