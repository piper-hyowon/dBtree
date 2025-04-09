import { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import "./BasicLemonTree.css";
import { useTheme } from "../../../hooks/useTheme";

interface BasicLemonTreeProps {
  onSceneCreated?: (
    scene: THREE.Scene,
    camera: THREE.PerspectiveCamera,
    renderer: THREE.WebGLRenderer,
    orbitControls: OrbitControls
  ) => void;
}

const BasicLemonTree: React.FC<BasicLemonTreeProps> = ({ onSceneCreated }) => {
  const mountRef = useRef<HTMLDivElement | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const sceneRef = useRef<THREE.Scene | null>(null);
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null);
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null);
  const controlsRef = useRef<OrbitControls | null>(null);
  const { theme } = useTheme();
  const isNight = theme === "dark";
  const animationFrameRef = useRef<number | null>(null);
  const modelRef = useRef<THREE.Object3D | null>(null);

  // 조명 참조 추가
  const ambientLightRef = useRef<THREE.AmbientLight | null>(null);
  const keyLightRef = useRef<THREE.DirectionalLight | null>(null);
  const groundRef = useRef<THREE.Mesh | null>(null);

  // 씬 초기화 - 테마 독립적
  useEffect(() => {
    if (!mountRef.current) {
      console.error("Mount ref is not available");
      return;
    }

    console.log("씬 초기화 시작...");
    const currentMount = mountRef.current;

    // 기존 렌더러가 있으면 제거
    if (
      rendererRef.current &&
      currentMount.contains(rendererRef.current.domElement)
    ) {
      console.log("Removing existing renderer");
      currentMount.removeChild(rendererRef.current.domElement);
      rendererRef.current.dispose();
      rendererRef.current = null;
    }

    // 애니메이션 프레임 취소
    if (animationFrameRef.current !== null) {
      cancelAnimationFrame(animationFrameRef.current);
      animationFrameRef.current = null;
    }

    // 새 씬 생성
    const scene = new THREE.Scene();
    sceneRef.current = scene;

    // 카메라 생성
    const camera = new THREE.PerspectiveCamera(
      45,
      currentMount.clientWidth / currentMount.clientHeight,
      0.1,
      1000
    );

    camera.position.set(4, -1, 6);
    // camera.rotation.set(Math.PI,0,0)
    cameraRef.current = camera;

    // 렌더러 생성
    const renderer = new THREE.WebGLRenderer({
      antialias: true,
      powerPreference: "high-performance",
      alpha: false,
    });
    rendererRef.current = renderer;
    renderer.setSize(currentMount.clientWidth, currentMount.clientHeight);
    renderer.setPixelRatio(window.devicePixelRatio);
    renderer.outputColorSpace = THREE.SRGBColorSpace;
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;

    // DOM에 렌더러 추가
    currentMount.appendChild(renderer.domElement);

    // 컨트롤 설정
    const controls = new OrbitControls(camera, renderer.domElement);
    controlsRef.current = controls;
    controls.enableDamping = true;
    controls.maxPolarAngle = Math.PI / 2;
    controls.minDistance = 10 * 0.2;
    controls.maxDistance = 10 * 1.5;
    controls.zoomSpeed = 0.6;
    controls.enablePan = false;

    scene.background = isNight
      ? new THREE.Color(
          getComputedStyle(document.documentElement)
            .getPropertyValue("--background-dark")
            .trim() || "#121212"
        )
      : new THREE.Color(
          getComputedStyle(document.documentElement)
            .getPropertyValue("--background")
            .trim() || "#ffffff"
        );

    // 조명 설정도 초기 테마에 맞게 조정
    // 조명 설정
    const ambientLight = new THREE.AmbientLight(0xffffff);
    ambientLightRef.current = ambientLight;

    const keyLight = new THREE.DirectionalLight(0xffffff);
    keyLight.position.set(3, 2, 3);
    keyLight.castShadow = true;
    keyLight.shadow.mapSize.width = 2048;
    keyLight.shadow.mapSize.height = 2048;
    keyLight.shadow.bias = -0.001;
    keyLightRef.current = keyLight;

    // 테마에 맞는 초기 조명 세기 설정
    if (isNight) {
      ambientLight.intensity = 0.01;
      keyLight.intensity = 0.2;
    } else {
      ambientLight.intensity = 1.5;
      keyLight.intensity = 1.6;
    }

    scene.add(ambientLight);
    scene.add(keyLight);

    // 바닥 생성도 초기 테마에 맞게 조정
    const groundMaterial = new THREE.MeshStandardMaterial({
      color: "#dbf4d8",
      roughness: 1,
    });

    // 테마에 맞는 초기 emissive 설정
    if (isNight) {
      groundMaterial.emissive = new THREE.Color("#dbf4d8");
      groundMaterial.emissiveIntensity = 0.4;
    } else {
      groundMaterial.emissive = new THREE.Color(0x000000);
      groundMaterial.emissiveIntensity = 0;
    }

    const ground = new THREE.Mesh(
      new THREE.BoxGeometry(5.3, 3, 0.2),
      groundMaterial
    );
    ground.rotation.x = -Math.PI / 2;
    ground.position.y = -1.45;
    ground.receiveShadow = true;
    scene.add(ground);
    groundRef.current = ground;

    // 3D 모델 로드
    const loader = new GLTFLoader();
    console.log("모델 로딩 시작...");
    loader.load(
      "/models/tree-new.gltf",
      (gltf) => {
        const model = gltf.scene;
        modelRef.current = model;
        model.name = "tree-model"; // 이름 지정
        model.userData.isTree = true; // userData에 표시

        model.rotation.set(0, Math.PI, 0); // Y축 기준 180도 회전
        model.scale.set(1, 1, 1);
        model.position.set(0, 0, 0);

        scene.add(model);

        // 상위 컴포넌트에 전달
        if (onSceneCreated) {
          onSceneCreated(scene, camera, renderer, controls);
        }

        console.log("모델 로드 완료");
        setIsLoading(false); // 로딩 완료
      },
      (progress) => {
        console.log(
          `Loading progress: ${(progress.loaded / progress.total) * 100}%`
        );
      },
      (error) => {
        console.error("모델 로드 오류:", error);
        setIsLoading(false);
      }
    );
    

    // 애니메이션 함수
    const animate = () => {
      if (!rendererRef.current) return;

      animationFrameRef.current = requestAnimationFrame(animate);

      if (controlsRef.current) {
        controlsRef.current.update();
      }

      if (sceneRef.current && cameraRef.current) {
        rendererRef.current.render(sceneRef.current, cameraRef.current);
      }
    };

    // 애니메이션 시작
    animate();

    // 창 크기 변경 처리
    const handleResize = () => {
      if (!currentMount || !cameraRef.current || !rendererRef.current) return;

      cameraRef.current.aspect =
        currentMount.clientWidth / currentMount.clientHeight;
      cameraRef.current.updateProjectionMatrix();
      rendererRef.current.setSize(
        currentMount.clientWidth,
        currentMount.clientHeight
      );
    };

    window.addEventListener("resize", handleResize);

    // 클린업 함수
    return () => {
      console.log("BasicLemonTree: cleaning up...");
      setIsLoading(false); // 컴포넌트 언마운트 시 로딩 상태 초기화
      window.removeEventListener("resize", handleResize);

      if (animationFrameRef.current !== null) {
        cancelAnimationFrame(animationFrameRef.current);
        animationFrameRef.current = null;
      }

      if (controlsRef.current) {
        controlsRef.current.dispose();
        controlsRef.current = null;
      }

      if (rendererRef.current) {
        rendererRef.current.dispose();
        if (currentMount.contains(rendererRef.current.domElement)) {
          currentMount.removeChild(rendererRef.current.domElement);
        }
        rendererRef.current = null;
      }

      // 씬 정리
      if (sceneRef.current) {
        // 씬의 모든 객체 제거
        while (sceneRef.current.children.length > 0) {
          sceneRef.current.remove(sceneRef.current.children[0]);
        }
        sceneRef.current = null;
      }

      cameraRef.current = null;
      modelRef.current = null;
      ambientLightRef.current = null;
      keyLightRef.current = null;
      groundRef.current = null;
    };
  }, []); // 의존성 배열에서 isNight 제거 - 씬은 한 번만 초기화

  // 테마 변경 처리 - 별도의 useEffect
  useEffect(() => {
    console.log("테마 변경 감지:", isNight ? "다크 모드" : "라이트 모드");

    if (!sceneRef.current || !modelRef.current) {
      console.log("테마 변경 적용 실패: 씬 또는 모델이 아직 준비되지 않음");
      return;
    }

    // 배경색 변경
    if (sceneRef.current) {
      sceneRef.current.background = isNight
        ? new THREE.Color(
            getComputedStyle(document.documentElement)
              .getPropertyValue("--background-dark")
              .trim() || "#121212"
          )
        : new THREE.Color(
            getComputedStyle(document.documentElement)
              .getPropertyValue("--background")
              .trim() || "#ffffff"
          );
    }

    // 조명 세기 조정
    if (ambientLightRef.current && keyLightRef.current) {
      if (isNight) {
        ambientLightRef.current.intensity = 0.01;
        keyLightRef.current.intensity = 0.2;
      } else {
        ambientLightRef.current.intensity = 1.5;
        keyLightRef.current.intensity = 1.6;
      }
    }

    // 바닥 재질 변경
    if (groundRef.current) {
      const groundMaterial = groundRef.current
        .material as THREE.MeshStandardMaterial;

      if (isNight) {
        groundMaterial.emissive = new THREE.Color("#dbf4d8");
        groundMaterial.emissiveIntensity = 0.4;
      } else {
        groundMaterial.emissive = new THREE.Color(0x000000);
        groundMaterial.emissiveIntensity = 0;
      }

      groundMaterial.needsUpdate = true;
    }

    // 트리 모델 재질 변경
    if (modelRef.current) {
      modelRef.current.traverse((child) => {
        if (child instanceof THREE.Mesh) {
          if (!child.name.includes("body")) {
            const material = child.material as THREE.MeshStandardMaterial;

            if (isNight) {
              material.emissive = new THREE.Color(material.color);
              material.emissiveIntensity = 0.8;
            } else {
              material.emissive = new THREE.Color(material.color);
              material.emissiveIntensity = 0;
            }

            material.needsUpdate = true;
          }
        }
      });
    }

    // 렌더링 업데이트
    if (rendererRef.current && sceneRef.current && cameraRef.current) {
      console.log("테마 변경 적용 후 렌더링");
      rendererRef.current.render(sceneRef.current, cameraRef.current);
    }
  }, [isNight]); // 테마 변경 시에만 실행

  return (
    <div className="gltf-container">
      <div ref={mountRef} className="gltf-mount" />
      {isLoading && <div className="gltf-loading">모델 로딩 중...</div>}
    </div>
  );
};

export default BasicLemonTree;
