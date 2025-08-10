import "./LemonTreeScene.css";
import {
    AvailableLemon,
    LemonTreeSceneProvider,
    useLemonTreeScene,
} from "../../contexts/LemonTreeSceneContext";
import Basket from "./Basket";
import Tree from "./Tree";
import Lemons from "./Lemons";
import {useState} from "react";

const LemonTreeSceneContent = () => {
    const {containerRef, scene} = useLemonTreeScene();

    const [_lemons, setLemons] = useState<AvailableLemon[]>([]);
    const [lemonsLoaded, setLemonsLoaded] = useState(false);
    const [showInstructions, setShowinstructions] = useState<boolean>(false);
    const [availableLemonCount, setAvailableLemonCount] = useState<number>(0);
    const [nextGrowthTime, setNextGrowthTime] = useState<string | null>(null);

    return (
        <div className="lemon-tree-container">
            <div
                ref={containerRef}
                className="lemon-tree-scene"
                id="threejs-container"
            >
                <Basket/>
                <Tree/>
                <Lemons
                    setLemons={setLemons}
                    lemonsLoaded={lemonsLoaded}
                    setLemonsLoaded={setLemonsLoaded}
                    setAvailableLemonCount={setAvailableLemonCount}
                    setNextGrowthTime={setNextGrowthTime}
                />
            </div>

            {/* 레몬 상태 표시 */}
            <div
                style={{
                    position: "absolute",
                    top: "3px",
                    left: "25%",
                    transform: "translateX(-50%)",
                    background: "rgba(255, 255, 255, 0.7)",
                    color: "#666",
                    padding: "6px 14px",
                    borderRadius: "15px",
                    zIndex: 500,
                    fontSize: "13px",
                    fontWeight: "normal",
                    boxShadow: "0 1px 4px rgba(0,0,0,0.08)",
                    backdropFilter: "blur(5px)",
                    border: "1px solid rgba(111, 207, 151, 0.2)",
                    pointerEvents: "none",
                }}
            >
                {availableLemonCount > 0 ? (
                    <span>🍋 <span
                        style={{color: "#4c9067", fontWeight: "500"}}>{availableLemonCount}개</span> 수확 가능</span>
                ) : (
                    <span style={{fontSize: "12px"}}>
            {nextGrowthTime ? (
                <> ↻ 다음 수확: {new Date(nextGrowthTime).toLocaleTimeString('ko-KR', {
                    hour: '2-digit',
                    minute: '2-digit'
                })}</>
            ) : (
                "곧 자라날 예정"
            )}
          </span>
                )}
            </div>

            <div className="instructions-button-container">
                <button
                    className="instructions-button"
                    onClick={() => setShowinstructions(!showInstructions)}
                    aria-label="레몬 수확 방법"
                >
                    <span className="button-content">?</span>
                </button>
            </div>

            {showInstructions && (
                <div className="instructions-modal">
                    <div className="instructions-content">
                        <div className="modal-header">
                            <h3>레몬 수확 방법</h3>
                            <button
                                className="close-button-icon"
                                onClick={() => setShowinstructions(false)}
                                aria-label="닫기"
                            >
                                ×
                            </button>
                        </div>
                        <ol>
                            <li>
                                나무에서 <span className="highlight">노란색 레몬</span>을
                                클릭하세요
                            </li>
                            <li>
                                <span className="highlight">DB 관련 퀴즈</span>에 정답을
                                선택하세요
                            </li>
                            <li>
                                정답 선택 후 나타나는{" "}
                                <span className="highlight">노란색 타겟</span>을 빠르게
                                클릭하세요
                            </li>
                        </ol>
                        <div className="tip-container">
                            <p className="tip">
                                퀴즈 정답을 맞추고 움직이는 타겟을 클릭해야 크레딧을 얻을 수
                                있습니다!
                            </p>
                        </div>
                        <div className="modal-footer">
                            <button
                                className="close-button"
                                onClick={() => setShowinstructions(false)}
                            >
                                닫기
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

const LemonTreeScene = () => {
    return (
        <LemonTreeSceneProvider>
            <LemonTreeSceneContent/>
        </LemonTreeSceneProvider>
    );
};

export default LemonTreeScene;