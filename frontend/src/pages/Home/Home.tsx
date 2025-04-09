import React, { useState } from "react";
import "./Home.css";
import dbtreeLogo from "../../assets/images/dbtree_logo.svg";
import ToggleThemeButton from "../../components/common/ToggleThemeButton/ToggleThemeButton";
import LoginModal from "../../components/auth/LoginModal";
import { useAuth } from "../../hooks/useAuth";
import GlobalStats from "../../components/GlobalStats/GlobalStats";
import LemonTreeScene from "../../components/LemonTreeScene/LemonTreeScene";
// import LemonTreeApp from "./NewLemonTree";
// import NewNewLemonTree from "../../components/NewNewLemonTree/NewNewLemonTree";

const Home: React.FC = () => {
  const [showLoginModal, setShowLoginModal] = useState(false);
  const { isLoggedIn } = useAuth();

  const handleStartNow = () => {
    setShowLoginModal(true);
  };

  const handleCloseModal = () => {
    setShowLoginModal(false);
  };

  const handleNavigateToDashboard = () => {
    window.location.href = "/dashboard";
  };

  return (
    <div className="home-container">
      <header className="header">
        <div className="logo-container">
          <img src={dbtreeLogo} alt="dBtree Logo" className="logo" />
        </div>
        <nav className="nav">
          <ToggleThemeButton />

          {isLoggedIn ? (
            <button
              className="login-button"
              onClick={handleNavigateToDashboard}
            >
              대시보드
            </button>
          ) : (
            <button className="login-button" onClick={handleStartNow}>
              로그인
            </button>
          )}
        </nav>
      </header>

      {/* 히어로 섹션 */}
      <section className="hero-section">
        <div className="hero-content">
          <h1 className="hero-title">
            레몬 나무에서
            <br />
            <span className="highlight">무료 데이터베이스</span>를 수확하세요
          </h1>
          <p className="hero-subtitle">
            쉽고 효율적인 크레딧 기반 DBaaS, 레몬을 먼저 수확한 사람이 임자!
          </p>
          <div className="global-stats">
            <GlobalStats />
          </div>
          <button className="cta-button" onClick={handleStartNow}>
            무료로 시작하기
          </button>
          <p className="limited-offer">
            매일 새로운 레몬이 자라납니다. 선착순 수확!
          </p>
        </div>

        <div className="lemon-tree-container">
          <LemonTreeScene />
        </div>
      </section>

      <section className="features-section">
        <div className="features-grid">
          <div className="feature-card">
            <div className="feature-icon">🍋</div>
            <h3>공유 레몬 나무</h3>
            <p>
              모든 사용자를 위한 레몬 나무에서 레몬을 수확하세요. 빠른 사람이
              임자!
            </p>
          </div>

          <div className="feature-card">
            <div className="feature-icon">⏰</div>
            <h3>일일 레몬</h3>
            <p>정해진 시간에 레몬이 다시 자랍니다.</p>
          </div>

          <div className="feature-card">
            <div className="feature-icon">🔄</div>
            <h3>레몬으로 DB 생성</h3>
            <p>
              수확한 레몬으로 데이터베이스 인스턴스를 무료로 생성하고
              사용해보세요.
            </p>
          </div>
        </div>
      </section>

      <footer className="footer">
        <p>© 2025 dBtree</p>
      </footer>

      {showLoginModal && <LoginModal onClose={handleCloseModal} />}
    </div>
  );
};

export default Home;
