import authApi from './auth.api';
import userApi from './user.api';
import homeApi from "./home.api";
import accountApi from "./account.api";
import quizApi from './quiz.api';
import supportApi from "./support.api";

const api = {
    auth: authApi,
    user: userApi,
    home: homeApi,
    account: accountApi,
    quiz: quizApi,
    support: supportApi,
};

export default api;

