import { useEffect, useRef, useCallback } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import { useTheme } from "../../../hooks/useTheme";


// TODO: 최적화
interface LemonProps {
  scene: THREE.Scene | null;
  camera: THREE.Camera | null;
  renderer: THREE.WebGLRenderer | null;
  orbitControls: OrbitControls | null;
  id: string;
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
  onDragStart?: (id: string) => void;
  onDragEnd?: (id: string, position: THREE.Vector3) => void;
}

const globalLemonMap = new Map<string, THREE.Group>();

const lemonPositions = new Map<string, { x: number; y: number; z: number }>();

const Lemon: React.FC<LemonProps> = ({
  scene,
  camera,
  renderer,
  orbitControls,
  id,
  position,
  rotation,
  onDragStart,
  onDragEnd,
}) => {
  const { theme } = useTheme();

  const lemonGroupRef = useRef<THREE.Group | null>(null);
  const mountedRef = useRef(false);
  const loadingRef = useRef(false);
  const isDraggingRef = useRef(false);
  const instanceIdRef = useRef(`lemon-${id}-${Date.now()}`);

  const currentPositionRef = useRef(
    lemonPositions.has(id) ? lemonPositions.get(id)! : position
  );

  const eventHandlersRef = useRef({
    mouseDown: null as ((e: MouseEvent) => void) | null,
    mouseMove: null as ((e: MouseEvent) => void) | null,
    mouseUp: null as ((e: MouseEvent) => void) | null,
  });

  const removeEventListeners = useCallback(() => {
    if (!renderer) return;

    const handlers = eventHandlersRef.current;

    if (handlers.mouseDown) {
      renderer.domElement.removeEventListener("mousedown", handlers.mouseDown);
    }

    if (handlers.mouseMove) {
      window.removeEventListener("mousemove", handlers.mouseMove);
    }

    if (handlers.mouseUp) {
      window.removeEventListener("mouseup", handlers.mouseUp);
    }

    eventHandlersRef.current = {
      mouseDown: null,
      mouseMove: null,
      mouseUp: null,
    };
  }, [renderer]);

  const removeLemon = useCallback(() => {
    if (lemonGroupRef.current && lemonGroupRef.current.parent) {
      if (lemonGroupRef.current && !isDraggingRef.current) {
        const pos = lemonGroupRef.current.position;
        lemonPositions.set(id, { x: pos.x, y: pos.y, z: pos.z });
      }

      lemonGroupRef.current.parent.remove(lemonGroupRef.current);

      lemonGroupRef.current.traverse((child) => {
        if (child instanceof THREE.Mesh) {
          if (child.geometry) child.geometry.dispose();
          if (child.material) {
            if (Array.isArray(child.material)) {
              child.material.forEach((m) => m.dispose());
            } else {
              child.material.dispose();
            }
          }
        }
      });

      const mapKey = `lemon-${id}`;
      if (globalLemonMap.get(mapKey) === lemonGroupRef.current) {
        globalLemonMap.delete(mapKey);
      }

      lemonGroupRef.current = null;
    }

    removeEventListeners();

    if (renderer && scene && camera) {
      renderer.render(scene, camera);
    }
  }, [id, scene, camera, renderer, removeEventListeners]);

  const setupDragEvents = useCallback(
    (group: THREE.Group) => {
      if (!group || !camera || !renderer || !scene) return;

      removeEventListeners();

      const raycaster = new THREE.Raycaster();
      const mouse = new THREE.Vector2();
      const dragPlane = new THREE.Plane();
      const dragOffset = new THREE.Vector3();

      const handleMouseDown = (event: MouseEvent) => {
        if (!group) return;

        const rect = renderer.domElement.getBoundingClientRect();
        mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
        mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

        raycaster.setFromCamera(mouse, camera);
        const intersects = raycaster.intersectObject(group, true);

        if (intersects.length > 0) {
          isDraggingRef.current = true;

          if (orbitControls) {
            orbitControls.enabled = false;
          }

          renderer.domElement.style.cursor = "grabbing";
          onDragStart?.(id);

          const planeNormal = new THREE.Vector3();
          camera.getWorldDirection(planeNormal);
          planeNormal.negate(); // 카메라 방향의 반대 방향

          dragPlane.setFromNormalAndCoplanarPoint(planeNormal, group.position);

          const intersectionPoint = new THREE.Vector3();
          if (raycaster.ray.intersectPlane(dragPlane, intersectionPoint)) {
            dragOffset.copy(group.position).sub(intersectionPoint);
          }
        }
      };

      const handleMouseMove = (event: MouseEvent) => {
        if (!isDraggingRef.current || !group) return;

        const rect = renderer.domElement.getBoundingClientRect();
        mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
        mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

        raycaster.setFromCamera(mouse, camera);
        const intersectionPoint = new THREE.Vector3();

        if (raycaster.ray.intersectPlane(dragPlane, intersectionPoint)) {
          const newPosition = intersectionPoint.clone().add(dragOffset);
          group.position.copy(newPosition);

          renderer.render(scene, camera);
        }
      };

      const handleMouseUp = () => {
        if (!isDraggingRef.current || !group) return;

        isDraggingRef.current = false;

        if (orbitControls) {
          orbitControls.enabled = true;
        }

        renderer.domElement.style.cursor = "auto";

        const pos = group.position;
        currentPositionRef.current = { x: pos.x, y: pos.y, z: pos.z };

        lemonPositions.set(id, currentPositionRef.current);

        onDragEnd?.(id, group.position.clone());

        renderer.render(scene, camera);
      };

      renderer.domElement.addEventListener("mousedown", handleMouseDown);
      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);

      eventHandlersRef.current = {
        mouseDown: handleMouseDown,
        mouseMove: handleMouseMove,
        mouseUp: handleMouseUp,
      };
    },
    [
      camera,
      renderer,
      scene,
      id,
      onDragStart,
      onDragEnd,
      orbitControls,
      removeEventListeners,
    ]
  );

  const createLemon = useCallback(() => {
    if (!scene || !camera || !renderer || loadingRef.current) return;

    loadingRef.current = true;

    const mapKey = `lemon-${id}`;
    if (globalLemonMap.has(mapKey)) {
      const existingLemon = globalLemonMap.get(mapKey);
      if (existingLemon && existingLemon.parent) {
        existingLemon.parent.remove(existingLemon);
      }
      globalLemonMap.delete(mapKey);
    }
    const group = new THREE.Group();
    group.name = mapKey;
    group.userData = { lemonId: id, instanceId: instanceIdRef.current };

    const posToUse = lemonPositions.get(id) || position;
    currentPositionRef.current = posToUse;

    group.position.set(posToUse.x, posToUse.y, posToUse.z);
    group.rotation.set(rotation.x, rotation.y, rotation.z);

    const loader = new GLTFLoader();
    loader.load(
      "/models/lemon.gltf",
      (gltf) => {
        if (!mountedRef.current) {
          loadingRef.current = false;
          return;
        }

        while (gltf.scene.children.length > 0) {
          const child = gltf.scene.children[0];

          child.userData = { lemonId: id, instanceId: instanceIdRef.current };

          if (child instanceof THREE.Mesh) {
            child.castShadow = true;
            child.receiveShadow = true;

            if (theme === "dark" && child.material) {
              const material = child.material as THREE.MeshStandardMaterial;
              material.emissive = new THREE.Color(material.color);
              material.emissiveIntensity = 0.8;
            }
          }

          group.add(child);
        }

        if (!mountedRef.current) {
          loadingRef.current = false;
          return;
        }

        scene.add(group);
        globalLemonMap.set(mapKey, group);
        lemonGroupRef.current = group;
        setupDragEvents(group);
        renderer.render(scene, camera);

        loadingRef.current = false;
      },
      undefined,
      (error) => {
        console.error("레몬 모델 로드 오류:", error);
        loadingRef.current = false;
      }
    );
  }, [scene, camera, renderer, id, position, rotation, theme, setupDragEvents]);

  useEffect(() => {
    mountedRef.current = true;

    if (scene && camera && renderer) {
      createLemon();
    }

    return () => {
      mountedRef.current = false;
      removeLemon();
    };
  }, [scene, camera, renderer, createLemon, removeLemon]);

  useEffect(() => {
    if (lemonGroupRef.current && !isDraggingRef.current) {
      if (!lemonPositions.has(id)) {
        lemonGroupRef.current.position.set(position.x, position.y, position.z);
        currentPositionRef.current = position;

        if (renderer && camera && scene) {
          renderer.render(scene, camera);
        }
      }
    }
  }, [position, renderer, camera, scene, id]);

  useEffect(() => {
    if (lemonGroupRef.current) {
      lemonGroupRef.current.rotation.set(rotation.x, rotation.y, rotation.z);

      if (renderer && camera && scene) {
        renderer.render(scene, camera);
      }
    }
  }, [rotation, renderer, camera, scene]);

  useEffect(() => {
    if (!lemonGroupRef.current) return;

    const currentPos = lemonGroupRef.current.position;

    lemonGroupRef.current.traverse((child) => {
      if (child instanceof THREE.Mesh && child.material) {
        const material = child.material as THREE.MeshStandardMaterial;

        if (theme === "dark") {
          material.emissive = new THREE.Color(material.color);
          material.emissiveIntensity = 0.8;
        } else {
          material.emissive = new THREE.Color(0x000000);
          material.emissiveIntensity = 0;
        }
      }
    });

    if (lemonPositions.has(id)) {
      const pos = lemonPositions.get(id)!;
      lemonGroupRef.current.position.set(pos.x, pos.y, pos.z);
    } else {
      lemonGroupRef.current.position.copy(currentPos);
    }

    if (renderer && camera && scene) {
      renderer.render(scene, camera);
    }
  }, [theme, renderer, camera, scene, id]);

  return null;
};

export default Lemon;
