import React, {useEffect, useState} from 'react';
import SystemResourceHappy from '../../assets/images/character/system-resource-happy.png';
import SystemResourceBusy from '../../assets/images/character/system-resource-busy.png';
import SystemResourceStressed from '../../assets/images/character/system-resource-stressed.png';
import api from "../../services/api";
import "./SystemResourceStatus.css";
import {SystemResourceStatusResponse} from "../../services/api/home.api";

const SystemResourceStatus: React.FC = () => {
    const [status, setStatus] = useState<SystemResourceStatusResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [isMinimized, setIsMinimized] = useState(false);
    const [isVisible, setIsVisible] = useState(true);
    const [position, setPosition] = useState({x: 300, y: 80}); // 초기 위치 유지
    const [isDragging, setIsDragging] = useState(false);
    const [dragStart, setDragStart] = useState({x: 0, y: 0});

    useEffect(() => {
        const fetchResourceStatus = async () => {
            try {
                const response = await api.home.getSystemResources();
                setStatus(response);
            } catch (error) {
                console.error('Failed to fetch system resources:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchResourceStatus();
        // 30초마다 갱신
        const interval = setInterval(fetchResourceStatus, 30000);
        return () => clearInterval(interval);
    }, []);

    // 드래그 이벤트 핸들러
    const handleMouseDown = (e: React.MouseEvent) => {
        // 닫기 버튼 클릭시는 드래그 시작하지 않음
        if ((e.target as HTMLElement).closest('.close-button')) return;

        e.preventDefault(); // 텍스트 선택 방지
        setIsDragging(true);
        setDragStart({
            x: e.clientX - position.x,
            y: e.clientY - position.y
        });
    };

    useEffect(() => {
        const handleMouseMove = (e: MouseEvent) => {
            if (!isDragging) return;

            const newX = e.clientX - dragStart.x;
            const newY = e.clientY - dragStart.y;

            // 화면 밖으로 나가지 않도록 제한
            const maxX = window.innerWidth - 280; // 컴포넌트 너비
            const maxY = window.innerHeight - 150; // 대략적인 컴포넌트 높이

            setPosition({
                x: Math.max(0, Math.min(newX, maxX)),
                y: Math.max(0, Math.min(newY, maxY))
            });
        };

        const handleMouseUp = () => {
            setIsDragging(false);
        };

        if (isDragging) {
            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);
        }

        return () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
        };
    }, [isDragging, dragStart]);

    if (loading || !status || !isVisible) return null;

    // CPU와 Memory 사용률 계산 - total 대비 used로 계산
    const cpuUsagePercent = Math.min(100, Math.round((status.system.used.cpu / status.system.total.cpu) * 100));
    const memoryUsagePercent = Math.min(100, Math.round((status.system.used.memory / status.system.total.memory) * 100));
    const overallUsage = Math.round((cpuUsagePercent + memoryUsagePercent) / 2);

    const getStatusClass = (percent: number) => {
        if (percent < 60) return 'good';
        if (percent < 85) return 'warning';
        return 'critical';
    };

    const getStatusMessage = () => {
        if (overallUsage < 30) return '시스템 상태 매우 양호';
        if (overallUsage < 60) return '정상 운영 중';
        if (overallUsage < 85) return '시스템 부하 증가';
        return '리소스 포화 상태';
    };

    const getMascotImage = () => {
        if (overallUsage < 60) return SystemResourceHappy;
        if (overallUsage < 85) return SystemResourceBusy;
        return SystemResourceStressed;
    };

    const getAvailabilityStatus = () => {
        if (status.canCreateMedium) return {text: 'Medium 생성 가능', class: 'available'};
        if (status.canCreateSmall) return {text: 'Small 생성 가능', class: 'limited'};
        if (status.canCreateTiny) return {text: 'Tiny만 생성 가능', class: 'limited'};
        return {text: '인스턴스 생성 기다려야함', class: 'full'};
    };

    const availability = getAvailabilityStatus();

    return (
        <>
            <div
                className={`system-resource-status ${isMinimized ? 'minimized' : ''} ${isDragging ? 'dragging' : ''}`}
                style={{
                    left: `${position.x}px`,
                    top: `${position.y}px`
                }}
                onMouseDown={handleMouseDown}
            >
                <div className="resource-compact-layout">
                    <div className="mascot-wrapper">
                        <img
                            src={getMascotImage()}
                            alt="System Status"
                            className="mascot-image"
                        />
                    </div>

                    <div className="resource-content">
                        <div className="resource-header">
                            <div className="resource-title-wrapper">
                                <h3 className="resource-title">
                                    시스템 리소스
                                </h3>
                                <p className={`status-message ${getStatusClass(overallUsage)}`}>
                                    {getStatusMessage()}
                                </p>
                            </div>
                            <button
                                className="close-button"
                                onClick={() => setIsVisible(false)}
                                aria-label="닫기"
                            >
                                <svg viewBox="0 0 16 16" fill="currentColor">
                                    <path
                                        d="M12.707 4.707a1 1 0 0 0-1.414-1.414L8 6.586 4.707 3.293a1 1 0 0 0-1.414 1.414L6.586 8l-3.293 3.293a1 1 0 1 0 1.414 1.414L8 9.414l3.293 3.293a1 1 0 0 0 1.414-1.414L9.414 8l3.293-3.293z"/>
                                </svg>
                            </button>
                        </div>

                        {!isMinimized && (
                            <>
                                <div className="resource-meters">
                                    <div className="meter-item">
                                        <div className="meter-label">
                                            <span>CPU</span>
                                            <span className={`meter-value ${getStatusClass(cpuUsagePercent)}`}>
                                                {cpuUsagePercent}%
                                            </span>
                                        </div>
                                        <div className="meter-bar">
                                            <div
                                                className={`meter-fill ${getStatusClass(cpuUsagePercent)}`}
                                                style={{width: `${cpuUsagePercent}%`}}
                                            />
                                        </div>
                                    </div>

                                    <div className="meter-item">
                                        <div className="meter-label">
                                            <span>Memory</span>
                                            <span className={`meter-value ${getStatusClass(memoryUsagePercent)}`}>
                                                {memoryUsagePercent}%
                                            </span>
                                        </div>
                                        <div className="meter-bar">
                                            <div
                                                className={`meter-fill ${getStatusClass(memoryUsagePercent)}`}
                                                style={{width: `${memoryUsagePercent}%`}}
                                            />
                                        </div>
                                    </div>
                                </div>

                                <div className="resource-footer">
                                    <div className="footer-stat">
                                        활성: <strong>{status.activeCount}</strong>개
                                    </div>
                                    <div className={`status-badge ${availability.class}`}>
                                        <span className="status-dot"/>
                                        <span>{availability.text}</span>
                                    </div>
                                </div>
                            </>
                        )}
                    </div>
                </div>
            </div>
        </>
    );
};

export default SystemResourceStatus;