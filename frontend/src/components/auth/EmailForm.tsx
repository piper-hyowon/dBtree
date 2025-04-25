import React, {useState} from "react";
import {useAuth} from "../../hooks/useAuth";
import "./EmailForm.css"

interface EmailFormProps {
    onOtpRequested: (email: string, isNewUser: boolean) => void;
    setError: (error: string | null) => void;
}

const EmailForm: React.FC<EmailFormProps> = ({onOtpRequested, setError}) => {
    const [emailUsername, setEmailUsername] = useState("");
    const [customDomain, setCustomDomain] = useState("");
    const [selectedDomain, setSelectedDomain] = useState("gmail.com");
    const [isCustomDomain, setIsCustomDomain] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const {requestOtp} = useAuth();

    const commonDomains = [
        "gmail.com",
        "naver.com",
        "daum.net",
        "kakao.com",
        "hotmail.com",
        "outlook.com",
        "yahoo.com",
        "직접 입력"
    ];

    const getFullEmail = () => {
        const domain = isCustomDomain ? customDomain : selectedDomain;
        return emailUsername && domain ? `${emailUsername}@${domain}` : "";
    };

    const validateEmail = (email: string) => {
        const regex = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        return regex.test(email);
    };

    const handleDomainChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        const value = e.target.value;
        if (value === "직접 입력") {
            setIsCustomDomain(true);
            setCustomDomain("");
        } else {
            setIsCustomDomain(false);
            setSelectedDomain(value);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const fullEmail = getFullEmail();

        if (!fullEmail) {
            setError("이메일을 입력해주세요.");
            return;
        }

        if (!validateEmail(fullEmail)) {
            setError("유효한 이메일 형식이 아닙니다.");
            return;
        }

        setIsLoading(true);
        setError(null);

        try {
            const result = await requestOtp(fullEmail);

            if (result.success && result.isNewUser !== undefined) {
                onOtpRequested(fullEmail, result.isNewUser);
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
            <div className="input-group email-input-group">
                <div className="email-input-container">
                    <input
                        type="text"
                        value={emailUsername}
                        onChange={(e) => setEmailUsername(e.target.value)}
                        placeholder="이메일 아이디"
                        className="email-username-input"
                        disabled={isLoading}
                    />
                    <span className="email-at">@</span>
                    {isCustomDomain ? (
                        <input
                            type="text"
                            value={customDomain}
                            onChange={(e) => setCustomDomain(e.target.value)}
                            placeholder="도메인 직접 입력"
                            className="email-domain-input"
                            disabled={isLoading}
                        />
                    ) : (
                        <select
                            value={selectedDomain}
                            onChange={handleDomainChange}
                            className="email-domain-select"
                            disabled={isLoading}
                        >
                            {commonDomains.map((domain) => (
                                <option key={domain} value={domain}>
                                    {domain}
                                </option>
                            ))}
                        </select>
                    )}
                </div>
                {isCustomDomain && (
                    <div className="custom-domain-actions">
                        <button
                            type="button"
                            className="back-to-select-btn"
                            onClick={() => {
                                setIsCustomDomain(false);
                                setSelectedDomain("gmail.com");
                            }}
                            disabled={isLoading}
                        >
                            ← 목록으로
                        </button>
                    </div>
                )}
            </div>

            <button
                className="cta-button modal-button"
                type="submit"
                disabled={isLoading || !emailUsername || (isCustomDomain && !customDomain)}
            >
                {isLoading ? "처리 중..." : "인증 코드 받기"}
            </button>
        </form>
    );
};

export default EmailForm;