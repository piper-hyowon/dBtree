import React from 'react';
import './DeleteModal.css';
import {characterImages} from '../../../utils/characterImages';

interface DeleteModalProps {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    itemName: string;
    itemType?: string;
}

const DeleteModal: React.FC<DeleteModalProps> = ({
                                                     isOpen,
                                                     onClose,
                                                     onConfirm,
                                                     itemName,
                                                     itemType
                                                 }) => {
    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="delete-modal" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header">
                    <img
                        src={characterImages.error}
                        alt="Warning"
                        className="modal-character"
                    />
                    <h3>정말로 삭제하시겠습니까?</h3>
                </div>

                <div className="modal-content">
                    <p className="modal-warning">
                        <strong>'{itemName}'</strong> {itemType}를 삭제하면 복구할 수 없습니다.
                    </p>
                    <p className="modal-info">
                        이 작업은 영구적이며, 모든 데이터가 손실됩니다.
                    </p>
                </div>

                <div className="modal-actions">
                    <button className="modal-btn cancel" onClick={onClose}>
                        취소
                    </button>
                    <button className="modal-btn confirm" onClick={onConfirm}>
                        삭제
                    </button>
                </div>
            </div>
        </div>
    );
};

export default DeleteModal;