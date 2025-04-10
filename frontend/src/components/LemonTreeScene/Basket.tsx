import { useEffect, useRef } from "react";
import * as THREE from "three";
import { useTheme } from "../../hooks/useTheme";
import { useLemonTreeScene } from "../../contexts/LemonTreeSceneContext";

export const BASKET_POSITION = new THREE.Vector3(2, -1.1, 0.6);

const Basket: React.FC = () => {
  const { scene } = useLemonTreeScene();
  const basketRef = useRef<THREE.Group | null>(null);
  const collisionAreaRef = useRef<THREE.Mesh | null>(null);
  const { isNight } = useTheme();

  const createBasket = () => {
    if (!scene) return null;

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

    // 바구니 몸
    const bodyGeometry = new THREE.CylinderGeometry(0.8, 0.6, 0.8, 16);
    const bodyMaterial = new THREE.MeshStandardMaterial({
      color: 0xc19a6b,
      roughness: 0.7,
      metalness: 0.2,
      side: THREE.DoubleSide,
      emissive: 0xc19a6b,
      emissiveIntensity: isNight ? 0.4 : 0,
    });
    const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
    body.position.y = 0.3;
    body.castShadow = true;
    body.receiveShadow = true;
    basketGroup.add(body);

    // 바구니 손잡이
    const handleGeometry = new THREE.TorusGeometry(0.5, 0.05, 8, 24, Math.PI);
    const handleMaterial = new THREE.MeshStandardMaterial({
      color: 0x8b4513,
      roughness: 0.5,
      emissive: 0x8b4513,
      emissiveIntensity: isNight ? 0.4 : 0,
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
      opacity: 0.0,
      wireframe: false,
    });

    const collisionArea = new THREE.Mesh(collisionGeometry, collisionMaterial);
    collisionArea.position.copy(basketGroup.position);
    collisionArea.position.y += 0.8;
    collisionArea.userData = { isCollisionArea: true, isBasketCollider: true };

    scene.add(basketGroup);
    scene.add(collisionArea);

    basketRef.current = basketGroup;
    collisionAreaRef.current = collisionArea;

    return { basketGroup, collisionArea };
  };

  useEffect(() => {
    createBasket();

    return () => {
      if (basketRef.current && scene) {
        scene.remove(basketRef.current);
        basketRef.current = null;
      }

      if (collisionAreaRef.current && scene) {
        scene.remove(collisionAreaRef.current);
        collisionAreaRef.current = null;
      }
    };
  }, [scene]);

  // 테마 적용
  useEffect(() => {
    if (basketRef.current) {
      basketRef.current.traverse((child) => {
        if (child instanceof THREE.Mesh) {
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
      });
    }
  }, [isNight]);
  return null;
};

export default Basket;
