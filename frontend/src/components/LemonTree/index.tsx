import React, { useRef, useState, useEffect } from "react";
import * as THREE from "three";
import { useAuth } from "../../hooks/useAuth";
import { useTheme } from "../../hooks/useTheme";
import BasicLemonTree from "./BasicLemonTree/BasicLemonTree";
import "./LemonTree.css";
import Basket from "./Basket/Basket";
import Lemon from "./Lemon/Lemon";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import { mockApi } from "../../services/mockApi";
import Lemons from "./Lemons/Lemons";
import { LEMONS } from "./constants/lemon.constant";

export interface AvailableLemon {
  id: number;
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
}

interface LemonTreeProps {
  onLoginRequired?: () => void;
}

const LemonTree: React.FC<LemonTreeProps> = ({
  onLoginRequired = () => {},
}) => {
  const sceneRef = useRef<THREE.Scene | null>(null);
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null);
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null);
  const [orbitControls, setOrbitControls] = useState<OrbitControls | null>(
    null
  );
  const { isLoggedIn } = useAuth();
  const [showDragInstruction, setShowDragInstruction] = useState(false);
  const [sceneInitialized, setSceneInitialized] = useState(false);
  const [reloadBasket, setReloadBasket] = useState(0); // 바구니 강제 리로드용 카운터

  const [availableLemons, setAvailableLemons] = useState<AvailableLemon[]>([]);

  useEffect(() => {
    if (sceneInitialized) {
      setTimeout(() => {
        setReloadBasket((prev) => prev + 1);
      }, 200);

      const fetchAvailableLemons = async () => {
        try {
          const response = await mockApi.availableLemons();
          console.log(response);
          if (response.data?.lemons.length) {
            setAvailableLemons(
              response.data.lemons.map((e: number) => LEMONS[e])
            );
          }
        } catch (error) {
          console.error("전역 통계 로드 실패:", error);
        }
      };

      fetchAvailableLemons();
    }
  }, [sceneInitialized]);

  const handleLemonDragEnd = (id: number, position: THREE.Vector3) => {
    console.log(`레몬 ${id}가 새 위치로 이동됨:`, position);
    // 여기서 상태를 업데이트하거나 API 호출 등을 수행할 수 있습니다
  };

  const handleSceneCreated = (
    scene: THREE.Scene,
    camera: THREE.PerspectiveCamera,
    renderer: THREE.WebGLRenderer,
    orbitControls: OrbitControls
  ) => {
    console.log("씬 생성됨, 참조 설정 중...");
    sceneRef.current = scene;
    cameraRef.current = camera;
    rendererRef.current = renderer;
    setOrbitControls(orbitControls);
    setSceneInitialized(true);

    if (isLoggedIn) {
      setTimeout(() => {
        setShowDragInstruction(true);
        setTimeout(() => setShowDragInstruction(false), 5000);
      }, 2000);
    }
  };

  const handleHarvest = (lemonId: string) => {
    alert(`레몬 ${lemonId}를 수확했습니다!`);
  };

  return (
    <div className="lemon-tree-container">
      <BasicLemonTree onSceneCreated={handleSceneCreated} />

      {sceneInitialized && sceneRef.current && cameraRef.current && (
        <>
          <Basket
            key={`basket-${reloadBasket}`}
            scene={sceneRef.current}
            renderer={rendererRef.current}
            camera={cameraRef.current}
            onHarvest={handleHarvest}
          />
          {/* <Lemons
            lemons={availableLemons}
            scene={sceneRef.current}
            camera={cameraRef.current}
            renderer={rendererRef.current}
            orbitControls={orbitControls}
            onLemonDragEnd={handleLemonDragEnd}
          /> */}
          {availableLemons.map((e) => (
            <Lemon
              key={`lemon-${e.id}`}
              scene={sceneRef.current}
              renderer={rendererRef.current}
              camera={cameraRef.current}
              orbitControls={orbitControls}
              id={e.id}
              position={e.position}
              rotation={e.rotation}
            />
          ))}

          {isLoggedIn && (
            <div
              className="credits-display"
              title="레몬은 DB 인스턴스를 생성하는 데 사용됩니다"
            >
              <span>🍋 10</span>
            </div>
          )}

          <div
            className={`drag-instruction ${
              showDragInstruction ? "visible" : ""
            }`}
          >
            레몬을 바구니로 드래그하여 수확하세요
          </div>
        </>
      )}
    </div>
  );
};

export default LemonTree;
