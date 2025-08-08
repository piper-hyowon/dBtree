import authApi from './auth.api';
import userApi from './user.api';
import homeStatsApi from "./home-stats.api";
import accountApi from "./account.api";

const api = {
    auth: authApi,
    user: userApi,
    homeStats: homeStatsApi,
    account: accountApi,
};

export default api;

