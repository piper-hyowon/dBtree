import React, {useState} from 'react';
import {InstanceResponse} from '../../types/database.types';
import {useToast} from '../../hooks/useToast';
import './CredentialsModal.css';

interface CredentialsModalProps {
    instance: InstanceResponse;
    onClose: () => void;
}

const CredentialsModal: React.FC<CredentialsModalProps> = ({instance, onClose}) => {
    const {showToast} = useToast();
    const [copiedField, setCopiedField] = useState<string | null>(null);

    // credentials는 생성 응답에만 포함됨
    const credentials = (instance as any).credentials;

    const copyToClipboard = (text: string, field: string) => {
        navigator.clipboard.writeText(text);
        setCopiedField(field);
        showToast('클립보드에 복사되었습니다', 'success', 2000);

        setTimeout(() => {
            setCopiedField(null);
        }, 2000);
    };

    if (!credentials) {
        return (
            <div className="modal-overlay" onClick={onClose}>
                <div className="modal-content credentials-modal" onClick={(e) => e.stopPropagation()}>
                    <div className="modal-header">
                        <h2>인스턴스 생성 완료</h2>
                    </div>
                    <div className="modal-body">
                        <p>인스턴스가 성공적으로 생성되었습니다.</p>
                        <p>상태: {instance.status}</p>
                    </div>
                    <div className="modal-footer">
                        <button className="btn-primary" onClick={onClose}>
                            확인
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content credentials-modal" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>🎉 인스턴스 생성 완료!</h2>
                    <button className="modal-close" onClick={onClose}>×</button>
                </div>

                <div className="modal-body">
                    <div className="credentials-warning">
                        <div className="warning-icon">⚠️</div>
                        <div className="warning-content">
                            <strong>중요!</strong> 이 정보는 다시 확인할 수 없습니다.
                            <br/>아래 인증 정보를 안전한 곳에 저장해주세요.
                        </div>
                    </div>

                    <div className="credentials-section">
                        <h3>기본 정보</h3>
                        <div className="credential-item">
                            <span className="credential-label">인스턴스 ID</span>
                            <div className="credential-value">
                                <code>{instance.id}</code>
                                <button
                                    className={`copy-btn ${copiedField === 'id' ? 'copied' : ''}`}
                                    onClick={() => copyToClipboard(instance.id, 'id')}
                                >
                                    {copiedField === 'id' ? '✓' : '📋'}
                                </button>
                            </div>
                        </div>

                        <div className="credential-item">
                            <span className="credential-label">인스턴스 이름</span>
                            <div className="credential-value">
                                <code>{instance.name}</code>
                            </div>
                        </div>

                        <div className="credential-item">
                            <span className="credential-label">상태</span>
                            <div className="credential-value">
                                <span className="status-badge provisioning">
                                    {instance.status === 'provisioning' ? '프로비저닝 중...' : instance.status}
                                </span>
                            </div>
                        </div>
                    </div>

                    <div className="credentials-section">
                        <h3>인증 정보</h3>
                        <div className="credential-item">
                            <span className="credential-label">사용자명</span>
                            <div className="credential-value">
                                <code>{credentials.username}</code>
                                <button
                                    className={`copy-btn ${copiedField === 'username' ? 'copied' : ''}`}
                                    onClick={() => copyToClipboard(credentials.username, 'username')}
                                >
                                    {copiedField === 'username' ? '✓' : '📋'}
                                </button>
                            </div>
                        </div>

                        <div className="credential-item">
                            <span className="credential-label">비밀번호</span>
                            <div className="credential-value">
                                <code className="password">{credentials.password}</code>
                                <button
                                    className={`copy-btn ${copiedField === 'password' ? 'copied' : ''}`}
                                    onClick={() => copyToClipboard(credentials.password, 'password')}
                                >
                                    {copiedField === 'password' ? '✓' : '📋'}
                                </button>
                            </div>
                        </div>
                    </div>

                    {credentials.externalHost && credentials.externalPort && (
                        <div className="credentials-section">
                            <h3>연결 정보</h3>
                            <div className="credential-item">
                                <span className="credential-label">호스트</span>
                                <div className="credential-value">
                                    <code>{credentials.externalHost}</code>
                                    <button
                                        className={`copy-btn ${copiedField === 'host' ? 'copied' : ''}`}
                                        onClick={() => copyToClipboard(credentials.externalHost, 'host')}
                                    >
                                        {copiedField === 'host' ? '✓' : '📋'}
                                    </button>
                                </div>
                            </div>

                            <div className="credential-item">
                                <span className="credential-label">포트</span>
                                <div className="credential-value">
                                    <code>{credentials.externalPort}</code>
                                    <button
                                        className={`copy-btn ${copiedField === 'port' ? 'copied' : ''}`}
                                        onClick={() => copyToClipboard(credentials.externalPort.toString(), 'port')}
                                    >
                                        {copiedField === 'port' ? '✓' : '📋'}
                                    </button>
                                </div>
                            </div>

                            {credentials.externalUri && (
                                <div className="credential-item">
                                    <span className="credential-label">연결 문자열</span>
                                    <div className="credential-value">
                                        <code className="connection-string">{credentials.externalUri}</code>
                                        <button
                                            className={`copy-btn ${copiedField === 'uri' ? 'copied' : ''}`}
                                            onClick={() => copyToClipboard(credentials.externalUri, 'uri')}
                                        >
                                            {copiedField === 'uri' ? '✓' : '📋'}
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}

                    <div className="credentials-info">
                        <p>
                            💡 인스턴스가 현재 <strong>프로비저닝 중</strong>입니다.
                            완전히 준비되기까지 몇 분이 소요될 수 있습니다.
                        </p>
                        <p>
                            대시보드에서 인스턴스 상태를 확인할 수 있습니다.
                        </p>
                    </div>
                </div>

                <div className="modal-footer">
                    <button className="btn-primary" onClick={onClose}>
                        대시보드로 이동
                    </button>
                </div>
            </div>
        </div>
    );
};

export default CredentialsModal;