import React, {useEffect, useState} from "react";
import "./MiniLeaderboard.css";
import api from "../../services/api";

interface UserRank {
    maskedEmail: string;
    score: number;
    rank: number;
}

interface LeaderboardData {
    lemonRichUsers: UserRank[] | null;
    quizMasters: UserRank[] | null;
}

const MiniLeaderboard: React.FC = () => {
    const [leaderboard, setLeaderboard] = useState<LeaderboardData | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchLeaderboard = async () => {
            try {
                setIsLoading(true);
                const data = await api.home.getLeaderboardMini();
                setLeaderboard(data);
            } catch (error) {
                console.error('Failed to fetch leaderboard:', error);
                // 에러 시 null 상태 유지
                setLeaderboard({
                    lemonRichUsers: null,
                    quizMasters: null
                });
            } finally {
                setIsLoading(false);
            }
        };

        fetchLeaderboard();
        const interval = setInterval(fetchLeaderboard, 300000); // 5분마다 갱신

        return () => clearInterval(interval);
    }, []);

    const renderLeaderboardSection = (
        title: string,
        icon: string,
        users: UserRank[] | null,
        scoreLabel: string
    ) => {
        if (isLoading) {
            return (
                <div className="leaderboard-section">
                    <h3 className="leaderboard-title">
                        <span className="icon">{icon}</span> {title}
                    </h3>
                    <div className="rank-list">
                        {[1, 2, 3].map((index) => (
                            <div key={index} className={`rank-item rank-${index} loading`}>
                                <span className="rank-badge">-</span>
                                <span className="user-name">로딩 중...</span>
                                <span className="score">-</span>
                            </div>
                        ))}
                    </div>
                </div>
            );
        }

        if (!users || users.length === 0) {
            return (
                <div className="leaderboard-section">
                    <h3 className="leaderboard-title">
                        <span className="icon">{icon}</span> {title}
                    </h3>
                    <div className="rank-list">
                        <div className="no-data">아직 데이터가 없습니다</div>
                    </div>
                </div>
            );
        }

        const rankBadges = ['🥇', '🥈', '🥉'];

        return (
            <div className="leaderboard-section">
                <h3 className="leaderboard-title">
                    <span className="icon">{icon}</span> {title}
                </h3>
                <div className="rank-list">
                    {users.slice(0, 3).map((user, index) => (
                        <div key={user.maskedEmail} className={`rank-item rank-${index + 1}`}>
                            <span className="rank-badge">{rankBadges[index]}</span>
                            <span className="user-name">{user.maskedEmail}</span>
                            <span className="score">{user.score.toString()}{scoreLabel}</span>
                        </div>
                    ))}
                </div>
            </div>
        );
    };

    return (
        <div className="mini-leaderboard-container">
            <div className="mini-leaderboard">
                {renderLeaderboardSection(
                    "레몬 부자",
                    "🍋",
                    leaderboard?.lemonRichUsers || null,
                    ""
                )}

                <div className="leaderboard-divider"></div>

                {renderLeaderboardSection(
                    "오늘의 퀴즈 마스터",
                    "🏆",
                    leaderboard?.quizMasters || null,
                    "문제"
                )}
            </div>
        </div>
    );
};

export default MiniLeaderboard;