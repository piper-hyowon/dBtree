import {useState, useEffect} from "react";
import {
    getCurrentUser,
    logout as logoutApi,
    sendOTP,
    User,
    verifyOTP,
    VerifyOTPResponse
} from "../services/api/auth.api";

// TODO: 순서 Custom Hook (Context 사용 X)
// 필요시 AuthContext 생성, AuthProvider 로 분리

interface UseAuthReturn {
    isLoggedIn: boolean;
    user: User | null;
    loading: boolean;
    error: string | null;
    requestOtp: (email: string) => Promise<boolean>;
    verifyOtp: (otp: string) => Promise<boolean>;
    resendOtp: () => Promise<boolean>;
    logout: () => Promise<void>;
    clearError: () => void;
}

let tempEmail: string | null = null;

export const useAuth = (): UseAuthReturn => {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        (async () => {
            try {
                const user = getCurrentUser();
                if (user) {
                    setUser(user);
                }
            } catch (err) {
                // 인증되지 않은 상태는 에러로 처리하지 않음
                console.log("사용자 인증이 필요합니다.");
            } finally {
                setLoading(false);
            }
        })();
    }, []);
    /**
     * @return isNewUser
     */
    const requestOtp = async (email: string): Promise<boolean> => {
        setLoading(true);
        setError(null);

        try {
            const isNewUser = await sendOTP(email);
            tempEmail = email;
            return isNewUser;
        } catch (err: any) {
            const errorMessage = err?.response?.data?.error || "인증 코드 발송에 실패했습니다";
            setError(errorMessage);
            return false;
        } finally {
            setLoading(false);
        }
    };

    const resendOtp = async (): Promise<boolean> => {
        if (!tempEmail) return false;

        setLoading(true);
        setError(null);

        try {
            await sendOTP(tempEmail);
            return true;
        } catch (err: any) {
            const errorMessage = err?.response?.data?.error || "인증 코드 재전송에 실패했습니다";
            setError(errorMessage);
            return false;
        } finally {
            setLoading(false);
        }
    };
    const verifyOtp = async (otp: string): Promise<boolean> => {
        if (!tempEmail) return false;

        setLoading(true);
        setError(null);

        try {
            const response: VerifyOTPResponse = await verifyOTP(tempEmail, otp);
            if (response && response.user) {
                setUser(response.user);
                tempEmail = null;
                return true;
            } else {
                setError("인증에 실패했습니다");
                return false;
            }
        } catch (err: any) {
            const errorMessage = err?.response?.data?.error || "인증 과정에서 오류가 발생했습니다";
            setError(errorMessage);
            return false;
        } finally {
            setLoading(false);
        }
    };

    const logout = async (): Promise<void> => {
        setLoading(true);

        try {
            await logoutApi();
            setUser(null);
        } catch (err) {
            setError("로그아웃 중 오류가 발생했습니다");
        } finally {
            setLoading(false);
        }
    };

    const clearError = () => {
        setError(null);
    };

    return {
        isLoggedIn: !!user,
        user,
        loading,
        error,
        requestOtp,
        verifyOtp,
        resendOtp,
        logout,
        clearError,
    };
};
