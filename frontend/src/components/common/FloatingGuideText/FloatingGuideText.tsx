import React, {useState} from 'react';
import './FloatingGuideText.css';

interface FloatingGuideTextProps {
    text?: string;
    emoji?: string;
    position?: 'top' | 'bottom' | 'left' | 'right';
    variant?: 'default' | 'golden' | 'subtle';
    className?: string;
    dismissible?: boolean;
    onDismiss?: () => void;
}

const FloatingGuideText: React.FC<FloatingGuideTextProps> = ({
                                                                 text,
                                                                 emoji ,
                                                                 position = 'top',
                                                                 variant = 'default',
                                                                 className = '',
                                                                 dismissible = true,
                                                                 onDismiss
                                                             }) => {
    const [isVisible, setIsVisible] = useState(true);

    const handleClick = () => {
        if (dismissible) {
            setIsVisible(false);
            onDismiss?.();
        }
    };

    if (!isVisible) return null;

    const positionClasses = {
        top: 'floating-text-top',
        bottom: 'floating-text-bottom',
        left: 'floating-text-left',
        right: 'floating-text-right'
    };

    const variantClasses = {
        default: 'floating-text-default',
        golden: 'floating-text-golden',
        subtle: 'floating-text-subtle'
    };

    return (
        <div
            className={`
        floating-guide-text 
        ${positionClasses[position]} 
        ${variantClasses[variant]}
        ${dismissible ? 'floating-text-dismissible' : ''}
        ${className}
      `}
            onClick={handleClick}
        >
            <div className="floating-text-content">
                <span className="floating-text-emoji">{emoji}</span>
                <span className="floating-text-message">{text}</span>
                {dismissible && (
                    <span className="floating-text-close">×</span>
                )}
            </div>
        </div>
    );
};

export default FloatingGuideText;