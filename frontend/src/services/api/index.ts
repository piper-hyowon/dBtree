import authApi from './auth.api';
import userApi from './user.api';
import homeStatsApi from "./home-stats.api";

const api = {
    auth: authApi,
    user: userApi,
    homeStats: homeStatsApi,
};

export default api;

