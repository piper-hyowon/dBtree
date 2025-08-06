import React from 'react';
import {CostResponse} from '../../types/database.types';
import {characterImages} from '../../utils/characterImages';
import './CostSummary.css';

interface CostSummaryProps {
    cost: CostResponse;
    lemonBalance: number;
}

const CostSummary: React.FC<CostSummaryProps> = ({cost, lemonBalance}) => {
    const totalInitialCost = cost.creationCost + cost.hourlyLemons; // 생성 비용 + 최소 1시간
    const canAfford = lemonBalance >= totalInitialCost;
    const runningDays = Math.floor(lemonBalance / cost.dailyLemons);

    // 레몬 잔액에 따른 캐릭터 이미지 선택
    const getCharacterImage = () => {
        const balanceRatio = lemonBalance / totalInitialCost;

        if (balanceRatio >= 1) {
            return characterImages.richInCredits;
        } else {
            return characterImages.lowCredits;
        }
    };

    const getCharacterMessage = () => {
        const balanceRatio = lemonBalance / totalInitialCost;

        if (balanceRatio >= 10) {
            return "레몬이 충분해요! 🤑";
        } else if (balanceRatio >= 3) {
            return "여유있게 생성 가능해요!";
        } else if (balanceRatio >= 1) {
            return "생성은 가능하지만 조금 빠듯해요";
        } else {
            return `${(totalInitialCost - lemonBalance).toLocaleString()}🍋 더 필요해요`;
        }
    };

    return (
        <div className="wizard-section cost-summary-section">
            <h3>비용 요약</h3>

            <div className="cost-breakdown">
                <div className="cost-item">
                    <span className="cost-label">인스턴스 생성 비용</span>
                    <span className="cost-value">{cost.creationCost} 🍋</span>
                </div>
                <div className="cost-item">
                    <span className="cost-label">
                        시간당 유지 비용
                        <span className="cost-help">실제 사용 시간만큼 과금</span>
                    </span>
                    <span className="cost-value">{cost.hourlyLemons} 🍋</span>
                </div>
                <div className="cost-separator"></div>
                <div className="cost-item total">
                    <span className="cost-label">
                        <strong>시작 시 필요한 레몬</strong>
                        <span className="cost-help">생성 비용 + 최소 1시간</span>
                    </span>
                    <span className="cost-value">
                        <strong>{totalInitialCost} 🍋</strong>
                    </span>
                </div>
            </div>

            <div className="cost-estimates">
                <h4>예상 비용</h4>
                <div className="estimate-grid">
                    <div className="estimate-item">
                        <span className="estimate-period">일간</span>
                        <span className="estimate-value">{cost.dailyLemons} 🍋</span>
                    </div>
                    <div className="estimate-item">
                        <span className="estimate-period">주간</span>
                        <span className="estimate-value">{(cost.dailyLemons * 7).toLocaleString()} 🍋</span>
                    </div>
                    <div className="estimate-item">
                        <span className="estimate-period">월간</span>
                        <span className="estimate-value">{cost.monthlyLemons.toLocaleString()} 🍋</span>
                    </div>
                </div>
            </div>

            <div className={`balance-check-with-character ${canAfford ? 'sufficient' : 'insufficient'}`}>
                <div className="balance-content">
                    <div className="balance-header">
                        <span className="balance-icon">{canAfford ? '✅' : '⚠️'}</span>
                        <span className="balance-title">레몬 잔액 확인</span>
                    </div>

                    <div className="balance-details">
                        <div className="balance-item">
                            <span>현재 보유 레몬</span>
                            <span className="balance-value">{lemonBalance.toLocaleString()} 🍋</span>
                        </div>
                        <div className="balance-item">
                            <span>필요한 레몬</span>
                            <span className="balance-value">{totalInitialCost} 🍋</span>
                        </div>
                        <div className="balance-item">
                            <span>생성 후 잔액</span>
                            <span className="balance-value">
                                {canAfford ? (lemonBalance - totalInitialCost).toLocaleString() : 0} 🍋
                            </span>
                        </div>
                    </div>

                    {canAfford ? (
                        <div className="balance-estimate">
                            이 인스턴스를 약 {runningDays}일 동안 운영할 수 있어요
                        </div>
                    ) : (
                        <div className="balance-warning">
                            레몬이 부족해요! {(totalInitialCost - lemonBalance).toLocaleString()} 🍋 더 필요합니다
                        </div>
                    )}
                </div>

                <div className="balance-character">
                    <img
                        src={getCharacterImage()}
                        alt="Lemon status"
                        className="character-image"
                    />
                    <div className="character-message">
                        {getCharacterMessage()}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default CostSummary;