import React, { useState } from 'react';
import './AccountDeleteModal.css';
import { useAuth } from '../../../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import { useToast } from '../../../hooks/useToast';
import api from "../../../services/api";

interface AccountDeleteModalProps {
    isOpen: boolean;
    onClose: () => void;
    userEmail?: string;
    lemonBalance?: number;
}

const AccountDeleteModal: React.FC<AccountDeleteModalProps> = ({
                                                                   isOpen,
                                                                   onClose,
                                                                   userEmail,
                                                                   lemonBalance = 0
                                                               }) => {
    const [confirmText, setConfirmText] = useState('');
    const [isDeleting, setIsDeleting] = useState(false);
    const { logout } = useAuth();
    const navigate = useNavigate();
    const { showToast } = useToast();

    if (!isOpen) return null;

    const handleDelete = async () => {
        if (confirmText !== 'DELETE') {
            showToast('확인 문구를 정확히 입력해주세요', 'error');
            return;
        }

        setIsDeleting(true);

        try {
            await api.auth.deleteAccount();

            // 임시 처리
            await new Promise(resolve => setTimeout(resolve, 2000));

            showToast('계정이 성공적으로 삭제되었습니다', 'success');
            await logout();
            navigate('/');
        } catch (error) {
            showToast('계정 삭제 중 오류가 발생했습니다', 'error');
            setIsDeleting(false);
        }
    };

    const handleBackdropClick = (e: React.MouseEvent) => {
        if (e.target === e.currentTarget && !isDeleting) {
            onClose();
        }
    };

    return (
        <div className="delete-account-modal-backdrop" onClick={handleBackdropClick}>
            <div className="delete-account-modal">
                <div className="delete-modal-header">
                    <h2 className="delete-modal-title">
                        계정 삭제 확인
                    </h2>
                    {!isDeleting && (
                        <button className="close-btn" onClick={onClose}>
                            ✕
                        </button>
                    )}
                </div>

                <div className="delete-modal-content">
                    <div className="warning-section">
                        <h3>다음 데이터가 영구적으로 삭제됩니다:</h3>
                        <ul className="deletion-list">
                            <li>
                                계정 정보: <strong>{userEmail}</strong>
                            </li>
                            <li>
                                보유 레몬: <strong>{lemonBalance.toLocaleString()} 개</strong>
                            </li>
                            <li>
                                모든 데이터베이스 인스턴스
                            </li>
                            <li>
                                모든 사용 기록 및 통계
                            </li>
                            <li>
                                획득한 업적 및 뱃지
                            </li>
                        </ul>
                    </div>

                    <div className="confirm-section">
                        <p className="confirm-instruction">
                            계정 삭제를 확인하려면 아래 입력란에 <strong>DELETE</strong>를 입력하세요:
                        </p>
                        <input
                            type="text"
                            className="confirm-input"
                            placeholder="DELETE 입력"
                            value={confirmText}
                            onChange={(e) => setConfirmText(e.target.value)}
                            disabled={isDeleting}
                        />
                    </div>

                    <div className="important-notice">
                        <strong>중요:</strong> 이 작업은 되돌릴 수 없습니다.
                        계정을 삭제하면 모든 데이터가 즉시 삭제되며,
                        동일한 이메일로 재가입하더라도 이전 데이터를 복구할 수 없습니다.
                    </div>
                </div>

                <div className="delete-modal-footer">
                    <button
                        className="cancel-btn"
                        onClick={onClose}
                        disabled={isDeleting}
                    >
                        취소
                    </button>
                    <button
                        className={`delete-btn ${confirmText === 'DELETE' ? 'enabled' : ''}`}
                        onClick={handleDelete}
                        disabled={confirmText !== 'DELETE' || isDeleting}
                    >
                        {isDeleting ? (
                            <>
                                <span className="spinner"></span>
                                삭제 중...
                            </>
                        ) : (
                            '계정 영구 삭제'
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default AccountDeleteModal;