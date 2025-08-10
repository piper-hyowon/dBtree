import React, {useEffect} from 'react';
import './Toast.css';
import {characterImages, isImageComponent} from '../../../utils/characterImages';

export type ToastType = 'info' | 'success' | 'warning' | 'error';

interface ToastProps {
    message: string;
    type?: ToastType;
    duration?: number;
    onClose: () => void;
}

const toastConfig = {
    info: {
        character: characterImages.loading,
        className: 'toast-info'
    },
    success: {
        character: characterImages.backupComplete,
        className: 'toast-success'
    },
    warning: {
        character: characterImages.maintenance,
        className: 'toast-warning'
    },
    error: {
        character: characterImages.error,
        className: 'toast-error'
    },
};

const Toast: React.FC<ToastProps> = ({
                                         message,
                                         type = 'info',
                                         duration = 3000,
                                         onClose
                                     }) => {
    useEffect(() => {
        const timer = setTimeout(() => {
            onClose();
        }, duration);

        return () => clearTimeout(timer);
    }, [duration, onClose]);

    const config = toastConfig[type];

    return (
        <div className={`toast ${config.className}`}>

            {
                isImageComponent(config.character) ? (
                    <div className="character-svg-wrapper">
                        {React.createElement(config.character)}
                    </div>
                ) : (
                    <img
                        src={config.character as string}
                        alt={type}
                        className="db-main-character"
                    />
                )
            }
            <span className="toast-message">{message}</span>
            <button className="toast-close" onClick={onClose}>
                ×
            </button>
        </div>
    );
};

export default Toast;

