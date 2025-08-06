import React, {useState} from 'react';
import {PresetResponse} from '../../types/database.types';
import TechnicalTermHighlighter from './TechnicalTermHighlighter';
import './PresetCard.css';

const ExpandableDescription: React.FC<{
    description: string;
    terms: Record<string, string>;
}> = ({description, terms}) => {
    const [isExpanded, setIsExpanded] = useState(false);

    return (
        <div className="expandable-description">
            <button
                className="expand-toggle"
                onClick={(e) => {
                    e.stopPropagation();
                    setIsExpanded(!isExpanded);
                }}
            >
                <span className="toggle-icon">{isExpanded ? '📖' : '📚'}</span>
                <span className="toggle-text">
                    {isExpanded ? '간단히 보기' : '자세히 보기'}
                </span>
                <span className="toggle-arrow">{isExpanded ? '▲' : '▼'}</span>
            </button>

            {isExpanded && (
                <div className="preset-friendly-description expanded">
                    <TechnicalTermHighlighter
                        text={description}
                        terms={terms}
                        className="friendly-description-text"
                    />
                </div>
            )}
        </div>
    );
};

interface PresetCardProps {
    preset: PresetResponse;
    selected: boolean;
    onSelect: () => void;
}

const PresetCard: React.FC<PresetCardProps> = ({preset, selected, onSelect}) => {
    const getModeDisplayName = (mode: string): string => {
        switch (mode) {
            case 'standalone':
                return 'Standalone';
            case 'replica_set':
                return 'Replica Set';
            case 'sharded':
                return 'Sharded Cluster';
            default:
                return mode;
        }
    };

    const getSizeColor = (size: string): string => {
        const sizeMap: Record<string, string> = {
            'small': 'size-small',
            'medium': 'size-medium',
            'large': 'size-large'
        };
        return sizeMap[size.toLowerCase()] || 'size-default';
    };

    return (
        <div
            className={`preset-card ${selected ? 'selected' : ''}`}
            onClick={onSelect}
        >
            {selected && (
                <div className="preset-selected-indicator">
                    <span className="checkmark">✓</span>
                </div>
            )}

            <div className="preset-header">
                <div className="preset-icon-wrapper">
                    <span className="preset-icon">{preset.icon}</span>
                </div>
                <div className="preset-title-section">
                    <h4 className="preset-title">{preset.name}</h4>
                    <div className="preset-badges">
                        <span className={`size-badge ${getSizeColor(preset.size)}`}>
                            {preset.size.toUpperCase()}
                        </span>
                        <span className="mode-badge">
                            {getModeDisplayName(preset.mode)}
                        </span>
                    </div>
                </div>
            </div>

            {/* Description with technical terms highlighted */}
            <div className="preset-description">
                <TechnicalTermHighlighter
                    text={preset.description}
                    terms={preset.technicalTerms || {}}
                    className="description-text"
                />
            </div>

            {/* Collapsible Friendly description with technical terms highlighted */}
            {preset.friendlyDescription && (
                <ExpandableDescription
                    description={preset.friendlyDescription}
                    terms={preset.technicalTerms || {}}
                />
            )}

            <div className="preset-specs">
                <div className="spec-row">
                    <div className="spec-item">
                        <span className="spec-icon">💾</span>
                        <div className="spec-details">
                            <span className="spec-value">{preset.resources.cpu} vCPU</span>
                            <span className="spec-label">프로세서</span>
                        </div>
                    </div>
                    <div className="spec-item">
                        <span className="spec-icon">🧠</span>
                        <div className="spec-details">
                            <span className="spec-value">{(preset.resources.memory / 1024).toFixed(1)} GB</span>
                            <span className="spec-label">메모리</span>
                        </div>
                    </div>
                    <div className="spec-item">
                        <span className="spec-icon">💿</span>
                        <div className="spec-details">
                            <span className="spec-value">{preset.resources.disk} GB</span>
                            <span className="spec-label">스토리지</span>
                        </div>
                    </div>
                </div>
            </div>

            <div className="preset-usecases">
                <div className="usecases-header">
                    <span className="usecases-icon">✨</span>
                    <span className="usecases-title">추천 사용 사례</span>
                </div>
                <div className="usecases-list">
                    {preset.useCases.map((useCase, index) => (
                        <span key={index} className="usecase-tag">{useCase}</span>
                    ))}
                </div>
            </div>

            <div className="preset-cost">
                <div className="cost-row">
                    <div className="cost-item-main">
                        <span className="cost-icon">⚡</span>
                        <div className="cost-details">
                            <span className="cost-label">시작 비용</span>
                            <span className="cost-value creation">{preset.cost.creationCost} 🍋</span>
                        </div>
                    </div>
                    <div className="cost-item-main">
                        <span className="cost-icon">⏱️</span>
                        <div className="cost-details">
                            <span className="cost-label">시간당</span>
                            <span className="cost-value hourly">{preset.cost.hourlyLemons} 🍋</span>
                        </div>
                    </div>
                </div>
                <div className="cost-monthly">
                    <span className="monthly-label">월 예상</span>
                    <span className="monthly-value">{preset.cost.monthlyLemons.toLocaleString()} 🍋</span>
                </div>
            </div>
        </div>
    );
};

export default PresetCard;