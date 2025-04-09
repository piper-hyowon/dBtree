import React, { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import { DragControls } from "three/examples/jsm/controls/DragControls";
import { AvailableLemon } from "..";

interface LemonsProps {
  lemons: AvailableLemon[];
  scene: THREE.Scene;
  camera: THREE.PerspectiveCamera;
  renderer: THREE.WebGLRenderer | null;
  orbitControls: OrbitControls | null;
  onLemonDragEnd?: (id: number, position: THREE.Vector3) => void;
}

const Lemons: React.FC<LemonsProps> = ({
  lemons,
  scene,
  camera,
  renderer,
  orbitControls,
  onLemonDragEnd,
}) => {
  // 레몬 컨테이너 참조
  const lemonsContainersRef = useRef<THREE.Group[]>([]);
  const dragControlsRef = useRef<DragControls | null>(null);
  const isDraggingRef = useRef<boolean>(false);
  const lemonModelRef = useRef<THREE.Group | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // 레몬 모델 로드
  useEffect(() => {
    if (!scene || !camera || !renderer) {
      console.log("장면, 카메라 또는 렌더러가 없습니다.");
      setIsLoading(false);
      return;
    }

    // 먼저 기존 레몬 제거 - 씬에서 레몬 컨테이너를 직접 찾아 제거
    const existingLemonContainers = scene.children.filter(
      (child) => child.name && child.name.startsWith("lemon-container-")
    );
    
    existingLemonContainers.forEach((container) => {
      scene.remove(container);
    });
    
    // 트리 모델에서도 레몬 제거 확인
    const treeModel = scene.children.find(
      (child) => child.type === "Group" && child.name.includes("tree")
    ) as THREE.Group;
    
    if (treeModel) {
      const treeLemons = treeModel.children.filter(
        (child) => child.name && child.name.startsWith("lemon-container-")
      );
      
      treeLemons.forEach((container) => {
        treeModel.remove(container);
      });
    }
    
    // 참조 배열 초기화
    lemonsContainersRef.current = [];

    // 드래그 컨트롤 정리
    if (dragControlsRef.current) {
      dragControlsRef.current.dispose();
      dragControlsRef.current = null;
    }

    const loader = new GLTFLoader();
    console.log("레몬 로딩 시작...");
    setIsLoading(true); // 로딩 시작

    loader.load(
      "/models/lemon-real.gltf", // 레몬 모델 경로
      (gltf) => {
        console.log("레몬 로드 성공");
        setIsLoading(false);

        const lemonModel = gltf.scene;
        lemonModelRef.current = lemonModel;
        
        // 모델 구조 디버깅
        console.log("레몬 모델 구조:", lemonModel);
        
        // 각 레몬 데이터에 따라 레몬 생성
        lemons.forEach((item, index) => {
          // 더미 그룹을 생성하여 레몬의 부모로 사용
          const lemonContainer = new THREE.Group();
          lemonContainer.userData.id = item.id;
          lemonContainer.name = `lemon-container-${item.id}`;

          // 레몬 복제
          const lemon = lemonModel.clone();
          lemon.name = `lemon-${item.id}`;

          // 그림자 설정
          lemon.traverse((child) => {
            if ((child as THREE.Mesh).isMesh) {
              child.castShadow = true;
              child.receiveShadow = true;
            }
          });

          // 레몬을 원점(0,0,0)에 배치
          lemon.position.set(0, 0, 0);
          lemonContainer.add(lemon); // 레몬을 컨테이너의 자식으로 추가

          // 좌표계 조정
          lemonContainer.position.set(
            item.position.x,
            item.position.y,
            -item.position.z // Z 부호 반전
          );

          lemonContainer.rotation.set(
            -item.rotation.x, // X 회전 반전
            -item.rotation.y, // Y 회전 반전
            -item.rotation.z // Z 회전 반전
          );

          // 컨테이너를 씬 또는 나무(있는 경우)에 추가
          if (treeModel) {
            treeModel.add(lemonContainer);
          } else {
            scene.add(lemonContainer);
          }

          // 참조 저장 - 드래그 컨트롤용 컨테이너만 저장
          lemonsContainersRef.current.push(lemonContainer);

          // 디버그 출력
          console.log(`레몬 ${index} - ID: ${item.id}`);
          console.log(`레몬 ${index} - 원본 위치:`, item.position);
          console.log(`레몬 ${index} - 변환 위치:`, lemonContainer.position);
        });

        // 레몬이 모두 로드된 후 드래그 컨트롤 설정
        if (camera && renderer) {
          dragControlsRef.current = new DragControls(
            lemonsContainersRef.current,
            camera,
            renderer.domElement
          );

          // 드래그 시작 시 OrbitControls 비활성화
          dragControlsRef.current.addEventListener("dragstart", (event) => {
            if (orbitControls) {
              orbitControls.enabled = false;
            }
            isDraggingRef.current = true;
            
            // 디버그 - 드래그 시작된 객체
            console.log("드래그 시작:", event.object.name);
          });

          // 드래그 종료 시 OrbitControls 활성화 및 콜백 호출
          dragControlsRef.current.addEventListener("dragend", (event) => {
            if (orbitControls) {
              orbitControls.enabled = true;
            }
            isDraggingRef.current = false;

            // 드래그된 레몬의 ID와 새 위치 전달
            if (onLemonDragEnd && event.object.userData.id !== undefined) {
              // 위치를 올바른 좌표계로 변환하여 전달
              const worldPos = event.object.position.clone();
              
              // Z 축 반전 (씬에서의 좌표계와 데이터 좌표계 간 변환)
              const convertedPos = new THREE.Vector3(
                worldPos.x,
                worldPos.y,
                -worldPos.z
              );
              
              console.log("드래그 종료:", event.object.name);
              console.log("새 위치:", convertedPos);
              
              onLemonDragEnd(event.object.userData.id, convertedPos);
            }
          });
        }

        setIsLoading(false);
      },
      (xhr) => {
        // 로딩 진행률 로깅
        if (xhr.lengthComputable) {
          console.log((xhr.loaded / xhr.total) * 100 + "% 레몬 로드됨");
        }
      },
      (error) => {
        console.error("레몬 로드 오류:", error);
        setIsLoading(false); // 로딩 에러 시 로딩 상태 해제
      }
    );

    // 정리 함수
    return () => {
      // 드래그 컨트롤 정리
      if (dragControlsRef.current) {
        dragControlsRef.current.dispose();
        dragControlsRef.current = null;
      }

      // 레몬 컨테이너 제거
      lemonsContainersRef.current.forEach((container) => {
        if (container.parent) {
          container.parent.remove(container);
        }
      });
      lemonsContainersRef.current = [];
    };
  }, [lemons, scene, camera, renderer, orbitControls, onLemonDragEnd]);

  return (
    <>
      {isLoading && (
        <div
          style={{
            position: "absolute",
            bottom: "10px",
            right: "10px",
            background: "rgba(0,0,0,0.5)",
            color: "white",
            padding: "5px 10px",
            borderRadius: "4px",
            zIndex: 1000,
            fontSize: "14px",
          }}
        >
          레몬 모델 로딩 중...
        </div>
      )}
    </>
  );
};

export default Lemons;



