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
    requestOtp: (email: string) => Promise<{ success: boolean, isNewUser?: boolean, message?: string, retryAfter?: number }>;
    verifyOtp: (otp: string) => Promise<{ success: boolean, isNewUser?: boolean, message?: string, retryAfter?: number }>;
    resendOtp: () => Promise<{ success: boolean, isNewUser?: boolean, message?: string, retryAfter?: number }>;
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
        const requestOtp = async (email: string): Promise<{ success: boolean, isNewUser?: boolean, message?: string, retryAfter?: number }> => {
            setLoading(true);
            setError(null);

            try {
                const isNewUser = await sendOTP(email);
                tempEmail = email;
                return {success: true, isNewUser};
            } catch (err: any) {
                const errorMessage = err?.response?.data?.error || "인증 코드 발송에 실패했습니다";
                const retryAfter = err?.response?.headers?.['retry-after'];
                const waitSeconds = retryAfter ? parseInt(retryAfter, 10) : undefined;

                setError(errorMessage);
                return {
                    success: false,
                    message: errorMessage,
                    retryAfter: waitSeconds && !isNaN(waitSeconds) ? waitSeconds : undefined
                };

            } finally {
                setLoading(false);
            }
        };

        const resendOtp = async (): Promise<{ success: boolean, message?: string, retryAfter?: number }> => {
            if (!tempEmail) return {success: false, message: "이메일 정보가 없습니다."};

            setLoading(true);
            setError(null);

            try {
                await sendOTP(tempEmail);
                return {success: true};
            } catch (err: any) {
                const errorMessage = err?.response?.data?.error || "인증 코드 재전송에 실패했습니다";
                const retryAfter = err.retryAfter || undefined;
                console.log("Resend OTP error:", errorMessage, "Retry after:", retryAfter);

                setError(errorMessage);

                return {
                    success: false,
                    message: errorMessage,
                    retryAfter                };
            } finally {
                setLoading(false);
            }
        };
        const verifyOtp = async (otp: string): Promise<{ success: boolean, message?: string }> => {
            if (!tempEmail) return {success: false, message: "이메일 정보가 없습니다."};

            setLoading(true);
            setError(null);


            try {
                const response: VerifyOTPResponse = await verifyOTP(tempEmail, otp);
                if (response && response.user) {
                    setUser(response.user);
                    tempEmail = null;
                    return {success: true};
                } else {
                    const msg = "인증에 실패했습니다";
                    setError(msg);
                    return {success: false, message: msg};
                }
            } catch (err: any) {
                const errorMessage = err?.response?.data?.error || "인증 과정에서 오류가 발생했습니다";
                setError(errorMessage);
                return {success: false, message: errorMessage};
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
    }
;
