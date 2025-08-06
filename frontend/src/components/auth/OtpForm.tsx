import React, {useEffect, useState} from "react";
import {useAuth} from "../../contexts/AuthContext";

const OTP_LENGTH = 6;

interface OtpFormProps {
    email: string;
    onAuthSuccess: () => void;
    setError: (error: string | null) => void;
}

const OtpForm: React.FC<OtpFormProps> = ({
                                             email,
                                             onAuthSuccess,
                                             setError,
                                         }) => {
    const [otp, setOtp] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [resendCountdown, setResendCountdown] = useState(0);
    const {verifyOtp, resendOtp} = useAuth();

    useEffect(() => {
        let timer: NodeJS.Timeout | null = null;

        if (resendCountdown > 0) {
            timer = setInterval(() => {
                setResendCountdown(prev => prev - 1);
            }, 1000);
        }

        return () => {
            if (timer) clearInterval(timer);
        };
    }, [resendCountdown]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        e.stopPropagation();

        if (!otp || otp.length !== OTP_LENGTH) {
            setError(`유효한 ${OTP_LENGTH}자리 인증 코드를 입력해주세요.`);
            return;
        }

        setIsLoading(true);
        setError(null);

        try {
            const response = await verifyOtp(otp);
            if (response.success) {
                onAuthSuccess();
            } else {
                setError(response.message || "잘못된 인증 코드입니다. 다시 확인해주세요.");
                setOtp("");
            }
        } catch (err: any) {
            setError(err?.message || "인증 과정에서 오류가 발생했습니다.");
            setOtp("");
        } finally {
            setIsLoading(false);
        }
    };

    const handleResendOtp = async (e: React.MouseEvent) => {
        e.stopPropagation();

        if (resendCountdown > 0) return;

        setIsLoading(true);
        setError(null);

        try {
            const result = await resendOtp();

            if (result.success) {
                setError("인증 코드가 재전송되었습니다.");
                setOtp("");
            } else {
                // 재시도 시간이 있는 경우
                if (result.retryAfter && result.retryAfter > 0) {
                    setResendCountdown(result.retryAfter);
                    setError(`${result.retryAfter}초 후에 다시 시도할 수 있습니다.`);
                } else {
                    setError(result.message || "재전송에 실패했습니다.");
                }
            }
        } catch (err) {
            setError("재전송 과정에서 오류가 발생했습니다.");
        } finally {
            setIsLoading(false);
        }
    };

    const formatTime = (seconds: number): string => {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${mins > 0 ? `${mins}분 ` : ''}${secs}초`;
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="email-display">
                <p>{email}로 인증 코드가 발송되었습니다.</p>
            </div>

            <div className="input-group">
                <input
                    type="text"
                    value={otp}
                    onChange={(e) => setOtp(e.target.value.replace(/[^0-9]/g, ""))}
                    placeholder={`${OTP_LENGTH}자리 인증 코드`}
                    maxLength={OTP_LENGTH}
                    className="otp-input"
                    disabled={isLoading}
                />
            </div>

            <button
                className="cta-button modal-button"
                type="submit"
                disabled={isLoading || otp.length !== OTP_LENGTH}
                onClick={(e) => e.stopPropagation()}
            >
                {isLoading ? "인증 중..." : "인증하기"}
            </button>

            <button
                className="resend-button"
                type="button"
                onClick={handleResendOtp}
                disabled={isLoading || resendCountdown > 0}
                style={{
                    opacity: resendCountdown > 0 ? 0.7 : 1,
                    cursor: resendCountdown > 0 ? 'not-allowed' : 'pointer',
                    color: resendCountdown > 0 ? '#ff6b6b' : '#3498db',
                    fontWeight: resendCountdown > 0 ? 'bold' : 'normal'
                }}
            >
                {resendCountdown > 0
                    ? `인증 코드 재전송 (${formatTime(resendCountdown)} 후 가능)`
                    : "인증 코드 재전송"}
            </button>
        </form>
    );
};

export default OtpForm;