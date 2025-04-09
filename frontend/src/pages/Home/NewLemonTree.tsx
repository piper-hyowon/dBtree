import React, { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import { DragControls } from "three/examples/jsm/controls/DragControls";
import { LEMONS } from "../../components/LemonTree/constants/lemon.constant";
import { useAuth } from "../../hooks/useAuth";

import "./NewLemonTree.css";

interface LemonTreeAppProps {
  avaiableLemonIds: number[];
}

const LemonTreeApp: React.FC<LemonTreeAppProps> = ({ avaiableLemonIds }) => {
  // const { isLoggedIn } = useAuth();
  const isLoggedIn = true;
  const [showDragInstruction, setShowDragInstruction] = useState(false);

  const containerRef = useRef<HTMLDivElement>(null);
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null);
  const sceneRef = useRef<THREE.Scene | null>(null);
  const treeModelRef = useRef<THREE.Group | null>(null);
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null);
  const orbitControlsRef = useRef<OrbitControls | null>(null);
  const dragControlsRef = useRef<DragControls | null>(null);
  const lemonsContainersRef = useRef<THREE.Group[]>([]);
  const lemonsRef = useRef<THREE.Group[]>([]);
  const requestRef = useRef<number | null>(null);
  const isDraggingRef = useRef<boolean>(false);

  const onLemonDragEnd = (id: number, position: THREE.Vector3) => {
    console.log(`레몬 ${id}가 새 위치로 이동됨:`, position);
    // 여기서 상태를 업데이트하거나 API 호출 등을 수행할 수 있습니다
  };

  // 씬 초기화 - 컴포넌트 마운트 시 한 번만 실행
  useEffect(() => {
    if (!containerRef.current) return;

    // 기존 캔버스 확인 및 제거
    const existingCanvas = containerRef.current.querySelector("canvas");
    if (existingCanvas) {
      containerRef.current.removeChild(existingCanvas);
    }

    // 씬 생성
    const scene = new THREE.Scene();
    sceneRef.current = scene;

    // 카메라 생성
    const camera = new THREE.PerspectiveCamera(
      75,
      containerRef.current.clientWidth / containerRef.current.clientHeight,
      0.1,
      1000
    );
    camera.position.set(3, 0, -7);
    camera.lookAt(0, 1, 0);
    cameraRef.current = camera;

    // 렌더러 생성
    const renderer = new THREE.WebGLRenderer({
      antialias: true,
      alpha: true,
    });
    renderer.setSize(
      containerRef.current.clientWidth,
      containerRef.current.clientHeight
    );
    renderer.setClearColor(0xffffff, 1); // 밝은 배경색 설정
    renderer.outputColorSpace = THREE.SRGBColorSpace;
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.0;
    renderer.shadowMap.enabled = true;
    containerRef.current.appendChild(renderer.domElement);
    rendererRef.current = renderer;

    // 조명 설정 - 더 밝게 조정
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.5);
    scene.add(ambientLight);

    const directionalLight = new THREE.DirectionalLight(0xffffff, 1);
    directionalLight.position.set(5, 10, 7);
    directionalLight.castShadow = true;
    scene.add(directionalLight);

    const directionalLight2 = new THREE.DirectionalLight(0xffffff, 0.8);
    directionalLight2.position.set(-5, 8, -7);
    scene.add(directionalLight2);

    // OrbitControls 설정
    const orbitControls = new OrbitControls(camera, renderer.domElement);
    orbitControls.target.set(0, 1, 0);
    orbitControls.update();
    orbitControlsRef.current = orbitControls;

    // 창 크기 변경 핸들러
    const handleResize = () => {
      if (!containerRef.current || !cameraRef.current || !rendererRef.current)
        return;

      const width = containerRef.current.clientWidth;
      const height = containerRef.current.clientHeight;

      cameraRef.current.aspect = width / height;
      cameraRef.current.updateProjectionMatrix();
      rendererRef.current.setSize(width, height);
    };

    // 리사이즈 이벤트 리스너 등록
    window.addEventListener("resize", handleResize);

    // 초기 리사이즈 트리거
    handleResize();

    // 애니메이션 루프
    const animate = () => {
      requestRef.current = requestAnimationFrame(animate);

      if (sceneRef.current && cameraRef.current && rendererRef.current) {
        if (!isDraggingRef.current && orbitControlsRef.current) {
          orbitControlsRef.current.update();
        }

        rendererRef.current.render(sceneRef.current, cameraRef.current);
      }
    };

    // 애니메이션 시작
    animate();

    if (isLoggedIn) {
      setTimeout(() => {
        setShowDragInstruction(true);
        setTimeout(() => setShowDragInstruction(false), 5000);
      }, 2000);
    }

    // 정리 함수
    return () => {
      if (requestRef.current !== null) {
        cancelAnimationFrame(requestRef.current);
      }

      if (rendererRef.current && containerRef.current) {
        const canvas = containerRef.current.querySelector("canvas");
        if (canvas) {
          containerRef.current.removeChild(canvas);
        }
        rendererRef.current.dispose();
      }

      if (orbitControlsRef.current) {
        orbitControlsRef.current.dispose();
      }

      if (dragControlsRef.current) {
        dragControlsRef.current.dispose();
      }

      window.removeEventListener("resize", handleResize);
    };
  }, []); // 빈 의존성 배열 - 마운트 시에만 실행

  // 나무 모델 로드
  useEffect(() => {
    if (!sceneRef.current) return;

    const loader = new GLTFLoader();

    // 로딩 중 표시 (옵션)
    console.log("나무 로딩 시작...");

    loader.load(
      "/models/tree-new.gltf", // 나무 모델 경로
      (gltf) => {
        console.log("나무 로드 성공");

        const treeModel = gltf.scene;

        // 그림자 설정
        treeModel.traverse((child) => {
          if ((child as THREE.Mesh).isMesh) {
            child.castShadow = true;
            child.receiveShadow = true;
          }
        });

        // 나무 모델의 방향 수정
        treeModel.rotation.set(0, Math.PI, 0); // Y축 기준 180도 회전

        treeModel.scale.set(1, 1, 1);
        treeModel.position.set(0, 0, 0);

        sceneRef.current?.add(treeModel);
        treeModelRef.current = treeModel;
      },
      (xhr) => {
        console.log((xhr.loaded / xhr.total) * 100 + "% 나무 로드됨");
      },
      (error) => {
        console.error("나무 로드 오류:", error);
      }
    );
  }, []);

  // 레몬 모델 로드 및 드래그 컨트롤 설정
  useEffect(() => {
    if (
      !sceneRef.current ||
      !cameraRef.current ||
      !rendererRef.current ||
      !treeModelRef.current
    )
      return;

    // 먼저 기존 레몬 제거
    lemonsContainersRef.current.forEach((container) => {
      container.parent?.remove(container);
    });
    lemonsContainersRef.current = [];
    lemonsRef.current = [];

    // 드래그 컨트롤 정리
    if (dragControlsRef.current) {
      dragControlsRef.current.dispose();
      dragControlsRef.current = null;
    }

    const loader = new GLTFLoader();
    console.log("레몬 로딩 시작...");

    loader.load(
      "/models/basic-lemon.gltf", // 레몬 모델 경로
      (gltf) => {
        console.log("레몬 로드 성공");

        const lemonModel = gltf.scene;

        // 각 레몬 데이터에 따라 레몬 생성
        avaiableLemonIds
          .map((e) => LEMONS[e])
          .forEach((item, index) => {
            // 더미 그룹을 생성하여 레몬의 부모로 사용
            const lemonContainer = new THREE.Group();
            lemonContainer.userData.id = item.id;

            // 레몬 복제
            const lemon = lemonModel.clone();

            // 그림자 설정
            lemon.traverse((child) => {
              if ((child as THREE.Mesh).isMesh) {
                child.castShadow = true;
                child.receiveShadow = true;
              }
            });

            // 레몬을 원점(0,0,0)에 배치
            lemonContainer.add(lemon); // 레몬을 컨테이너의 자식으로 추가

            // 컨테이너의 위치/회전 설정
            lemonContainer.position.set(
              item.position.x,
              item.position.y,
              -item.position.z // Z 부호 반전
            );

            lemonContainer.rotation.set(
              item.rotation.x, // X 회전 반전
              item.rotation.y, // Y 회전 반전
              -item.rotation.z // Z 회전 반전
            );

            // 컨테이너를 나무에 추가
            treeModelRef.current?.add(lemonContainer);

            // 참조 저장
            lemonsContainersRef.current.push(lemonContainer); // 드래그 컨트롤용
            // lemonsRef.current.push(lemon); // 실제 레몬 모델

            // 디버그 출력
            console.log(`레몬 ${index} - ID: ${item.id}`);
            console.log(`레몬 ${index} - 원본 위치:`, item.position);
            console.log(`레몬 ${index} - 변환 위치:`, lemonContainer.position);
            console.log(
              `레몬 ${index} - 월드 위치:`,
              lemonContainer.getWorldPosition(new THREE.Vector3())
            );
          });

        // 레몬이 모두 로드된 후 드래그 컨트롤 설정
        if (cameraRef.current && rendererRef.current) {
          dragControlsRef.current = new DragControls(
            lemonsContainersRef.current,
            cameraRef.current,
            rendererRef.current.domElement
          );

          // 드래그 시작 시 OrbitControls 비활성화
          dragControlsRef.current.addEventListener("dragstart", () => {
            if (orbitControlsRef.current) {
              orbitControlsRef.current.enabled = false;
            }
            isDraggingRef.current = true;
          });

          // 드래그 중 레몬 위치 업데이트
          dragControlsRef.current.addEventListener("drag", () => {
            // 추가 처리가 필요한 경우
          });

          // 드래그 종료 시 OrbitControls 활성화 및 콜백 호출
          dragControlsRef.current.addEventListener("dragend", (event) => {
            if (orbitControlsRef.current) {
              orbitControlsRef.current.enabled = true;
            }
            isDraggingRef.current = false;

            // 드래그된 레몬의 ID와 새 위치 전달
            if (onLemonDragEnd && event.object.userData.id !== undefined) {
              onLemonDragEnd(
                event.object.userData.id,
                event.object.position.clone()
              );
            }
          });
        }
      },
      (xhr) => {
        console.log((xhr.loaded / xhr.total) * 100 + "% 레몬 로드됨");
      },
      (error) => {
        console.error("레몬 로드 오류:", error);
      }
    );
  }, [avaiableLemonIds, onLemonDragEnd]); // lemonData나 onLemonDragEnd가 변경될 때 실행

  return (
    <div
      className="lemon-tree-container"
      ref={containerRef}
      style={{
        width: "100%",
        height: "100%",
        minHeight: "500px",
        position: "relative",
      }}
    >
      {isLoggedIn && (
        <div
          className="credits-display"
          title="레몬은 DB 인스턴스를 생성하는 데 사용됩니다"
        >
          <span>🍋 10</span>
        </div>
      )}

      <div
        className={`drag-instruction ${showDragInstruction ? "visible" : ""}`}
      >
        레몬을 바구니로 드래그하여 수확하세요
      </div>
    </div>
  );
};

export default LemonTreeApp;
