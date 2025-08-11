import React, {useState} from 'react';
import './SupportModal.css';
import {useAuth} from '../../contexts/AuthContext';
import api from '../../services/api';
import {useToast} from '../../hooks/useToast';
import PigeonModalIllustration from '../../assets/images/pigeon_modal.png';

interface SupportModalProps {
    isOpen: boolean;
    onClose: () => void;
}

type SupportCategory = 'bug' | 'feature' | 'billing' | 'other';

const SupportModal: React.FC<SupportModalProps> = ({isOpen, onClose}) => {
    const {user} = useAuth();
    const {showToast} = useToast();

    const [category, setCategory] = useState<SupportCategory>('other');
    const [subject, setSubject] = useState('');
    const [message, setMessage] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!subject.trim() || !message.trim()) {
            showToast('제목과 내용을 모두 입력해주세요', 'error');
            return;
        }

        try {
            setIsSubmitting(true);

            // 사용자 컨텍스트
            const context = {
                userEmail: user?.email,
                lemonBalance: user?.lemonBalance,
                timestamp: new Date().toISOString(),
                userAgent: navigator.userAgent,
                currentPage: window.location.pathname,
            };

            await api.support.sendInquiry({
                category,
                subject,
                message,
                context
            });

            showToast('문의가 성공적으로 전송되었습니다!', 'success');

            // 폼 리셋
            setSubject('');
            setMessage('');
            setCategory('other');
            onClose();

        } catch (error) {
            showToast('문의 전송에 실패했습니다. 잠시 후 다시 시도해주세요.', 'error');
        } finally {
            setIsSubmitting(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="support-modal-overlay" onClick={onClose}>
            <div className="support-modal" onClick={(e) => e.stopPropagation()}>
                <div className="support-modal-illustration">
                    <img
                        src={PigeonModalIllustration}
                        alt="Customer Support"
                        className="illustration-image"
                    />
                    <div className="illustration-overlay">
                        <h2>무엇을 도와드릴까요?</h2>
                    </div>
                </div>

                <div className="support-modal-header">
                    <button className="close-btn" onClick={onClose}>×</button>
                </div>

                <form onSubmit={handleSubmit} className="support-form">
                    <div className="form-group">
                        <label htmlFor="category">문의 유형</label>
                        <select
                            id="category"
                            value={category}
                            onChange={(e) => setCategory(e.target.value as SupportCategory)}
                        >
                            <option value="bug">버그 신고</option>
                            <option value="feature">기능 제안</option>
                            <option value="billing">레몬 관련</option>
                            <option value="other">기타 문의</option>
                        </select>
                    </div>

                    <div className="form-group">
                        <label htmlFor="subject">제목</label>
                        <input
                            id="subject"
                            type="text"
                            value={subject}
                            onChange={(e) => setSubject(e.target.value)}
                            placeholder="문의 제목을 입력해주세요"
                            maxLength={100}
                        />
                    </div>

                    <div className="form-group">
                        <label htmlFor="message">내용</label>
                        <textarea
                            id="message"
                            value={message}
                            onChange={(e) => setMessage(e.target.value)}
                            placeholder="문의 내용을 자세히 작성해주세요"
                            rows={6}
                            maxLength={1000}
                        />
                        <div className="char-count">{message.length}/1000</div>
                    </div>

                    <div className="support-info">
                        <p>• 답변은 가입하신 이메일로 24시간 내 발송됩니다</p>
                        <p>• 문의 내용과 함께 현재 상태 정보가 자동으로 전송됩니다</p>
                    </div>

                    <div className="modal-actions">
                        <button
                            type="button"
                            className="cancel-btn"
                            onClick={onClose}
                            disabled={isSubmitting}
                        >
                            취소
                        </button>
                        <button
                            type="submit"
                            className="submit-btn"
                            disabled={isSubmitting}
                        >
                            {isSubmitting ? '전송 중...' : '문의 전송'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default SupportModal;