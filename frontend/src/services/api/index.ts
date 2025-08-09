import authApi from './auth.api';
import userApi from './user.api';
import homeApi from "./home.api";
import accountApi from "./account.api";

const api = {
    auth: authApi,
    user: userApi,
    home: homeApi,
    account: accountApi,
};

export default api;

