import React, {useState, useEffect} from 'react';
import './AccountPage.css';
import api from "../../services/api";
import {User} from "../../services/api/auth.api";
import {DailyHarvest, TransactionWithInstance, UserInstanceSummary} from "../../services/api/account.api";
import ToggleThemeButton from '../../components/common/ToggleThemeButton/ToggleThemeButton';
import AccountDeleteModal from '../../components/common/AccountDeleteModal/AccountDeleteModal';
import {useNavigate} from 'react-router-dom';
import accountIcon from "../../assets/images/character/account-icon.png";
import welcomeBadge from "../../assets/images/badges/badge_welcome.png";

interface AccountStats {
    totalHarvested: number;
    quizAttempts: number;
    correctRate: number;
}

type TransactionTab = 'all' | 'by-instance';

const AccountPage: React.FC = () => {
    const [user, setUser] = useState<User | null>(null);
    const [dailyHarvest, setDailyHarvest] = useState<DailyHarvest[]>([]);
    const [transactions, setTransactions] = useState<TransactionWithInstance[]>([]);
    const [instances, setInstances] = useState<UserInstanceSummary[]>([]);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState<TransactionTab>('all');
    const [selectedInstance, setSelectedInstance] = useState<string>('');
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set());
    const [showDeleteModal, setShowDeleteModal] = useState(false);

    const navigate = useNavigate();

    // TODO:
    const accountStats: AccountStats = {
        totalHarvested: 100,
        quizAttempts: 200,
        correctRate: 50
    };

    useEffect(() => {
        loadData();
    }, []);

    useEffect(() => {
        loadTransactions();
    }, [activeTab, selectedInstance, currentPage]);

    const loadData = async () => {
        try {
            setLoading(true);
            const [userResponse, harvestResponse, instancesResponse] = await Promise.all([
                api.user.getUserProfile(),
                api.account.getDailyHarvest({days: 7}),
                api.account.getInstanceNames()
            ]);

            setUser(userResponse);
            setDailyHarvest(harvestResponse || []);
            setInstances(instancesResponse || []);
        } catch (error) {
            console.error('Failed to load account data:', error);
            setDailyHarvest([]);
            setInstances([]);
        } finally {
            setLoading(false);
        }
    };

    const loadTransactions = async () => {
        try {
            const params: any = {
                page: currentPage,
                limit: 10
            };

            if (activeTab === 'by-instance' && selectedInstance && selectedInstance !== 'all') {
                params.instance_name = selectedInstance;
            }

            const response = await api.account.getTransactions(params);
            setTransactions(response.data);
            setTotalPages(response.pagination.totalPages);
        } catch (error) {
            console.error('Failed to load transactions:', error);
        }
    };

    const handleTabChange = (tab: TransactionTab) => {
        setActiveTab(tab);
        setCurrentPage(1);
        if (tab === 'all') {
            setSelectedInstance('');
        }
    };

    const handleInstanceChange = (instanceName: string) => {
        setSelectedInstance(instanceName);
        setCurrentPage(1);
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

    const formatAmount = (amount: number) => {
        return amount > 0 ? `+${amount}` : amount.toString();
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('ko-KR', {
            month: 'short',
            day: 'numeric'
        });
    };

    // 7일간의 전체 데이터를 생성 (없는 날은 null로, 0과 구분)
    const generateFullWeekData = (harvestData: DailyHarvest[]) => {
        const fullWeekData: (DailyHarvest | { date: string; amount: null })[] = [];
        const today = new Date();

        for (let i = 6; i >= 0; i--) {
            const date = new Date(today);
            date.setDate(today.getDate() - i);

            // YYYY-MM-DD 형식으로 비교
            const dateString = date.toISOString().split('T')[0];

            // 해당 날짜의 데이터 찾기 (날짜 부분만 비교)
            const existingData = harvestData.find(item => {
                const itemDate = new Date(item.date).toISOString().split('T')[0];
                return itemDate === dateString;
            });

            if (existingData) {
                fullWeekData.push(existingData);
            } else {
                // 데이터가 없는 날 (가입 전 또는 API에서 제외된 날)
                fullWeekData.push({
                    date: date.toISOString(),
                    amount: null
                });
            }
        }

        return fullWeekData;
    };

    const getActionTypeLabel = (actionType: string) => {
        const labels: Record<string, string> = {
            welcome_bonus: '가입 보너스',
            harvest: '레몬 수확',
            instance_create: '인스턴스 생성',
            instance_maintain: '유지 비용',
            instance_create_refund: '환불'
        };
        return labels[actionType] || actionType;
    };

    const generateChartPath = (data: (DailyHarvest | { date: string; amount: null })[]): {
        linePath: string;
        areaPath: string
    } => {
        const actualData = data.filter(item => item.amount !== null) as DailyHarvest[];
        if (actualData?.length === 0) return {linePath: '', areaPath: ''};

        const maxAmount = Math.max(...actualData.map(item => item.amount), 1);
        const width = 100; // 백분율
        const height = 100; // 백분율
        const stepX = width / (data?.length - 1);

        let path = '';
        let areaPath = '';

        data.forEach((item, index) => {
            const x = index * stepX;
            const y = item.amount !== null ? height - (item.amount / maxAmount) * height : height;

            if (index === 0) {
                path += `M ${x} ${y}`;
                areaPath += `M ${x} ${height} L ${x} ${y}`;
            } else {
                path += ` L ${x} ${y}`;
                areaPath += ` L ${x} ${y}`;
            }
        });

        areaPath += ` L ${width} ${height} Z`;

        return {linePath: path, areaPath};
    };

    if (loading) {
        return (
            <div className="account-page">
                <div className="loading-state">✨ 데이터를 불러오는 중...</div>
            </div>
        );
    }

    const fullWeekData = generateFullWeekData(dailyHarvest);
    const chartPaths = generateChartPath(fullWeekData);

    return (
        <div className="account-page">
            {/* Header */}
            <header className="account-header">
                <div className="header-content">
                    <div className="header-left">
                        <div className="logo-text" onClick={() => navigate('/')}>
                            d<span className="logo-b">B</span>tree
                        </div>
                        <div className="header-separator">|</div>
                        <h1 className="page-title">계정 정보</h1>
                    </div>

                    <div className="user-info">
                        <button
                            className="user-email"
                            onClick={() => navigate('/dashboard')}
                            title="대시보드로 이동"
                        >
                            <img src={accountIcon} alt="account icon"/>
                            <span className="user-email-text">{user?.email}</span>
                        </button>
                        <div className="lemon-balance" title="보유 레몬">
                            <span className="lemon-emoji">🍋</span>
                            <span>{user?.lemonBalance?.toLocaleString() || 0}</span>
                        </div>
                        <ToggleThemeButton/>
                    </div>
                </div>
            </header>

            <div className="account-container">
                {/* Character Header Section */}
                <div className="character-header-section">
                    <div className="character-section">
                        <img
                            src={accountIcon}
                            alt="Account Character"
                            className="account-character"
                        />
                    </div>
                    <div className="user-info-section">
                        <div className="user-basic-info">
                            <h2 className="user-name">{user?.email}</h2>
                            {user?.lastHarvestAt ? `마지막 수확 ${user?.lastHarvestAt}` : ''}
                        </div>

                        <div className="user-stats-grid">
                            <div className="user-stat-item">
                                <div className="stat-value">🍋 {user?.lemonBalance?.toLocaleString() || 0}</div>
                                <div className="stat-label">현재 레몬</div>
                            </div>
                            <div className="user-stat-item">
                                <div className="stat-value">🍋 {user?.totalEarnedLemons?.toLocaleString() || 0}</div>
                                <div className="stat-label">총 획득</div>
                            </div>
                            <div className="user-stat-item">
                                <div className="stat-value">🍋 {user?.totalSpentLemons?.toLocaleString() || 0}</div>
                                <div className="stat-label">총 사용</div>
                            </div>
                            <div className="user-stat-item">
                                <div
                                    className="stat-value"> {new Date(user?.joinedAt || '').toLocaleDateString('ko-KR')}</div>
                                <div className="stat-label">가입일</div>
                            </div>
                        </div>

                        <div className="account-actions">
                            <button
                                className="delete-account-btn-simple"
                                onClick={() => setShowDeleteModal(true)}
                                title="계정을 삭제하면 모든 데이터가 영구 삭제됩니다"
                            >
                                계정 삭제
                            </button>
                        </div>
                    </div>
                </div>

                {/* Achievement Badges Section */}
                <div className="section-card achievements-section">
                    <div
                        className="section-header"
                        onClick={() => toggleSection('achievements')}
                    >
                        <h3 className="section-title">업적 뱃지</h3>
                        <button className="collapse-btn">
                            {collapsedSections.has('achievements') ? '▶' : '▼'}
                        </button>
                    </div>
                    {!collapsedSections.has('achievements') && (
                        <div className="achievements-grid">
                            {[{
                                id: 0,
                                earnedAt: user?.joinedAt,
                                iconUrl: welcomeBadge,
                                name: 'Welcome Farmer!',
                                description: ""
                            }].map((achievement) => (
                                <div
                                    key={achievement.id}
                                    className="achievement-badge"
                                    title={`획득일: ${achievement.earnedAt}`}
                                >
                                    <img
                                        src={achievement.iconUrl}
                                        alt={achievement.name}
                                        className="achievement-icon"
                                        onError={(e) => {
                                            console.log(`${achievement.name} 로드 실패`)
                                            e.currentTarget.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDJMMTMuMDkgOC4yNkwyMCA5TDEzLjA5IDE1Ljc0TDEyIDIyTDEwLjkxIDE1Ljc0TDQgOUwxMC45MSA4LjI2TDEyIDJaIiBmaWxsPSIjRkZEOTNEIi8+Cjwvc3ZnPgo=';
                                        }}
                                    />
                                    <div className="achievement-info">
                                        <div className="achievement-name">{achievement.name}</div>
                                        <div className="achievement-description">{achievement.description}</div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* 일일 레몬 수확량 - 꺾은선 그래프 */}
                <div className="section-card">
                    <div
                        className="section-header"
                        onClick={() => toggleSection('harvest')}
                    >
                        <h3 className="section-title">최근 7일 수확량</h3>
                        <button className="collapse-btn">
                            {collapsedSections.has('harvest') ? '▶' : '▼'}
                        </button>
                    </div>
                    {!collapsedSections.has('harvest') && (
                        <div className="section-content">
                            <div className="harvest-chart-container">
                                <div className="harvest-chart">
                                    {/* SVG 라인 차트 */}
                                    <div className="chart-line">
                                        <svg viewBox="0 0 100 100" preserveAspectRatio="none">
                                            <defs>
                                                <linearGradient id="lineGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                                                    <stop offset="0%"
                                                          style={{stopColor: 'var(--lemon-yellow)', stopOpacity: 1}}/>
                                                    <stop offset="100%"
                                                          style={{stopColor: 'var(--lemon-green)', stopOpacity: 1}}/>
                                                </linearGradient>
                                                <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                                                    <stop offset="0%"
                                                          style={{stopColor: 'var(--lemon-yellow)', stopOpacity: 0.3}}/>
                                                    <stop offset="100%"
                                                          style={{stopColor: 'var(--lemon-green)', stopOpacity: 0.05}}/>
                                                </linearGradient>
                                            </defs>
                                            {chartPaths.areaPath && (
                                                <path d={chartPaths.areaPath} className="chart-area"/>
                                            )}
                                            {chartPaths.linePath && (
                                                <path d={chartPaths.linePath}/>
                                            )}
                                        </svg>
                                    </div>

                                    {/* 데이터 포인트들 */}
                                    {(fullWeekData ?? []).map((item, index) => (
                                        <div key={`${item.date}-${index}`} className="harvest-point">
                                            <div className="chart-dot"></div>
                                            <div className={`harvest-amount ${item.amount === null ? 'no-data' : ''}`}>
                                                {item.amount === null ? '-' : item.amount === 0 ? '0' : item.amount.toString()}
                                            </div>
                                            <div className="harvest-date">{formatDate(item.date)}</div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* 계정 설정 섹션 - 탈퇴 버튼 포함 (버튼은 여기 하나만 있음!) */}
                <div className="section-card account-settings-section">
                    <div
                        className="section-header"
                        onClick={() => toggleSection('settings')}
                    >
                        <h3 className="section-title">계정 설정</h3>
                        <button className="collapse-btn">
                            {collapsedSections.has('settings') ? '▶' : '▼'}
                        </button>
                    </div>
                    {!collapsedSections.has('settings') && (
                        <div className="section-content compact">
                            <div className="danger-zone">
                                <div className="danger-zone-header">
                                    <h4 className="danger-zone-title">계정 삭제</h4>
                                    <p className="danger-zone-description">
                                        이 작업은 되돌릴 수 없으며, 모든 데이터가 영구 삭제됩니다.
                                    </p>
                                </div>
                                <button
                                    className="delete-account-btn"
                                    onClick={() => setShowDeleteModal(true)}
                                >
                                    계정 삭제
                                </button>
                            </div>
                        </div>
                    )}
                </div>

                {/* 레몬 입출금 내역 - 테이블 형태 */}
                <div className="section-card">
                    <div className="section-header">
                        <h3 className="section-title">레몬 입출금 내역</h3>
                        <div className="transaction-tabs">
                            <button
                                className={`tab-button ${activeTab === 'all' ? 'active' : ''}`}
                                onClick={() => handleTabChange('all')}
                            >
                                전체 내역
                            </button>
                            <button
                                className={`tab-button ${activeTab === 'by-instance' ? 'active' : ''}`}
                                onClick={() => handleTabChange('by-instance')}
                            >
                                인스턴스별
                            </button>
                        </div>
                    </div>

                    <div className="section-content">
                        {/* 인스턴스 드롭다운 */}
                        {activeTab === 'by-instance' && (
                            <div className="instance-filter">
                                <label htmlFor="instance-select" className="instance-label">
                                    인스턴스 선택:
                                </label>
                                <select
                                    id="instance-select"
                                    value={selectedInstance}
                                    onChange={(e) => handleInstanceChange(e.target.value)}
                                    className="instance-select"
                                >
                                    <option value="">전체 인스턴스</option>
                                    {(instances ?? []).map((instance) => (
                                        <option key={instance.id} value={instance.name}>
                                            {instance.name}
                                        </option>
                                    ))}
                                </select>
                                <div className="instance-count">
                                    {instances?.length > 0 ? `${instances?.length}개 인스턴스 보유` : '인스턴스 없음'}
                                </div>
                            </div>
                        )}

                        {/* 트랜잭션 테이블 */}
                        {transactions?.length === 0 ? (
                            <div className="no-transactions">
                                거래 내역이 없습니다
                            </div>
                        ) : (
                            <table className="transactions-table">
                                <thead>
                                <tr>
                                    <th>인스턴스</th>
                                    <th>거래 유형</th>
                                    <th>금액</th>
                                    <th>날짜</th>
                                </tr>
                                </thead>
                                <tbody>
                                {(transactions ?? []).map((transaction) => (
                                    <tr key={transaction.id}>
                                        <td className="transaction-instance-cell">
                                            {transaction.instanceName || '-'}
                                        </td>
                                        <td className="transaction-type-cell">
                                            {getActionTypeLabel(transaction.actionType)}
                                        </td>
                                        <td className="transaction-amount-cell">
                                                <span
                                                    className={`amount ${transaction.amount > 0 ? 'positive' : 'negative'}`}>
                                                    {formatAmount(transaction.amount)} 🍋
                                                </span>
                                        </td>
                                        <td className="transaction-date-cell">
                                            {new Date(transaction.createdAt).toLocaleString('ko-KR')}
                                        </td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        )}

                        {/* 페이지네이션 */}
                        {totalPages > 1 && (
                            <div className="pagination">
                                <button
                                    onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                                    disabled={currentPage === 1}
                                    className="pagination-button"
                                >
                                    ← 이전
                                </button>
                                <span className="pagination-info">
                                    {currentPage} / {totalPages}
                                </span>
                                <button
                                    onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                                    disabled={currentPage === totalPages}
                                    className="pagination-button"
                                >
                                    다음 →
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            <AccountDeleteModal
                isOpen={showDeleteModal}
                onClose={() => setShowDeleteModal(false)}
                userEmail={user?.email}
                lemonBalance={user?.lemonBalance}
            />
        </div>
    );
};

export default AccountPage;