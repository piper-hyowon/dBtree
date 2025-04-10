import "./LemonTreeScene.css";
import {
  AvailableLemon,
  LemonTreeSceneProvider,
  useLemonTreeScene,
} from "../../contexts/LemonTreeSceneContext";
import Basket from "./Basket";
import Tree from "./Tree";
import Lemons from "./Lemons";
import { useCallback, useState } from "react";
import { mockApi } from "../../services/mockApi";

const LemonTreeSceneContent = () => {
  const { containerRef, scene } = useLemonTreeScene();

  const [lemons, setLemons] = useState<AvailableLemon[]>([]);
  const [lemonsLoaded, setLemonsLoaded] = useState(false);
  const [showInstructions, setShowinstructions] = useState<boolean>(false);

  const addLemonToBasket = useCallback(
    async (id: number): Promise<boolean> => {
      try {
        const success = (await mockApi.harvestLemon(id)).data;
        if (success) {
          // 레몬 모델 제거
          const lemonModel = scene.getObjectByName(`lemon-${id}`);
          if (lemonModel) {
            scene.remove(lemonModel);
          }
          setLemons((prev) => prev.filter((lemon) => lemon.id !== id));

          alert(
            `축하합니다! 레몬(ID: ${id})이 성공적으로 바구니에 담겼습니다!`
          );

          return success;
        } else {
          alert("레몬을 바구니에 담는데 실패했습니다. 다시 시도해주세요.");
          return false;
        }
      } catch (err) {
        console.error("바구니에 레몬 담기 오류:", err);
        alert("네트워크 오류가 발생했습니다.");
        return false;
      }
    },
    [scene]
  );

  return (
    <div className="lemon-tree-container">
      <div
        ref={containerRef}
        className="lemon-tree-scene"
        id="threejs-container"
      >
        <Basket />
        <Tree />
        <Lemons
          lemons={lemons}
          setLemons={setLemons}
          lemonsLoaded={lemonsLoaded}
          setLemonsLoaded={setLemonsLoaded}
          addLemonToBasket={addLemonToBasket}
        />
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
      <LemonTreeSceneContent />
    </LemonTreeSceneProvider>
  );
};

export default LemonTreeScene;
