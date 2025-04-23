import React, {useState} from "react";
import {useAuth} from "../../hooks/useAuth";

interface EmailFormProps {
    onOtpRequested: (email: string, isNewUser: boolean) => void;
    setError: (error: string | null) => void;
}

const EmailForm: React.FC<EmailFormProps> = ({onOtpRequested, setError}) => {
    const [email, setEmail] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const {requestOtp} = useAuth();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!email || !email.includes("@")) {
            setError("유효한 이메일을 입력해주세요.");
            return;
        }

        setIsLoading(true);
        setError(null);

        try {
            const result = await requestOtp(email);

            if (result.success && result.isNewUser !== undefined) {
                onOtpRequested(email, result.isNewUser);
            } else {
                if (result.retryAfter && result.retryAfter > 0) {
                    setError(`${result.retryAfter}초 후에 다시 시도할 수 있습니다.`);
                } else {
                    setError(result.message || "인증 코드 발송에 실패했습니다.");
                }
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (

        <form onSubmit={handleSubmit}>
            <div className="input-group">
                <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="이메일 입력"
                    className="email-input"
                    disabled={isLoading}
                />
            </div>

            <button
                className="cta-button modal-button"
                type="submit"
                disabled={isLoading}
            >
                {isLoading ? "처리 중..." : "인증 코드 받기"}
            </button>
        </form>
    );
};

export default EmailForm;
