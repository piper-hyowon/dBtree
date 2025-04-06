import { useEffect, useRef } from "react";
import * as THREE from "three";
import { useTheme } from "../../../hooks/useTheme";

const BASKET_POSITION = new THREE.Vector3(2, -1.3, 1);

interface BasketProps {
  scene: THREE.Scene | null;
  renderer: THREE.WebGLRenderer | null;
  camera: THREE.PerspectiveCamera | null;
  onHarvest: (lemonId: string) => void;
}

const Basket: React.FC<BasketProps> = ({
  scene,
  onHarvest,
  renderer,
  camera,
}) => {
  const basketRef = useRef<THREE.Group | null>(null);
  const collisionAreaRef = useRef<THREE.Mesh | null>(null);
  const eventListenerRef = useRef<((event: MouseEvent) => void) | null>(null);
  const { theme } = useTheme();

  const createBasket = () => {
    if (!scene || !camera) {
      return null;
    }

    if (basketRef.current) {
      scene.remove(basketRef.current);
      basketRef.current = null;
    }

    if (collisionAreaRef.current) {
      scene.remove(collisionAreaRef.current);
      collisionAreaRef.current = null;
    }

    const basketGroup = new THREE.Group();
    basketGroup.position.copy(BASKET_POSITION);
    basketGroup.userData = { isBasket: true };

    const bodyGeometry = new THREE.CylinderGeometry(0.8, 0.6, 0.8, 16);
    const bodyMaterial = new THREE.MeshStandardMaterial({
      color: 0xc19a6b,
      roughness: 0.7,
      metalness: 0.2,
      side: THREE.DoubleSide,
      emissive: 0xc19a6b,
      emissiveIntensity: 0.4,
    });
    const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
    body.position.y = 0.3;
    body.castShadow = true;
    body.receiveShadow = true;
    basketGroup.add(body);

    const handleGeometry = new THREE.TorusGeometry(0.5, 0.05, 8, 24, Math.PI);
    const handleMaterial = new THREE.MeshStandardMaterial({
      color: 0x8b4513,
      roughness: 0.5,
      emissive: 0x8b4513,
      emissiveIntensity: 0.4,
    });
    const handle = new THREE.Mesh(handleGeometry, handleMaterial);
    handle.rotation.x = Math.PI / 2;
    handle.position.y = 0.8;
    handle.castShadow = true;
    basketGroup.add(handle);

    // 충돌 감지 영역
    const collisionGeometry = new THREE.CylinderGeometry(1.0, 0.8, 0.8, 16);
    const collisionMaterial = new THREE.MeshBasicMaterial({
      color: 0xff0000,
      transparent: true,
      opacity: 0.0, // 완전 투명
      wireframe: false,
    });

    const collisionArea = new THREE.Mesh(collisionGeometry, collisionMaterial);
    collisionArea.position.copy(basketGroup.position);
    collisionArea.position.y += 0.8; // 바구니 위에 위치
    collisionArea.userData = { isCollisionArea: true };

    // 씬에 바구니와 충돌 영역 추가
    scene.add(basketGroup);
    scene.add(collisionArea);

    basketRef.current = basketGroup;
    collisionAreaRef.current = collisionArea;

    // 렌더러가 있으면 즉시 렌더링
    if (renderer && camera) {
      renderer.render(scene, camera);
    }

    return { basketGroup, collisionArea };
  };

  // 레몬 드래그 종료 감지 이벤트 리스너
  const createEventListener = () => {
    return (event: MouseEvent) => {
      if (!scene || !camera) return;

      // 마우스 위치 정규화
      const mouse = new THREE.Vector2();
      mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
      mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;

      const raycaster = new THREE.Raycaster();
      raycaster.setFromCamera(mouse, camera);

      // 충돌 감지
      const intersects = raycaster.intersectObjects(scene.children, true);

      let isLemon = false;
      let lemonId = "";
      let isCollisionArea = false;

      // 교차 객체 확인
      for (const intersect of intersects) {
        // 레몬 찾기
        if (!isLemon && intersect.object.parent?.userData?.isLemon) {
          isLemon = true;
          lemonId = intersect.object.parent.userData.lemonId;
          console.log("Found lemon:", lemonId);
        }

        // 충돌 영역 찾기
        if (!isCollisionArea && intersect.object.userData?.isCollisionArea) {
          isCollisionArea = true;
          console.log("Found collision area");
        }

        // 둘 다 찾았으면 종료
        if (isLemon && isCollisionArea) break;
      }

      // 레몬이 충돌 영역에 있으면 수확
      if (isLemon && isCollisionArea && lemonId) {
        console.log("Harvesting lemon:", lemonId);
        onHarvest(lemonId);
      }
    };
  };

  useEffect(() => {
    if (!scene || !camera) {
      return;
    }
    createBasket();

    if (eventListenerRef.current) {
      window.removeEventListener("mouseup", eventListenerRef.current);
    }

    const handleLemonDrop = createEventListener();
    window.addEventListener("mouseup", handleLemonDrop);
    eventListenerRef.current = handleLemonDrop;

    return () => {
      if (eventListenerRef.current) {
        window.removeEventListener("mouseup", eventListenerRef.current);
        eventListenerRef.current = null;
      }

      if (basketRef.current && scene) {
        scene.remove(basketRef.current);
        basketRef.current = null;
      }

      if (collisionAreaRef.current && scene) {
        scene.remove(collisionAreaRef.current);
        collisionAreaRef.current = null;
      }
    };
  }, [scene, camera, onHarvest]);

  useEffect(() => {
    if (scene && camera) {
      setTimeout(() => {
        createBasket();
      }, 100);
    }
  }, [theme, scene, camera]);

  return null;
};

export default Basket;
