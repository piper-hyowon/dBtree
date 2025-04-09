import "./LemonTreeScene.css";
import {
  LemonTreeSceneProvider,
  useLemonTreeScene,
} from "../../contexts/LemonTreeSceneContext";
import Basket from "./Basket";

const LemonTreeSceneContent = () => {
  const { containerRef, isLoading } = useLemonTreeScene();

  return (
    <div className="lemon-tree-container">
      <div
        ref={containerRef}
        className="lemon-tree-scene"
        id="threejs-container"
      >
        <Basket />
      </div>
      {isLoading && <div className="loading-overlay">모델 로딩 중...</div>}
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
