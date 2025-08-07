import {useEffect, useState} from "react";
import "./GlobalStats.css";
import api from "../../services/api";

const GlobalStats: React.FC = () => {
    const [globalStats, setGlobalStats] = useState({
        totalHarvested: 0,
        totalDbInstances: 0,
        totalUsers: 0,
    });
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchGlobalStats = async () => {
            try {
                setIsLoading(true);
                const response = await api.homeStats.getGlobalStats();
                setGlobalStats(response);
            } catch (error) {
                console.error("전역 통계 로드 실패:", error);
                // 에러 발생 시 기본값 유지
            } finally {
                setIsLoading(false);
            }
        };

        fetchGlobalStats();
        const interval = setInterval(fetchGlobalStats, 60000); // 1분마다 갱신

        return () => clearInterval(interval);
    }, []);

    if (isLoading) {
        return (
            <>
                <div className="stat-item">
                    <span className="stat-number">-</span>
                    <span className="stat-label">수확된 레몬</span>
                </div>
                <div className="stat-item">
                    <span className="stat-number">-</span>
                    <span className="stat-label">생성된 DB</span>
                </div>
                <div className="stat-item">
                    <span className="stat-number">-</span>
                    <span className="stat-label">총 가입 유저</span>
                </div>
            </>
        );
    }

    return (
        <>
            <div className="stat-item">
        <span className="stat-number">
          {globalStats.totalHarvested ?? 0}
        </span>
                <span className="stat-label">수확된 레몬</span>
            </div>
            <div className="stat-item">
        <span className="stat-number">
          {globalStats.totalDbInstances ?? 0}
        </span>
                <span className="stat-label">생성된 DB</span>
            </div>
            <div className="stat-item">
        <span className="stat-number">
          {globalStats.totalUsers ?? 0}
        </span>
                <span className="stat-label">총 가입 유저</span>
            </div>
        </>
    );
};

export default GlobalStats;