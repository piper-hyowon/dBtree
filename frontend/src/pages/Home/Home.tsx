import React, { useState } from "react";
import "./Home.css";
import dbtreeLogo from "../../assets/images/dbtree_logo.svg";
import LemonTree from "../../components/LemonTree/LemonTree";
import ToggleThemeButton from "../../components/common/ToggleThemeButton/ToggleThemeButton";
import LoginModal from "../../components/auth/LoginModal";
import { useAuth } from "../../hooks/useAuth";

const Home: React.FC = () => {
  const [showLoginModal, setShowLoginModal] = useState(false);
  const { isLoggedIn, user } = useAuth();

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
              className="dashboard-button"
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
            레몬을 수확하고
            <br />
            <span className="highlight">데이터베이스</span>를 키우세요
          </h1>
          <p className="hero-subtitle">
            쉽고 효율적인 크레딧 기반 DBaaS, 지금 경험해 보세요.
          </p>
          <button className="cta-button" onClick={handleStartNow}>
            무료로 시작하기
          </button>
        </div>

        <div className="lemon-tree-container">
          <LemonTree />
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
