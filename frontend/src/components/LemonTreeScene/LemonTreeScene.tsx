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

const LemonTreeSceneContent = () => {
  const { containerRef, scene } = useLemonTreeScene();

  const [lemons, setLemons] = useState<AvailableLemon[]>([]);
  const [lemonsLoaded, setLemonsLoaded] = useState(false);

  const addLemonToBasket = useCallback(
    async (id: number): Promise<boolean> => {
      try {
        // API 호출 시뮬레이션
        await new Promise((resolve) => setTimeout(resolve, 300));
        const success = Math.random() > 0.2; // 80% 성공 확률

        if (success) {
          // 씬에서 레몬 모델 제거
          const lemonModel = scene.getObjectByName(`lemon-${id}`);
          if (lemonModel) {
            scene.remove(lemonModel);
          }

          // 상태에서 레몬 제거
          setLemons((prev) => prev.filter((lemon) => lemon.id !== id));

          alert("레몬이 성공적으로 바구니에 담겼습니다!");
        } else {
          alert("레몬을 바구니에 담는데 실패했습니다. 다시 시도해주세요.");
        }

        return success;
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
        />
      </div>
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
