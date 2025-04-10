import React, { useState } from "react";
import { useAuth } from "../../hooks/useAuth";
import dbtreeLogo from "../../assets/images/dbtree_logo.svg";
import EmailForm from "./EmailForm";
import OtpForm from "./OtpForm";

interface LoginModalProps {
  onClose: () => void;
}

const LoginModal: React.FC<LoginModalProps> = ({ onClose }) => {
  const [email, setEmail] = useState("");
  const [authStep, setAuthStep] = useState<"email" | "otp">("email");
  const [error, setError] = useState<string | null>(null);

  const setErrorMessage = (message: string | null) => {
    setError(message);
  };

  const handleOtpRequested = async (emailValue: string) => {
    setEmail(emailValue);
    setAuthStep("otp");
  };

  const handleAuthSuccess = () => {
    onClose();
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="login-modal" onClick={(e) => e.stopPropagation()}>
        <button className="close-button" onClick={onClose}>
          ×
        </button>
        <img src={dbtreeLogo} alt="dBtree Logo" className="logo" />

        <h2 className="modal-title">
          {authStep === "email" ? "시작하기" : "인증하기"}
        </h2>

        {error && <div className="error-message">{error}</div>}

        {authStep === "email" ? (
          <EmailForm
            onOtpRequested={handleOtpRequested}
            setError={setErrorMessage}
          />
        ) : (
          <OtpForm
            email={email}
            onAuthSuccess={handleAuthSuccess}
            setError={setErrorMessage}
          />
        )}
      </div>
    </div>
  );
};

export default LoginModal;
