import React, {useState} from 'react';
import './FloatingSupportButton.css';
import SupportModal from './SupportModal';
import PigeonFloatingIcon from '../../assets/images/pigeon_floating.png'

const FloatingSupportButton: React.FC = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isHovered, setIsHovered] = useState(false);

    return (
        <>
            <div
                className="floating-support-button"
                onMouseEnter={() => setIsHovered(true)}
                onMouseLeave={() => setIsHovered(false)}
                onClick={() => setIsModalOpen(true)}
            >
                <img src={PigeonFloatingIcon} width={90} height={90}/>
                {isHovered && (
                    <span className="support-tooltip">도움이 필요하신가요?</span>
                )}
            </div>

            <SupportModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
            />
        </>
    );
};

export default FloatingSupportButton;
