import React, { useEffect, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { DragControls } from "three/examples/jsm/controls/DragControls";

// 바구니 위치 상수
const BASKET_POSITION = new THREE.Vector3(2, 0, 2);

interface LemonManagerProps {
  scene: THREE.Scene | null;
  camera: THREE.PerspectiveCamera | null;
  renderer: THREE.WebGLRenderer | null;
  isLoggedIn: boolean;
  onLoginRequired: () => void;
}

const LemonManager: React.FC<LemonManagerProps> = ({
  scene,
  camera,
  renderer,
  isLoggedIn,
  onLoginRequired,
}) => {
  // const { lemons, harvestLemon } = useLemonStore();
  // Define the AvailableLemon type
  interface AvailableLemon {
    id: string;
    position: { x: number; y: number; z: number };
    rotation: { x: number; y: number; z: number };
    harvestableAt: string | null;
  }
  
    const lemons: AvailableLemon[] = [];
  const [lemonObjects, setLemonObjects] = useState<THREE.Object3D[]>([]);
  const [dragControls, setDragControls] = useState<DragControls | null>(null);

  // 레몬 로드 및 렌더링
  useEffect(() => {
    if (!scene || !camera || !renderer || lemons.length === 0) return;

    // 이전 레몬 객체 제거
    lemonObjects.forEach((obj) => {
      if (scene.children.includes(obj)) {
        scene.remove(obj);
      }
    });

    const newLemonObjects: THREE.Object3D[] = [];
    const loader = new GLTFLoader();

    // 수확 가능한 레몬만 로드
    const harvestableLemons = lemons.filter(
      (lemon) =>
        lemon.harvestableAt === null ||
        new Date(lemon.harvestableAt) <= new Date()
    );

    // 각 레몬 위치에 레몬 모델 배치
    const loadLemon = (index: number) => {
      if (index >= harvestableLemons.length) {
        // 모든 레몬 로드 완료 후 드래그 컨트롤 설정
        if (newLemonObjects.length > 0) {
          setupDragControls(newLemonObjects);
        }
        return;
      }

      const lemon = harvestableLemons[index];

      loader.load(
        "/models/lemon.gltf",
        (gltf) => {
          const lemonModel = gltf.scene;
          lemonModel.scale.set(0.3, 0.3, 0.3);
          lemonModel.position.set(
            lemon.position.x,
            lemon.position.y,
            lemon.position.z
          );
          lemonModel.rotation.set(
            lemon.rotation.x,
            lemon.rotation.y,
            lemon.rotation.z
          );

          // 사용자 데이터 설정
          lemonModel.userData = {
            isLemon: true,
            lemonId: lemon.id,
            originalPosition: new THREE.Vector3(
              lemon.position.x,
              lemon.position.y,
              lemon.position.z
            ),
          };

          // 그림자 설정
          lemonModel.traverse((child) => {
            if (child instanceof THREE.Mesh) {
              child.castShadow = true;
              child.receiveShadow = true;
            }
          });

          scene.add(lemonModel);
          newLemonObjects.push(lemonModel);

          // 다음 레몬 로드
          loadLemon(index + 1);
        },
        undefined,
        (error) => {
          console.error("레몬 모델 로드 오류:", error);
          // 오류가 발생해도 다음 레몬 로드 시도
          loadLemon(index + 1);
        }
      );
    };

    // 첫 번째 레몬부터 로드 시작
    loadLemon(0);

    // 상태 업데이트
    setLemonObjects(newLemonObjects);

    return () => {
      // 클린업 시 레몬 제거
      newLemonObjects.forEach((obj) => {
        if (scene.children.includes(obj)) {
          scene.remove(obj);
        }
      });

      // 드래그 컨트롤 정리
      if (dragControls) {
        dragControls.dispose();
      }
    };
  }, [scene, camera, renderer, lemons]);

  // 드래그 컨트롤 설정
  const setupDragControls = (objects: THREE.Object3D[]) => {
    if (!camera || !renderer) return;

    // 이전 드래그 컨트롤 정리
    if (dragControls) {
      dragControls.dispose();
    }

    const controls = new DragControls(objects, camera, renderer.domElement);

    controls.addEventListener("dragstart", (event) => {
      if (!isLoggedIn) {
        // 로그인하지 않은 경우 회원가입 유도
        onLoginRequired();
        controls.enabled = false;
        setTimeout(() => {
          controls.enabled = true;
        }, 100);
        return;
      }

      // 드래그 시작 시 다른 컨트롤 일시 중지 로직 추가 가능
    });

    controls.addEventListener("drag", (event) => {
      // 드래그 중 높이 제한 (바닥에 닿지 않도록)
      if (event.object.position.y < 0.5) {
        event.object.position.y = 0.5;
      }
    });

    controls.addEventListener("dragend", (event) => {
      if (!isLoggedIn) return;

      // 바구니와의 거리 계산
      const distance = event.object.position.distanceTo(BASKET_POSITION);

      // 바구니에 가까이 드롭된 경우 수확 처리
      if (distance < 2) {
        const lemonId = event.object.userData.lemonId;
        // harvestLemon(lemonId);
      } else {
        // 원래 위치로 되돌리기
        const originalPosition = event.object.userData
          .originalPosition as THREE.Vector3;
        event.object.position.copy(originalPosition);
      }
    });

    setDragControls(controls);
  };

  return null;
};

export default LemonManager;
