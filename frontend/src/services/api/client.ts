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

export const handleApiError = (error: unknown): void => {
    if (axios.isAxiosError(error) && error.response) {
        if (error.response.status === 401) {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/';
        }

        if (error.response.status === 429 && error.response.headers['retry-after']) {
            const retryAfter = parseInt(error.response.headers['retry-after'], 10);
            console.log(`Retry after ${retryAfter} seconds`);
        }
    }
};