import React, { useState, useEffect } from 'react';
import './SecurityNotice.css';

const SecurityNotice: React.FC = () => {
    const [isVisible, setIsVisible] = useState(false);

    useEffect(() => {
        setTimeout(() => setIsVisible(true), 500);

        const handleEsc = (e: KeyboardEvent) => {
            if (e.key === 'Escape') setIsVisible(false);
        };

        document.addEventListener('keydown', handleEsc);
        return () => document.removeEventListener('keydown', handleEsc);
    }, []);

    if (!isVisible) return null;

    return (
        <div className="security-notice-overlay" onClick={() => setIsVisible(false)}>
            <div className="security-notice-modal" onClick={(e) => e.stopPropagation()}>
                <div className="notice-header">
                    <h3>보안 안내</h3>
                    <span className="badge-beta">BETA</span>
                </div>

                <div className="notice-content">
                    <p className="main-text">
                        dBtree는 학습 및 테스트 목적의 DBaaS 플랫폼입니다.
                    </p>

                    <div className="warning-box">
                        <strong>중요: 프로덕션 데이터 저장 금지</strong>
                        <ul>
                            <li>현재 베타 버전 한정</li>
                            {/*<li>K8s Secret은 Base64 인코딩만 적용 (암호화 X)</li>*/}
                            {/*<li>PVC는 Local Path Provisioner 사용 (암호화 X)</li>*/}
                            {/*<li>NodePort 30000-31999 범위 직접 노출</li>*/}
                        </ul>
                    </div>

                    <div className="commitment">
                        <p>
                            사용자 데이터에 접근하지 않으며,<br/>
                            계속해서 보안을 강화하고 있습니다.
                        </p>
                        <a href="/security" target="_blank">보안 로드맵 (업데이트 예정) →</a>
                    </div>
                </div>

                <button className="close-button-sn" onClick={() => setIsVisible(false)}>
                    확인 (ESC)
                </button>
            </div>
        </div>
    );
};

export default SecurityNotice;