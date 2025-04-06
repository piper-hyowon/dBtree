import React, { useState } from "react";
import { useAuth } from "../../hooks/useAuth";

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
  const { verifyOtp, requestOtp } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!otp || otp.length !== 6) {
      setError("유효한 6자리 인증 코드를 입력해주세요.");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const success = await verifyOtp(otp);
      if (success) {
        onAuthSuccess();
      } else {
        setError("잘못된 인증 코드입니다. 다시 확인해주세요.");
      }
    } catch (err) {
      setError("인증 과정에서 오류가 발생했습니다.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendOtp = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const success = await requestOtp(email);
      if (success) {
        setError("인증 코드가 재전송되었습니다.");
      } else {
        setError("인증 코드 재전송에 실패했습니다.");
      }
    } catch (err) {
      setError("서버 오류가 발생했습니다.");
    } finally {
      setIsLoading(false);
    }
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
          placeholder="6자리 인증 코드"
          maxLength={6}
          className="otp-input"
          disabled={isLoading}
        />
      </div>

      <button
        className="cta-button modal-button"
        type="submit"
        disabled={isLoading || otp.length !== 6}
      >
        {isLoading ? "인증 중..." : "인증하기"}
      </button>

      <button
        className="resend-button"
        type="button"
        onClick={handleResendOtp}
        disabled={isLoading}
      >
        인증 코드 재전송
      </button>
    </form>
  );
};

export default OtpForm;
