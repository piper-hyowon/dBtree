import React, {useEffect, useState} from "react";
import "./Home.css";
import dbtreeLogo from "../../assets/images/dbtree_logo.svg";
import accountIcon from "../../assets/images/character/account-icon.png";
import ToggleThemeButton from "../../components/common/ToggleThemeButton/ToggleThemeButton";
import LoginModal from "../../components/auth/LoginModal";
import GlobalStats from "../../components/GlobalStats/GlobalStats";
import LemonTreeScene from "../../components/LemonTreeScene/LemonTreeScene";
import FloatingGuideText from "../../components/common/FloatingGuideText/FloatingGuideText";
import {useAuth} from "../../contexts/AuthContext";
import MiniLeaderboard from "../../components/MiniLeaderboard/MiniLeaderboard";
import {User} from "../../services/api/auth.api";
import api from "../../services/api";

const Home: React.FC = () => {
    const [showLoginModal, setShowLoginModal] = useState(false);
    const {isLoggedIn, logout} = useAuth();
    const [loading, setLoading] = useState(true);
    const [user, setUser] = useState<User | null>(null);

    const loadData = async () => {
        if (isLoggedIn) {
            try {
                setLoading(true);
                const userResponse = await api.user.getUserProfile()

                setUser(userResponse);
            } catch (error) {
                console.error('Failed to load account data:', error);
            } finally {
                setLoading(false);
            }
        }
    };

    useEffect(() => {
        loadData();
    }, []);

    const handleStartNow = () => {
        setShowLoginModal(true);
    };

    const handleCloseModal = () => {
        setShowLoginModal(false);
    };

    const handleNavigateToDashboard = () => {
        window.location.href = "/dashboard";
    };

    const handleNavigateToMyPage = () => {
        window.location.href = "/account";
    };

    const handleLogout = async () => {
        await logout();
        window.location.href = "/";
    };

    return (
        <div className="home-container">
            <header className="header">
                <div className="logo-container">
                    <img src={dbtreeLogo} alt="dBtree Logo" className="logo"/>
                </div>
                <nav className="nav">
                    {isLoggedIn ? (
                        <>
                            <div className="user-info">
                                <button
                                    className="user-email"
                                    onClick={handleNavigateToMyPage}
                                    title="내 프로필로 이동"
                                >
                                    <img src={accountIcon} alt="account icon"/>
                                    <span className="user-email-text">{user?.email}</span>
                                </button>
                                <div className="lemon-balance" title="보유 레몬">
                                    <span className="lemon-emoji">🍋</span>
                                    <span>{user?.lemonBalance || 0}</span>
                                </div>
                            </div>

                            <div className="nav-actions">
                                <button
                                    className="nav-button dashboard-button"
                                    onClick={handleNavigateToDashboard}
                                    title="대시보드로 이동"
                                >
                                    대시보드
                                </button>
                                <button
                                    className="nav-button logout-button"
                                    onClick={handleLogout}
                                    title="로그아웃"
                                >
                                    로그아웃
                                </button>
                            </div>
                        </>
                    ) : (
                        <button className="nav-button login-button" onClick={handleStartNow}>
                            로그인
                        </button>
                    )}
                    <ToggleThemeButton/>
                </nav>
            </header>

            {/* 히어로 섹션 */}
            <section className="hero-section">
                <div className="hero-content">
                    <h1 className="hero-title">
                        레몬 나무에서
                        <br/>
                        <span className="highlight">무료 데이터베이스</span>를 수확하세요
                    </h1>
                    <p className="hero-subtitle">
                        쉽고 효율적인 크레딧 기반 DBaaS, 레몬을 먼저 수확한 사람이 임자!
                    </p>

                    {/* GlobalStats 컴포넌트 - 내부에서 API 호출 */}
                    <div className="global-stats">
                        <GlobalStats/>
                    </div>

                    {!isLoggedIn && (
                        <button className="cta-button" onClick={handleStartNow}>
                            무료로 시작하기
                        </button>
                    )}

                    <p className="limited-offer">
                        매일 새로운 레몬이 자라납니다. 선착순 수확!
                    </p>
                    <p className="golden-lemon-alert">
                        가끔 등장하는 황금 레몬을 놓치지 마세요!
                    </p>
                    <MiniLeaderboard/>
                </div>

                <div className="lemon-tree-container">
                    <LemonTreeScene/>
                    <FloatingGuideText
                        text="레몬을 클릭해서 수확해보세요"
                        emoji="🍋"
                        position="right"
                        variant="default"
                        dismissible={true}
                    />
                </div>
            </section>

            <footer className="footer">
                <p>© 2025 dBtree</p>
            </footer>

            {showLoginModal && <LoginModal onClose={handleCloseModal}/>}
        </div>
    );
};

export default Home;