import React, {
  createContext,
  useContext,
  useRef,
  useState,
  useEffect,
  useCallback,
} from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import { useTheme } from "../hooks/useTheme";
import { mockApi } from "../services/mockApi";
import { LEMONS } from "../components/LemonTree/constants/lemon.constant";
import { BASKET_POSITION } from "../components/LemonTreeScene/Basket";

export interface AvailableLemon {
  id: number;
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
}

interface LemonTreeSceneContextType {
  scene: THREE.Scene;
  camera: THREE.PerspectiveCamera;
  renderer: THREE.WebGLRenderer | null;
  controls: OrbitControls | null;
  containerRef: React.RefObject<HTMLDivElement | null>;
  isLoading: boolean;
  isDraggingLemon: boolean;
  setIsDraggingLemon: (isDragging: boolean) => void;
  lemons: AvailableLemon[];
  addLemonToBasket: (id: number) => Promise<boolean>;
}

const LemonTreeSceneContext = createContext<
  LemonTreeSceneContextType | undefined
>(undefined);

export const LemonTreeSceneProvider: React.FC<{
  children: React.ReactNode;
}> = ({ children }) => {
  const { theme } = useTheme();
  const isNight = theme === "dark";
  const containerRef = useRef<HTMLDivElement>(null);

  const [isInitialized, setIsInitialized] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isDraggingLemon, setIsDraggingLemon] = useState(false);
  const [lemons, setLemons] = useState<AvailableLemon[]>([]);
  const [lemonsLoaded, setLemonsLoaded] = useState(false);

  const sceneRef = useRef<THREE.Scene>(new THREE.Scene());
  const cameraRef = useRef<THREE.PerspectiveCamera>(
    new THREE.PerspectiveCamera(45, 1, 0.1, 1000)
  );
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null);
  const controlsRef = useRef<OrbitControls | null>(null);

  const animationFrameIdRef = useRef<number | null>(null);

  const treeModelRef = useRef<THREE.Object3D | null>(null);
  const ambientLightRef = useRef<THREE.AmbientLight>(
    new THREE.AmbientLight(0xffffff)
  );
  const keyLightRef = useRef<THREE.DirectionalLight>(
    new THREE.DirectionalLight(0xffffff)
  );
  const groundRef = useRef<THREE.Mesh | null>(null);

  // 씬 초기화
  useEffect(() => {
    if (isInitialized) return;
    if (!containerRef.current || isInitialized) return;

    const container = containerRef.current;
    console.log("컨테이너 확인:", container);

    console.log("Scene initialization started, container size:", {
      width: container.clientWidth,
      height: container.clientHeight,
    });

    while (container.firstChild) {
      container.removeChild(container.firstChild);
    }

    const scene = sceneRef.current;
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

    const camera = cameraRef.current;
    camera.position.set(0, 2, 10);
    camera.aspect = container.clientWidth / container.clientHeight;
    camera.updateProjectionMatrix();

    try {
      const renderer = new THREE.WebGLRenderer({
        antialias: true,
        powerPreference: "high-performance",
      });
      renderer.setSize(container.clientWidth, container.clientHeight);
      renderer.setPixelRatio(window.devicePixelRatio);
      renderer.outputColorSpace = THREE.SRGBColorSpace;
      renderer.shadowMap.enabled = true;
      renderer.shadowMap.type = THREE.PCFSoftShadowMap;

      container.appendChild(renderer.domElement);
      rendererRef.current = renderer;

      console.log("렌더러 생성 및 DOM에 추가 완료:", renderer.domElement);

      renderer.render(scene, camera);

      const controls = new OrbitControls(camera, renderer.domElement);
      controls.enableDamping = true;
      controls.maxPolarAngle = Math.PI / 2;
      controls.minDistance = 2;
      controls.maxDistance = 15;
      controls.zoomSpeed = 0.6;
      controls.enablePan = false;
      controlsRef.current = controls;

      const ambientLight = ambientLightRef.current;
      ambientLight.intensity = isNight ? 0.01 : 1.5;

      const keyLight = keyLightRef.current;
      keyLight.position.set(3, 5, 3);
      keyLight.intensity = isNight ? 0.2 : 1.6;
      keyLight.castShadow = true;
      keyLight.shadow.mapSize.width = 2048;
      keyLight.shadow.mapSize.height = 2048;
      keyLight.shadow.bias = -0.001;

      scene.add(ambientLight);
      scene.add(keyLight);

      const groundMaterial = new THREE.MeshStandardMaterial({
        color: "#dbf4d8",
        roughness: 1,
        side: THREE.DoubleSide, // 양면 렌더링
      });

      if (isNight) {
        groundMaterial.emissive = new THREE.Color("#dbf4d8");
        groundMaterial.emissiveIntensity = 0.4;
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

      const loader = new GLTFLoader();
      console.log("트리 모델 로드 시도...");

      loader.load(
        "/models/tree-new.gltf",
        (gltf) => {
          console.log("트리 모델 로드 성공!");
          const model = gltf.scene;
          model.rotation.set(0, Math.PI, 0);
          model.position.set(0, 0, 0);
          scene.add(model);
          treeModelRef.current = model;

          fetchLemonData();
          setIsLoading(false);
        },
        (progress) => {
          console.log(
            `모델 로딩 진행률: ${Math.round(
              (progress.loaded / progress.total) * 100
            )}%`
          );
        },
        (error) => {
          console.error("모델 로드 오류:", error);
          fetchLemonData();
          setIsLoading(false);
        }
      );

      const animate = () => {
        if (!rendererRef.current) return;

        animationFrameIdRef.current = requestAnimationFrame(animate);

        if (controlsRef.current) controlsRef.current.update();
        rendererRef.current.render(sceneRef.current, cameraRef.current);
      };

      animate();

      const handleResize = () => {
        if (!containerRef.current || !rendererRef.current) return;

        const container = containerRef.current;
        const camera = cameraRef.current;
        const renderer = rendererRef.current;

        camera.aspect = container.clientWidth / container.clientHeight;
        camera.updateProjectionMatrix();
        renderer.setSize(container.clientWidth, container.clientHeight);
      };

      window.addEventListener("resize", handleResize);

      setIsInitialized(true);

      return () => {
        console.log("Three.js cleanup running");
        window.removeEventListener("resize", handleResize);

        if (animationFrameIdRef.current !== null) {
          console.log(
            "Cancelling animation frame:",
            animationFrameIdRef.current
          );
          cancelAnimationFrame(animationFrameIdRef.current);
          animationFrameIdRef.current = null;
        }

        if (renderer && container.contains(renderer.domElement)) {
          console.log("Removing renderer from container");
          container.removeChild(renderer.domElement);
        }

        if (rendererRef.current) {
          console.log("Disposing renderer");
          rendererRef.current.dispose();
          rendererRef.current = null;
        }
      };
    } catch (error) {
      console.error("렌더러 생성 오류:", error);
    }
  }, []);

  // 테마 적용
  useEffect(() => {
    if (!isInitialized) return;

    const scene = sceneRef.current;
    const ambientLight = ambientLightRef.current;
    const keyLight = keyLightRef.current;

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

    if (isNight) {
      ambientLight.intensity = 0.01;
      keyLight.intensity = 0.2;
    } else {
      ambientLight.intensity = 1.5;
      keyLight.intensity = 1.6;
    }

    if (groundRef.current) {
      const material = groundRef.current.material as THREE.MeshStandardMaterial;
      if (isNight) {
        material.emissive = new THREE.Color(material.color);
        material.emissiveIntensity = 0.4;
      } else {
        material.emissive = new THREE.Color(material.color);
        material.emissiveIntensity = 0;
      }
      material.needsUpdate = true;
    }

    if (treeModelRef.current) {
      treeModelRef.current.traverse((child) => {
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

    lemons.forEach((lemon) => {
      const lemonObj = sceneRef.current.getObjectByName(`lemon-${lemon.id}`);
      if (!lemonObj) {
        console.warn(`lemon-${lemon.id} not found in scene`);
        return;
      }

      lemonObj.traverse((child) => {
        if (child instanceof THREE.Mesh) {
          console.log(`Processing lemon mesh: ${child.name}`);

          if (Array.isArray(child.material)) {
            child.material.forEach((mat, i) => {
              if (isNight) {
                if (!mat.userData.originalColor) {
                  mat.userData.originalColor = mat.color.clone();
                }
                mat.emissive = mat.userData.originalColor.clone();
                mat.emissiveIntensity = 0.8;
              } else {
                mat.emissiveIntensity = 0;
              }
              mat.needsUpdate = true;
            });
          }
          else if (child.material) {
            const mat = child.material as THREE.MeshStandardMaterial;

            if (!mat.userData.originalColor) {
              mat.userData.originalColor = mat.color.clone();
              console.log(
                `Saved original color for ${child.name}:`,
                mat.userData.originalColor.getHexString()
              );
            }

            if (isNight) {
              mat.emissive = mat.userData.originalColor.clone();
              mat.emissiveIntensity = 0.8;
              console.log(
                `Night mode: Set emissive for ${child.name} to`,
                mat.emissive.getHexString()
              );
            } else {
              mat.emissive = new THREE.Color(0x000000);
              mat.emissiveIntensity = 0;
              console.log(`Day mode: Reset emissive for ${child.name}`);
            }
            mat.needsUpdate = true;
          }
        }
      });
    });
  }, [isNight, isInitialized, lemons]);

  useEffect(() => {
    if (!isInitialized || !controlsRef.current) return;
    controlsRef.current.enabled = !isDraggingLemon;
  }, [isDraggingLemon, isInitialized]);

  const fetchLemonData = useCallback(async () => {
    if (lemonsLoaded) return;
    try {
      const response = await mockApi.availableLemons();
      if (response.data?.lemons.length) {
        const lemonData: AvailableLemon[] = response.data.lemons.map((e) => ({
          id: e,
          position: LEMONS[e].position,
          rotation: LEMONS[e].rotation,
        }));

        setLemons(lemonData);
        console.log("레몬 데이터 로드 성공:", lemonData.length, "개의 레몬");

        const loader = new GLTFLoader();
        lemonData.forEach((lemon) => {
          loader.load(
            "/models/basic-lemon.gltf",
            (gltf) => {
              const model = gltf.scene;
              model.name = `lemon-${lemon.id}`;
              model.userData.lemonId = lemon.id;
              model.position.set(
                lemon.position.x,
                lemon.position.y,
                lemon.position.z
              );
              model.rotation.set(
                lemon.rotation.x,
                lemon.rotation.y,
                lemon.rotation.z
              );

              // 테마에 맞는 재질 설정
              model.traverse((child) => {
                if (child instanceof THREE.Mesh) {
                  const material = child.material as THREE.MeshStandardMaterial;
                  if (isNight) {
                    material.emissive = new THREE.Color(material.color);
                    material.emissiveIntensity = 0.8;
                  }
                  material.needsUpdate = true;
                }
              });

              sceneRef.current.add(model);
            },
            undefined,
            (error) => console.error(`레몬 ${lemon.id} 로드 오류:`, error)
          );
        });
      }
      setLemonsLoaded(true); // 로드 완료 표시
    } catch (err) {
      console.error("레몬 데이터 로드 오류:", err);
    }
  }, [lemonsLoaded]);

  const addLemonToBasket = useCallback(async (id: number): Promise<boolean> => {
    try {
      // API 호출 시뮬레이션
      await new Promise((resolve) => setTimeout(resolve, 300));
      const success = Math.random() > 0.2; // 80% 성공 확률

      if (success) {
        // 씬에서 레몬 모델 제거
        const lemonModel = sceneRef.current.getObjectByName(`lemon-${id}`);
        if (lemonModel) {
          sceneRef.current.remove(lemonModel);
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
  }, []);

  useEffect(() => {
    if (!isInitialized || !rendererRef.current) return;

    const renderer = rendererRef.current;
    const camera = cameraRef.current;
    const scene = sceneRef.current;

    const raycaster = new THREE.Raycaster();
    const mouse = new THREE.Vector2();

    const handleClick = async (event: MouseEvent) => {
      const rect = renderer.domElement.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

      raycaster.setFromCamera(mouse, camera);
      const intersects = raycaster.intersectObjects(scene.children, true);

      const lemonIntersect = intersects.find((intersect) => {
        let current = intersect.object;
        while (current && !current.userData.lemonId) {
          current = current.parent as THREE.Object3D;
        }
        return current?.userData?.lemonId;
      });

      if (lemonIntersect) {
        let current = lemonIntersect.object;
        while (current && !current.userData.lemonId) {
          current = current.parent as THREE.Object3D;
        }

        if (current?.userData?.lemonId) {
          const lemonId = current.userData.lemonId;

          const startPosition = current.position.clone();
          const basketPosition = BASKET_POSITION.clone();

          const animateSelectToBasket = () => {
            animateMoveToBasket();
          };

          const animateMoveToBasket = async () => {
            const duration = 1000; // 1초
            const startTime = Date.now();

            const animate = () => {
              const elapsedTime = Date.now() - startTime;
              const progress = Math.min(elapsedTime / duration, 1);

              const pos = new THREE.Vector3();
              pos.x =
                startPosition.x +
                (basketPosition.x - startPosition.x) * progress;
              pos.y = startPosition.y + 2 * Math.sin(progress * Math.PI); // 아치형 경로
              pos.z =
                startPosition.z +
                (basketPosition.z - startPosition.z) * progress;

              current.position.copy(pos);

              if (progress < 1) {
                requestAnimationFrame(animate);
              } else {
                // 애니메이션 완료 후 API 호출
                addLemonToBasket(lemonId);
              }
            };

            animate();
          };

          animateSelectToBasket();
        }
      }
    };

    renderer.domElement.addEventListener("click", handleClick);

    return () => {
      renderer.domElement.removeEventListener("click", handleClick);
    };
  }, [isInitialized]);

  const contextValue: LemonTreeSceneContextType = {
    scene: sceneRef.current,
    camera: cameraRef.current,
    renderer: rendererRef.current,
    controls: controlsRef.current,
    containerRef,
    isLoading,
    isDraggingLemon,
    setIsDraggingLemon,
    lemons,
    addLemonToBasket,
  };

  return (
    <LemonTreeSceneContext.Provider value={contextValue}>
      {children}
    </LemonTreeSceneContext.Provider>
  );
};

export const useLemonTreeScene = () => {
  const context = useContext(LemonTreeSceneContext);
  if (context === undefined) {
    throw new Error(
      "useLemonTreeScene must be used within a LemonTreeSceneProvider"
    );
  }
  return context;
};
