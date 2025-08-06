import React from 'react';
import './BackupOptions.css';

interface BackupOptionsProps {
    enabled: boolean;
    schedule: string;
    retentionDays: number;
    onEnabledChange: (enabled: boolean) => void;
    onScheduleChange: (schedule: string) => void;
    onRetentionChange: (days: number) => void;
}

const BackupOptions: React.FC<BackupOptionsProps> = ({
                                                         enabled,
                                                         schedule,
                                                         retentionDays,
                                                         onEnabledChange,
                                                         onScheduleChange,
                                                         onRetentionChange
                                                     }) => {
    return (
        <div className="wizard-section backup-section-minimal">
            <div className="backup-coming-soon">
                <div className="coming-soon-badge">
                    <span className="badge-icon">🚧</span>
                    <span className="badge-text">Coming Soon</span>
                </div>
                <div className="coming-soon-content">
                    <h4>자동 백업 기능 준비 중</h4>
                    <p>일일 자동 백업, 원클릭 복원, 백업 스케줄링 기능이 곧 추가됩니다</p>
                </div>
            </div>
        </div>
    );
};

export default BackupOptions;