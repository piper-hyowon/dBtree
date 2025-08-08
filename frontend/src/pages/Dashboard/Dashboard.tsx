import React, {useState, useRef, useEffect} from 'react';
import './Dashboard.css';
import {InstanceResponse, DBType, InstanceStatus} from '../../types/database.types';
import {getInstances, getInstance, deleteInstance} from '../../services/api/database.api';
import ToggleThemeButton from '../../components/common/ToggleThemeButton/ToggleThemeButton';
import DeleteModal from '../../components/common/DeleteModal/DeleteModal';
import {getCharacterByStatus, characterImages, isImageComponent} from '../../utils/characterImages';
import {useToast} from '../../hooks/useToast';
import {useTheme} from '../../hooks/useTheme';
import {useAuth} from '../../contexts/AuthContext';
import {useNavigate} from 'react-router-dom';
import CreateInstanceWizard from "../../components/CreateInstanceWizard/CreateInstanceWizard";
import accountIcon from "../../assets/images/character/account-icon.png";

type ViewType = 'empty' | 'detail' | 'create';

const Dashboard: React.FC = () => {
    const [instances, setInstances] = useState<InstanceResponse[]>([]);
    const [selectedInstance, setSelectedInstance] = useState<InstanceResponse | null>(null);
    const [currentView, setCurrentView] = useState<ViewType>('empty');
    const [sidebarWidth, setSidebarWidth] = useState(320);
    const [isResizing, setIsResizing] = useState(false);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set());
    const [isLoading, setIsLoading] = useState(true);

    const sidebarRef = useRef<HTMLDivElement>(null);
    const refreshIntervalRef = useRef<NodeJS.Timeout | null>(null);
    const {showToast} = useToast();
    const {theme} = useTheme();
    const {user, isLoggedIn, loading: authLoading} = useAuth();
    const navigate = useNavigate();

    // 로그인 체크 - loading이 끝난 후에만 체크
    useEffect(() => {
        if (!authLoading && !isLoggedIn) {
            navigate('/');
        }
    }, [authLoading, isLoggedIn, navigate]);

    // 인스턴스 목록 로드
    useEffect(() => {
        if (!authLoading && isLoggedIn) {
            loadInstances();
        }
    }, [authLoading, isLoggedIn]);

    // Provisioning 상태인 인스턴스만 주기적으로 업데이트
    useEffect(() => {
        // 기존 interval 정리
        if (refreshIntervalRef.current) {
            clearInterval(refreshIntervalRef.current);
            refreshIntervalRef.current = null;
        }

        // provisioning 상태인 선택된 인스턴스가 있을 때만
        if (selectedInstance?.status === 'provisioning') {
            refreshIntervalRef.current = setInterval(async () => {
                try {
                    const updated = await getInstance(selectedInstance.id);
                    setSelectedInstance(updated);

                    // 목록에서도 업데이트
                    setInstances(prev => prev.map(inst =>
                        inst.id === updated.id ? updated : inst
                    ));

                    // provisioning이 끝나면 interval 정리
                    if (updated.status !== 'provisioning') {
                        if (refreshIntervalRef.current) {
                            clearInterval(refreshIntervalRef.current);
                            refreshIntervalRef.current = null;
                        }

                        // 상태 변경 알림
                        if (updated.status === 'running') {
                            showToast(`${updated.name}이(가) 준비되었습니다!`, 'success');
                        } else if (updated.status === 'error') {
                            showToast(`${updated.name} 생성 중 오류가 발생했습니다`, 'error');
                        }
                    }
                } catch (error) {
                    console.error('Failed to refresh instance:', error);
                }
            }, 5000);
        }

        return () => {
            if (refreshIntervalRef.current) {
                clearInterval(refreshIntervalRef.current);
            }
        };
    }, [selectedInstance?.id, selectedInstance?.status]);

    const loadInstances = async () => {
        try {
            setIsLoading(true);
            const data = await getInstances();
            setInstances(data);
        } catch (error) {
            showToast('인스턴스 목록을 불러오는데 실패했습니다', 'error');
        } finally {
            setIsLoading(false);
        }
    };

    const handleInstanceSelect = (instance: InstanceResponse) => {
        setSelectedInstance(instance);
        setCurrentView('detail');
    };

    const handleCreateClick = () => {
        setSelectedInstance(null);
        setCurrentView('create');
    };

    const getStatusDot = (status: InstanceStatus) => {
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

    const getStatusText = (status: InstanceStatus): string => {
        switch (status) {
            case 'running':
                return '실행 중';
            case 'stopped':
                return '중지됨';
            case 'provisioning':
                return '프로비저닝 중';
            case 'error':
                return '오류';
            case 'maintenance':
                return '유지보수';
            default:
                return status;
        }
    };

    const getDatabaseIcon = (type: DBType) => {
        if (type === 'mongodb') {
            return theme === 'dark'
                ? '/images/mongodb_logo_dark.svg'
                : '/images/mongodb_logo_light.svg';
        }
        // TODO: 현재는 레디스 지원 안 함 (로고도 없음)
        return '/images/redis_logo.svg';
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

    const handleStop = () => {
        showToast('인스턴스 중지 기능을 준비 중입니다', 'info');
    };

    const handleRestart = () => {
        showToast('인스턴스 재시작 기능을 준비 중입니다', 'info');
    };

    const handleDelete = () => {
        setShowDeleteModal(true);
    };

    const confirmDelete = async () => {
        if (!selectedInstance) return;

        try {
            await deleteInstance(selectedInstance.id);
            showToast(`${selectedInstance.name}이(가) 삭제되었습니다`, 'success');

            // 목록에서 제거
            const newInstances = instances.filter(inst => inst.id !== selectedInstance.id);
            setInstances(newInstances);

            // 뷰 초기화
            setSelectedInstance(null);
            setCurrentView('empty');
            setShowDeleteModal(false);
        } catch (error) {
            showToast('인스턴스 삭제에 실패했습니다', 'error');
        }
    };

    const getModeDisplayName = (type: DBType, mode: string): string => {
        if (type === 'mongodb') {
            switch (mode) {
                case 'standalone':
                    return 'Standalone';
                case 'replica_set':
                    return 'Replica Set';
                case 'sharded':
                    return 'Sharded';
                default:
                    return mode;
            }
        } else if (type === 'redis') {
            switch (mode) {
                case 'basic':
                    return 'Basic';
                case 'sentinel':
                    return 'Sentinel';
                case 'cluster':
                    return 'Cluster';
                default:
                    return mode;
            }
        }
        return mode;
    };

    const getSizeDisplayName = (size: string): string => {
        switch (size) {
            case 'tiny':
                return 'Tiny';
            case 'small':
                return 'Small';
            case 'medium':
                return 'Medium';
            case 'large':
                return 'Large';
            case 'xlarge':
                return 'XLarge';
            default:
                return size;
        }
    };

    return (
        <div className="dashboard-container">
            {/* Header */}
            <header className="dashboard-header">
                <div className="header-content">
                    <div className="header-left">
                        <div className="logo-text" onClick={() => navigate('/')}>
                            d<span className="logo-b">B</span>tree
                        </div>
                        <div className="header-separator">|</div>
                        <h1 className="dashboard-title">대시보드</h1>
                    </div>

                    <div className="user-info">
                        <button
                            className="user-email"
                            onClick={() => window.location.href = "/profile"}
                            title="내 프로필로 이동"
                        >
                            <img src={accountIcon} alt="account icon"/>
                            <span className="user-email-text">{user?.email}</span>
                        </button>
                        <div className="lemon-balance" title="보유 레몬">
                            <span className="lemon-emoji">🍋</span>
                            <span>{user?.lemonBalance || 0}</span>
                        </div>
                        <ToggleThemeButton/>
                    </div>
                </div>
            </header>

            <div className="dashboard-body">
                {/* Sidebar */}
                <aside
                    className="sidebar"
                    ref={sidebarRef}
                    style={{width: `${sidebarWidth}px`}}
                >
                    <div className="sidebar-content">
                        <h2 className="sidebar-title">데이터베이스 인스턴스</h2>

                        <button className="create-db-btn" onClick={handleCreateClick}>
                            <span className="btn-icon">+</span>
                            <span>새 인스턴스 생성</span>
                        </button>

                        <div className="db-list">
                            {isLoading ? (
                                <div className="loading-state">로딩 중...</div>
                            ) : instances.length === 0 ? (
                                <div className="empty-list-message">
                                    <p>아직 인스턴스가 없습니다</p>
                                    <p className="empty-list-hint">위 버튼을 눌러 첫 인스턴스를 생성해보세요!</p>
                                </div>
                            ) : (
                                instances.map((inst) => (
                                    <div
                                        key={inst.id}
                                        className={`db-item ${selectedInstance?.id === inst.id ? 'active' : ''}`}
                                        onClick={() => handleInstanceSelect(inst)}
                                    >
                                        <div className="db-item-header">
                                            <span className={`status-dot ${getStatusDot(inst.status)}`}></span>
                                            <img
                                                src={getDatabaseIcon(inst.type)}
                                                alt={inst.type}
                                                className="db-type-icon"
                                            />
                                            <span className="db-name">{inst.name}</span>
                                        </div>
                                        <div className="db-item-info">
                                            <span className="db-type">
                                                {inst.type.toUpperCase()} • {getSizeDisplayName(inst.size)}
                                            </span>
                                            <span className="db-cost">{inst.cost.hourlyLemons} 🍋/시간</span>
                                        </div>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>

                    <div
                        className="resize-handle"
                        onMouseDown={handleMouseDown}
                    />
                </aside>

                {/* Main Content */}
                <main className="main-content">
                    {authLoading ? (
                        <div className="loading-state">인증 확인 중...</div>
                    ) : currentView === 'empty' ? (
                        <div className="empty-state">
                            <img
                                src={characterImages.default}
                                alt="Welcome"
                                className="empty-character"
                            />
                            <p className="empty-description">
                                왼쪽 목록에서 인스턴스를 선택하거나<br/>
                                새로운 인스턴스를 생성해보세요
                            </p>
                            <button className="empty-create-btn" onClick={handleCreateClick}>
                                <span>🍋</span>
                                첫 인스턴스 생성하기
                            </button>
                        </div>
                    ) : currentView === 'create' ? (
                        <CreateInstanceWizard
                            currentInstanceCount={instances.length}
                            onSuccess={() => {
                                loadInstances();
                                setCurrentView('empty');
                            }}
                            onCancel={() => setCurrentView('empty')}
                        />

                    ) : currentView === 'detail' && selectedInstance ? (
                        <div className="content-wrapper">
                            {/* Database Header with Status Overview */}
                            <div className="db-detail-header">
                                <div className="character-section">
                                    {
                                        isImageComponent(getCharacterByStatus(selectedInstance.status)) ? (
                                            <div className="character-svg-wrapper">
                                                {React.createElement(getCharacterByStatus(selectedInstance.status), {
                                                    message: "프로비저닝 중...",
                                                    subMessage: "DB 준비 중"
                                                })}
                                            </div>
                                        ) : (
                                            <img
                                                src={getCharacterByStatus(selectedInstance.status) as string}
                                                alt={selectedInstance.status}
                                                className="db-main-character"
                                            />
                                        )
                                    }

                                    <div className="character-bubble">
                                        {selectedInstance.status === 'running' ? '정상 작동 중!' :
                                            selectedInstance.status === 'stopped' ? '휴식 중...' :
                                                selectedInstance.status === 'provisioning' ? '준비 중!' :
                                                    '문제 발생!'}
                                    </div>
                                </div>

                                <div className="db-info-section">
                                    <div className="db-header-top">
                                        <div className="db-header-info">
                                            <h2 className="db-detail-title">{selectedInstance.name}</h2>
                                            <div className="db-badges">
                                                <span className="badge badge-type">
                                                    {selectedInstance.type.toUpperCase()}
                                                </span>
                                                <span className="badge badge-size">
                                                    {getSizeDisplayName(selectedInstance.size)}
                                                </span>
                                                <span className="badge badge-mode">
                                                    {getModeDisplayName(selectedInstance.type, selectedInstance.mode)}
                                                </span>
                                                <span
                                                    className={`badge badge-status status-${selectedInstance.status}`}>
                                                    {getStatusText(selectedInstance.status)}
                                                </span>
                                            </div>
                                        </div>

                                        <div className="db-actions">
                                            <button className="action-btn" onClick={handleConfiguration}>구성 변경</button>
                                            <button className="action-btn" onClick={handleStop}>중지</button>
                                            <button className="action-btn" onClick={handleRestart}>재시작</button>
                                            <button className="action-btn danger" onClick={handleDelete}>삭제</button>
                                        </div>
                                    </div>

                                    {/* System Status Overview */}
                                    <div className="system-status-overview">
                                        <div className="status-badge">
                                            <div className="status-badge-dot loading"></div>
                                            <span className="status-badge-label">CPU</span>
                                            <span className="status-badge-value loading">
                                                준비중<span className="loading-dots"></span>
                                            </span>
                                        </div>

                                        <div className="status-badge">
                                            <div className="status-badge-dot loading"></div>
                                            <span className="status-badge-label">메모리</span>
                                            <span className="status-badge-value loading">
                                                준비중<span className="loading-dots"></span>
                                            </span>
                                        </div>

                                        <div className="status-badge">
                                            <div className="status-badge-dot loading"></div>
                                            <span className="status-badge-label">디스크</span>
                                            <span className="status-badge-value loading">
                                                준비중<span className="loading-dots"></span>
                                            </span>
                                        </div>

                                        <div className="status-badge">
                                            <div className="status-badge-dot loading"></div>
                                            <span className="status-badge-label">네트워크</span>
                                            <span className="status-badge-value loading">
                                                준비중<span className="loading-dots"></span>
                                            </span>
                                        </div>

                                        <button
                                            className="metrics-detail-btn"
                                            onClick={() => showToast('상세 메트릭 페이지를 준비 중입니다', 'info')}
                                        >
                                            상세 메트릭 →
                                        </button>
                                    </div>
                                </div>
                            </div>

                            {/* Cost Details Section with Character */}
                            <div className="section-card cost-section">
                                <div className="cost-character-section">
                                    <img
                                        src={(user?.lemonBalance || 0) > selectedInstance.cost.monthlyLemons
                                            ? characterImages.richInCredits
                                            : characterImages.lowCredits}
                                        alt="Credits Status"
                                        className="cost-character"
                                    />
                                    <div className="character-message">
                                        {(user?.lemonBalance || 0) > selectedInstance.cost.monthlyLemons
                                            ? '크레딧이 충분해요!'
                                            : '크레딧이 부족해요!'}
                                    </div>
                                </div>

                                <div className="cost-details">
                                    <h3 className="section-title">비용 상세</h3>
                                    <div className="cost-grid">
                                        <div className="cost-item primary">
                                            <span className="cost-label">시간당</span>
                                            <span className="cost-value">{selectedInstance.cost.hourlyLemons} 🍋</span>
                                        </div>
                                        <div className="cost-item">
                                            <span className="cost-label">일일 (24시간)</span>
                                            <span className="cost-value">{selectedInstance.cost.dailyLemons} 🍋</span>
                                        </div>
                                        <div className="cost-item">
                                            <span className="cost-label">월간 (30일)</span>
                                            <span className="cost-value">
                                                {selectedInstance.cost.monthlyLemons.toLocaleString()} 🍋
                                            </span>
                                        </div>
                                        <div className="cost-item secondary">
                                            <span className="cost-label">프로비저닝</span>
                                            <span className="cost-value">{selectedInstance.cost.creationCost} 🍋</span>
                                            <span className="cost-note">최초 1회</span>
                                        </div>
                                    </div>
                                    <div className="cost-summary">
                                        💡 현재 잔액으로
                                        약 {Math.floor((user?.lemonBalance || 0) / selectedInstance.cost.hourlyLemons)}시간
                                        사용 가능
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
                                        {selectedInstance.externalHost && selectedInstance.externalPort ? (
                                            <>
                                                <div className="connection-row">
                                                    <span className="connection-label">엔드포인트</span>
                                                    <div className="connection-value">
                                                        <code>
                                                            {selectedInstance.externalHost}:{selectedInstance.externalPort}
                                                        </code>
                                                        <button
                                                            className="copy-btn"
                                                            onClick={() => copyToClipboard(
                                                                `${selectedInstance.externalHost}:${selectedInstance.externalPort}`
                                                            )}
                                                        >
                                                            📋
                                                        </button>
                                                    </div>
                                                </div>
                                                {selectedInstance.externalUriTemplate && (
                                                    <div className="connection-row">
                                                        <span className="connection-label">연결 문자열</span>
                                                        <div className="connection-value">
                                                            <code className="connection-uri">
                                                                {selectedInstance.externalUriTemplate}
                                                            </code>
                                                            <button
                                                                className="copy-btn"
                                                                onClick={() => copyToClipboard(selectedInstance.externalUriTemplate!)}
                                                            >
                                                                📋
                                                            </button>
                                                        </div>
                                                    </div>
                                                )}
                                            </>
                                        ) : (
                                            <div className="connection-pending">
                                                {selectedInstance.status === 'provisioning'
                                                    ? '프로비저닝이 완료되면 연결 정보가 표시됩니다.'
                                                    : '연결 정보를 불러오는 중...'}
                                            </div>
                                        )}
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
                                                    <span className="detail-value">
                                                        {selectedInstance.resources.cpu} vCPU
                                                    </span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">메모리</span>
                                                    <span className="detail-value">
                                                        {(selectedInstance.resources.memory / 1024).toFixed(1)} GB
                                                    </span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">스토리지</span>
                                                    <span className="detail-value">
                                                        {selectedInstance.resources.disk} GB
                                                    </span>
                                                </div>
                                            </div>

                                            <div className="detail-group">
                                                <h4 className="detail-group-title">구성</h4>
                                                <div className="detail-item">
                                                    <span className="detail-label">타입</span>
                                                    <span className="detail-value">
                                                        {selectedInstance.type.toUpperCase()}
                                                    </span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">모드</span>
                                                    <span className="detail-value">
                                                        {getModeDisplayName(selectedInstance.type, selectedInstance.mode)}
                                                    </span>
                                                </div>
                                                <div className="detail-item">
                                                    <span className="detail-label">백업</span>
                                                    <span className="detail-value">
                                                        {selectedInstance.backupEnabled ? '활성화' : '비활성화'}
                                                    </span>
                                                </div>
                                                {selectedInstance.config?.version && (
                                                    <div className="detail-item">
                                                        <span className="detail-label">버전</span>
                                                        <span className="detail-value">
                                                            {selectedInstance.config.version}
                                                        </span>
                                                    </div>
                                                )}
                                            </div>

                                            <div className="detail-group">
                                                <h4 className="detail-group-title">기타</h4>
                                                <div className="detail-item">
                                                    <span className="detail-label">생성일</span>
                                                    <span className="detail-value">
                                                        {formatDate(selectedInstance.createdAt)}
                                                    </span>
                                                </div>
                                                {selectedInstance.createdFromPreset && (
                                                    <div className="detail-item">
                                                        <span className="detail-label">프리셋</span>
                                                        <span className="detail-value preset">
                                                            {selectedInstance.createdFromPreset}
                                                        </span>
                                                    </div>
                                                )}
                                                {selectedInstance.pausedAt && (
                                                    <div className="detail-item">
                                                        <span className="detail-label">중지 시간</span>
                                                        <span className="detail-value">
                                                            {formatDate(selectedInstance.pausedAt)}
                                                        </span>
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    ) : null}
                </main>
            </div>

            {/* Delete Modal */}
            <DeleteModal
                isOpen={showDeleteModal}
                onClose={() => setShowDeleteModal(false)}
                onConfirm={confirmDelete}
                itemName={selectedInstance?.name || ''}
            />
        </div>
    );
};

export default Dashboard;