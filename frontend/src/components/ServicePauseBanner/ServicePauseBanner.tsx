import React, { useState } from "react";
import "./ServicePauseBanner.css";
import sadTree from "../../assets/images/character/sleepy_pause.png";

const ServicePauseBanner: React.FC = () => {
    const [copied, setCopied] = useState(false);
    const email = "support@dbtree.cloud";

    const handleEmailClick = () => {
        navigator.clipboard.writeText(email).then(() => {
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        });
    };

    return (
        <div className="service-pause-banner">
            <div className="banner-content">
                <img src={sadTree} alt="슬픈 디비트리" className="banner-image" />
                <div className="banner-right">
                    <div className="banner-text-section">
                        <h3 className="banner-title">🍋 디비트리의 겨울잠</h3>
                        <p className="banner-main-text">
                            개인 프로젝트로 운영되는 우리 작은 레몬 농장이 서버 비용 문제로 잠시 휴식을 취하게 되었어요.
                        </p>
                        <p className="banner-sub-text">
                            에너지를 충전하고 더 나은 모습으로 돌아올게요!
                        </p>
                    </div>
                    <div className="email-section">
                        <div className="notification-box">
                            <div className="notification-icon">🔔</div>
                            <div className="notification-content">
                                <p className="notification-title">재오픈 알림 받기</p>
                                <p className="notification-description">
                                    디비트리가 돌아올 때 알림을 받고 싶으시다면
                                </p>
                                <button onClick={handleEmailClick} className="email-button">
                                    {copied ? "✓ 복사됨!" : "support@dbtree.cloud"}
                                </button>
                                <p className="notification-hint">
                                    {copied ? "이메일 주소가 복사되었어요" : "클릭하면 이메일이 복사돼요"}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ServicePauseBanner;