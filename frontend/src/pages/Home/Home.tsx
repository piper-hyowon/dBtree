import React, { useState } from "react";
import "./Home.css";
// import LemonTree from "../../components/LemonTree/LemonTree";
import dbtreeLogo from "../../assets/images/dbtree_logo.svg";

const Home: React.FC = () => {
  const [email, setEmail] = useState("");
  const [showLoginModal, setShowLoginModal] = useState(false);

  const handleStartNow = () => {
    setShowLoginModal(true);
  };

  const handleSendOtp = () => {
    // TODO: OTP 요청 로직 구현
    alert(`${email}로 인증 코드가 발송되었습니다!`);
  };

  return (
    <div className="home-container">
      {/* 헤더 */}
      <header className="header">
        <div className="logo-container">
          <img src={dbtreeLogo} alt="dBtree Logo" className="logo" />
        </div>
        <nav className="nav">
          <button
            className="login-button"
            onClick={() => setShowLoginModal(true)}
          >
            로그인
          </button>
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

        {/* <div className="lemon-tree-container">
          <LemonTree
            onHarvest={(amount) => {
              alert(
                `${amount} 레몬 수확! 회원가입하고 더 많은 레몬을 모아보세요.`
              );
              setShowLoginModal(true);
            }}
          />
        </div> */}
      </section>

      <footer className="footer">
        <p>© 2025 dBtree</p>
      </footer>

      {/* 로그인 모달 */}
      {showLoginModal && (
        <div className="modal-overlay" onClick={() => setShowLoginModal(false)}>
          <div className="login-modal" onClick={(e) => e.stopPropagation()}>
            <button
              className="close-button"
              onClick={() => setShowLoginModal(false)}
            >
              ×
            </button>
            <img src={dbtreeLogo} alt="dBtree Logo" className="logo" />

            <h2 className="modal-title">시작하기</h2>

            <div className="input-group">
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="이메일 입력"
                className="email-input"
              />
            </div>

            <button className="cta-button modal-button" onClick={handleSendOtp}>
              인증 코드 받기
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Home;
