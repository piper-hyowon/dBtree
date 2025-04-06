import { useEffect, useState } from "react";
import { mockApi } from "../../services/mockApi";
import "./GlobalStats.css";

const GlobalStats: React.FC = () => {
  const [globalStats, setGlobalStats] = useState({
    totalHarvested: 0,
    totalDbInstances: 0,
    activeUsers: 0,
  });

  useEffect(() => {
    const fetchGlobalStats = async () => {
      try {
        const response = await mockApi.globalStats();
        if (response.data) {
          setGlobalStats(response.data);
        }
      } catch (error) {
        console.error("전역 통계 로드 실패:", error);
      }
    };

    fetchGlobalStats();
    const interval = setInterval(fetchGlobalStats, 60000); // 1분마다 갱신

    return () => clearInterval(interval);
  }, []);

  return (
    <>
      <div className="stat-item">
        <span className="stat-number">
          {globalStats.totalHarvested.toLocaleString()}
        </span>
        <span className="stat-label">수확된 레몬</span>
      </div>
      <div className="stat-item">
        <span className="stat-number">
          {globalStats.totalDbInstances.toLocaleString()}
        </span>
        <span className="stat-label">생성된 DB</span>
      </div>
      <div className="stat-item">
        <span className="stat-number">
          {globalStats.activeUsers.toLocaleString()}
        </span>
        <span className="stat-label">오늘의 방문자</span>
      </div>
    </>
  );
};

export default GlobalStats;
