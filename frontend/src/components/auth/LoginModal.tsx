import React, {useState} from "react";
import dbtreeLogo from "../../assets/images/dbtree_logo.svg";
import EmailForm from "./EmailForm";
import OtpForm from "./OtpForm";
import CharacterNewUser from "../../assets/images/character/new-user.svg";
import CharaterReturningUser from "../../assets/images/character/returning-user.svg";
import "./LoginModal.css";

interface LoginModalProps {
    onClose: () => void;
}

const LoginModal: React.FC<LoginModalProps> = ({onClose}) => {
    const [email, setEmail] = useState("");
    const [authStep, setAuthStep] = useState<"email" | "otp">("email");
    const [error, setError] = useState<string | null>(null);
    const [isNewUser, setIsNewUser] = useState<boolean | null>(null);

    const setErrorMessage = (message: string | null) => {
        setError(message);
        if (message) {
            setTimeout(() => {
                setError(null);
            }, 5000);
        }
    };

    const handleOtpRequested = async (emailValue: string, newUser: boolean) => {
        setEmail(emailValue);
        setIsNewUser(newUser);
        setAuthStep("otp");
    };

    const handleAuthSuccess = () => {
        onClose();
    };

    const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
        if (e.target === e.currentTarget) {
            onClose();
        }
    };

    return (
        <div className="modal-overlay" onClick={handleOverlayClick}>
            <div className="login-modal" onClick={(e) => e.stopPropagation()}>
                <button className="close-button" onClick={onClose}>
                    ×
                </button>
                {authStep === "otp" && isNewUser !== null ? (
                    <div style={{
                        textAlign: 'center',
                        marginBottom: '1.5rem',
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center'
                    }}>
                        <div style={{
                            width: '100%',
                            maxWidth: '550px',
                            marginBottom: '1rem'
                        }}>
                            <img
                                src={isNewUser ? CharacterNewUser : CharaterReturningUser}
                                alt={isNewUser ? "CharacterNewUser" : "CharaterReturningUser"}
                                style={{
                                    width: '100%',
                                    height: 'auto'
                                }}
                            />
                        </div>
                        <h2 className="modal-title">
                            {isNewUser ? "Welcome!" : "Welcome back!"}
                        </h2>
                        <p className="modal-subtitle">
                            {isNewUser ? "처음 오셨네요!" : "또 오셨네요, 반가워요!"}
                        </p>
                    </div>
                ) : (
                    <>
                        <img src={dbtreeLogo} alt="dBtree Logo" className="logo"/>
                        <h2 className="modal-title">
                            {authStep === "email" ? "시작하기" : "인증하기"}
                        </h2>
                    </>
                )}

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