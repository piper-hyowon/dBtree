import React, { useState, useRef, useEffect } from 'react';
import './Dashboard.css';
import { mockDatabases, mockDatabaseDetails, lemonCredits } from '../../data/mockData';
import { Database, DatabaseDetail } from '../../types/dashboard.types';
import ToggleThemeButton from '../../components/common/ToggleThemeButton/ToggleThemeButton';
import DeleteModal from '../../components/common/DeleteModal/DeleteModal';
import { getCharacterByStatus, characterImages } from '../../utils/characterImages';
import { useToast } from '../../hooks/useToast';
import { useTheme } from '../../hooks/useTheme';

const Dashboard: React.FC = () => {
    const [databases, setDatabases] = useState<Database[]>(mockDatabases);
    const [selectedDb, setSelectedDb] = useState<Database | null>(databases[0]);
    const [selectedDbDetail, setSelectedDbDetail] = useState<DatabaseDetail | null>(null);
    const [sidebarWidth, setSidebarWidth] = useState(320);
    const [isResizing, setIsResizing] = useState(false);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set());
    const sidebarRef = useRef<HTMLDivElement>(null);
    const { showToast } = useToast();
    const { theme } = useTheme();

    // Load database details when selection changes
    useEffect(() => {
        if (selectedDb) {
            const detail = mockDatabaseDetails[selectedDb.id];
            setSelectedDbDetail(detail || null);
        }
    }, [selectedDb]);

    const getStatusDot = (status: string) => {
        switch (status) {
            case 'running':
                return 'status-dot-running';
            case 'stopped':
                return 'status-dot-stopped';
            case 'provisioning':
                return 'status-dot-provisioning';
            case 'error':
                return 'status-dot-error';
            case 'maintenance':
                return 'status-dot-maintenance';
            default:
                return '';
        }
    };

    const getDatabaseIcon = () => {
        return theme === 'dark'
            ? '/images/mongodb_logo_dark.svg'
            : '/images/mongodb_logo_light.svg';
    };

    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('ko-KR', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        showToast('클립보드에 복사되었습니다', 'success', 2000);
    };

    const toggleSection = (section: string) => {
        const newCollapsed = new Set(collapsedSections);
        if (newCollapsed.has(section)) {
            newCollapsed.delete(section);
        } else {
            newCollapsed.add(section);
        }
        setCollapsedSections(newCollapsed);
    };

    // Resize handler
    const handleMouseDown = () => {
        setIsResizing(true);
    };

    useEffect(() => {
        const handleMouseMove = (e: MouseEvent) => {
            if (!isResizing) return;

            const newWidth = e.clientX;
            if (newWidth >= 280 && newWidth <= 480) {
                setSidebarWidth(newWidth);
            }
        };

        const handleMouseUp = () => {
            setIsResizing(false);
        };

        if (isResizing) {
            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);
        }

        return () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
        };
    }, [isResizing]);

    const handleConfiguration = () => {
        showToast('인스턴스 구성 변경 기능을 준비 중입니다', 'info');
    };

    const handlePause = () => {
        showToast('인스턴스 중지 기능을 준비 중입니다', 'info');
    };

    const handleRestart = () => {
        showToast('인스턴스 재시작 기능을 준비 중입니다', 'info');
    };

    const handleDelete = () => {
        setShowDeleteModal(true);
    };

    const confirmDelete = () => {
        if (selectedDb) {
            setDatabases(prev => prev.filter(db => db.id !== selectedDb.id));
            showToast(`${selectedDb.name}이(가) 삭제되었습니다`, 'success');
            setSelectedDb(databases.find(db => db.id !== selectedDb.id) || null);
            setShowDeleteModal(false);
        }
    };

    return (
        <div className="dashboard-container">
            {/* Header */}
            <header className="dashboard-header">
                <div className="header-content">
                    <div className="header-left">
                        <div className="logo-text" onClick={() => window.location.href = '/'}>
                            d<span className="logo-b">B</span>tree
                        </div>
                        <div className="header-separator">|</div>
                        <h1 className="dashboard-title">대시보드</h1>
                    </div>

                    <div className="header-right">
                        <div className="credit-section">
                            <span className="credit-label">레몬 크레딧</span>
                            <div className="credit-display">
                                <img
                                    src={lemonCredits > 50 ? characterImages.richInCredits : characterImages.lowCredits}
                                    alt="Credits"
                                    className="credit-icon"
                                />
                                <span className="credit-amount">{lemonCredits}</span>
                            </div>
                        </div>
                        <ToggleThemeButton />
                    </div>
                </div>
            </header>

            <div className="dashboard-body">
                {/* Sidebar */}
                <aside
                    className="sidebar"
                    ref={sidebarRef}
                    style={{ width: `${sidebarWidth}px` }}
                >
                    <div className="sidebar-content">
                        <h2 className="sidebar-title">데이터베이스 인스턴스</h2>

                        <button className="create-db-btn">
                            <span className="btn-icon">+</span>
                            <span>새 인스턴스 생성</span>
                        </button>

                        <div className="db-list">
                            {databases.map((db) => (
                                <div
                                    key={db.id}
                                    className={`db-item ${selectedDb?.id === db.id ? 'active' : ''}`}
                                    onClick={() => setSelectedDb(db)}
                                >
                                    <div className="db-item-header">
                                        <span className={`status-dot ${getStatusDot(db.status)}`}></span>
                                        <img
                                            src={getDatabaseIcon()}
                                            alt="MongoDB"
                                            className="db-type-icon"
                                        />
                                        <span className="db-name">{db.name}</span>
                                    </div>
                                    <div className="db-item-info">
                                        <span className="db-type">{db.type} • {db.size}</span>
                                        <span className="db-cost">{db.cost.hourlyLemons} 🍋/시간</span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <div
                        className="resize-handle"
                        onMouseDown={handleMouseDown}
                    />
                </aside>

                {/* Main Content */}
                <main className="main-content">
                    {selectedDb && selectedDbDetail ? (
                        <div className="content-wrapper">
                            {/* Database Header with Status Overview */}
                            <div className="db-detail-header">
                                <div className="character-section">
                                    <img
                                        src={getCharacterByStatus(selectedDb.status)}
                                        alt={selectedDb.status}
                                        className="db-main-character"
                                    />
                                    <div className="character-bubble">
                                        {selectedDb.status === 'running' ? '정상 작동 중!' :
                                            selectedDb.status === 'stopped' ? '휴식 중...' :
                                                selectedDb.status === 'provisioning' ? '준비 중!' :
                                                    '문제 발생!'}
                                    </div>
                                </div>

                                <div className="db-info-section">
                                    <div className="db-header-top">
                                        <div className="db-header-info">
                                            <h2 className="db-detail-title">{selectedDb.name}</h2>
                                            <div className="db-badges">
                                                <span className="badge badge-type">{selectedDb.type}</span>
                                                <span className="badge badge-size">{selectedDb.size}</span>
                                                <span className={`badge badge-status status-${selectedDb.status}`}>
                          {selectedDb.status === 'running' ? '실행 중' :
                              selectedDb.status === 'stopped' ? '중지됨' :
                                  selectedDb.status === 'provisioning' ? '프로비저닝 중' :
                                      selectedDb.status}
                        </span>
                                            </div>
                                        </div>

                                        <div className="db-actions">
                                            <button className="action-btn" onClick={handleConfiguration}>구성 변경</button>
                                            <button className="action-btn" onClick={handlePause}>중지</button>
                                            <button className="action-btn" onClick={handleRestart}>재시작</button>
                                            <button className="action-btn danger" onClick={handleDelete}>삭제</button>
                                        </div>
                                    </div>

                                    {/* System Status Overview */}
                                    <div className="ㅈ">
                                        <div className="status-item">
                                            <span className="status-label">CPU</span>
                                            <span className="status-value">준비 중</span>
                                        </div>
                                        <div className="status-item">
                                            <span className="status-label">메모리</span>
                                            <span className="status-value">준비 중</span>
                                        </div>
                                        <div className="status-item">
                                            <span className="status-label">디스크</span>
                                            <span className="status-value">준비 중</span>
                                        </div>
                                        <div className="status-item">
                                            <span className="status-label">연결</span>
                                            <span className="status-value">준비 중</span>
                                        </div>
                                        <button className="metrics-detail-btn" onClick={() => showToast('상세 메트릭 페이지를 준비 중입니다', 'info')}>
                                            상세 메트릭 →
                                        </button>
                                    </div>
                                </div>
                            </div>

                            {/* Cost Details Section with Character */}
                            <div className="section-card cost-section">
                                <div className="cost-character-section">
                                    <img
                                        src={lemonCredits > selectedDb.cost.monthlyLemons ? characterImages.richInCredits : characterImages.lowCredits}
                                        alt="Credits Status"
                                        className="cost-character"
                                    />
                                    <div className="character-message">
                                        {lemonCredits > selectedDb.cost.monthlyLemons
                                            ? '크레딧이 충분해요!'
                                            : '크레딧이 부족해요!'}
                                    </div>
                                </div>

                                <div className="cost-details">
                                    <h3 className="section-title">비용 상세</h3>
                                    <div className="cost-grid">
                                        <div className="cost-item primary">
                                            <span className="cost-label">시간당</span>
                                            <span className="cost-value">{selectedDb.cost.hourlyLemons} 🍋</span>
                                        </div>
                                        <div className="cost-item">
                                            <span className="cost-label">일일 (24시간)</span>
                                            <span className="cost-value">{selectedDb.cost.dailyLemons} 🍋</span>
                                        </div>
                                        <div className="cost-item">
                                            <span className="cost-label">월간 (30일)</span>
                                            <span className="cost-value">{selectedDb.cost.monthlyLemons.toLocaleString()} 🍋</span>
                                        </div>
                                        <div className="cost-item secondary">
                                            <span className="cost-label">프로비저닝</span>
                                            <span className="cost-value">{selectedDb.cost.creationCost} 🍋</span>
                                            <span className="cost-note">최초 1회</span>
                                        </div>
                                    </div>
                                    <div className="cost-summary">
                                        💡 현재 잔액으로 약 {Math.floor(lemonCredits / selectedDb.cost.hourlyLemons)}시간 사용 가능
                                    </div>
                                </div>
                            </div>

                            {/* Connection Info Section */}
                            <div className="section-card">
                                <div
                                    className="section-header"
                                    onClick={() => toggleSection('connection')}
                                >
                                    <h3 className="section-title">연결 정보</h3>
                                    <button className="collapse-btn">
                                        {collapsedSections.has('connection') ? '▶' : '▼'}
                                    </button>
                                </div>
                                {!collapsedSections.has('connection') && (
                                    <div className="section-content">
                                        <div className="connection-row">
                                            <span className="connection-label">엔드포인트</span>
                                            <div className="connection-value">
                                                <code>{selectedDbDetail.externalHost}:{selectedDbDetail.externalPort}</code>
                                                <button
                                                    className="copy-btn"
                                                    onClick={() => copyToClipboard(`${selectedDbDetail.externalHost}:${selectedDbDetail.externalPort}`)}
                                                >
                                                    📋
                                                </button>
                                            </div>
                                        </div>
                                        <div className="connection-row">
                                            <span className="connection-label">연결 문자열</span>
                                            <div className="connection-value">
                                                <code className="connection-uri">{selectedDbDetail.externalUriTemplate}</code>
                                                <button
                                                    className="copy-btn"
                                                    onClick={() => copyToClipboard(selectedDbDetail.externalUriTemplate)}
                                                >
                                                    📋
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>

                            {/* Detailed Info Section */}
                            <div className="section-card">
                                <div
                                    className="section-header"
                                    onClick={() => toggleSection('details')}
                                >
                                    <h3 className="section-title">상세 정보</h3>
                                    <button className="collapse-btn">
                                        {collapsedSections.has('details') ? '▶' : '▼'}
                                    </button>
                                </div>
                                {!collapsedSections.has('details') && (
                                    <div className="section-content">
                                        <div className="detail-grid">
                                            <div className="detail-group">
                                                <h4 className="detail-group-title">리소스</h4>
                                                <div className="detail-item">
                                                    <span className="detail-label">CPU</span>
                                                    <span className="detail-value">{selectedDb.resources.cpu} vCPU</span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">메모리</span>
                                                    <span className="detail-value">{(selectedDb.resources.memory / 1024).toFixed(1)} GB</span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">스토리지</span>
                                                    <span className="detail-value">{selectedDb.resources.disk} GB</span>
                                                </div>
                                            </div>

                                            <div className="detail-group">
                                                <h4 className="detail-group-title">구성</h4>
                                                <div className="detail-item">
                                                    <span className="detail-label">버전</span>
                                                    <span className="detail-value">{selectedDb.config.version}</span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">모드</span>
                                                    <span className="detail-value">{selectedDb.mode}</span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">백업</span>
                                                    <span className="detail-value">
                            {selectedDb.backupEnabled ? '활성화' : '비활성화'}
                          </span>
                                                </div>
                                            </div>

                                            <div className="detail-group">
                                                <h4 className="detail-group-title">기타</h4>
                                                <div className="detail-item">
                                                    <span className="detail-label">생성일</span>
                                                    <span className="detail-value">{formatDate(selectedDb.createdAt)}</span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">프리셋</span>
                                                    <span className="detail-value preset">{selectedDb.createdFromPreset}</span>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    ) : (
                        <div className="empty-state">
                            <img
                                src={characterImages.default}
                                alt="Select database"
                                className="empty-character"
                            />
                            <h3>데이터베이스를 선택해주세요</h3>
                            <p>왼쪽 목록에서 데이터베이스를 선택하면 상세 정보를 확인할 수 있습니다.</p>
                        </div>
                    )}
                </main>
            </div>

            {/* Delete Modal */}
            <DeleteModal
                isOpen={showDeleteModal}
                onClose={() => setShowDeleteModal(false)}
                onConfirm={confirmDelete}
                itemName={selectedDb?.name || ''}
            />
        </div>
    );
}

export default Dashboard;