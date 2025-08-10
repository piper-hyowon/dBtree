import { useEffect, useRef } from "react";
import * as THREE from "three";
import { useTheme } from "../../hooks/useTheme";
import { useLemonTreeScene } from "../../contexts/LemonTreeSceneContext";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";

const Tree: React.FC = () => {
  const { scene } = useLemonTreeScene();
  const treeRef = useRef<THREE.Group | null>(null);
  const { isNight } = useTheme();

  const groundRef = useRef<THREE.Mesh | null>(null);

  const createTree = () => {
    if (!scene) return null;

    if (treeRef.current) {
      scene.remove(treeRef.current);
      treeRef.current = null;
    }

    const groundMaterial = new THREE.MeshStandardMaterial({
      color: "#dbf4d8",
      roughness: 1,
      side: THREE.FrontSide,
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

    loader.load(
      "/models/tree.gltf",
      (gltf) => {
        const model = gltf.scene;
        model.rotation.set(0, Math.PI, 0);
        model.position.set(0, 0, 0);
        scene.add(model);
        treeRef.current = model;
      },
      (progress) => {
        // console.log(
        //   `모델 로딩 진행률: ${Math.round(
        //     (progress.loaded / progress.total) * 100
        //   )}%`
        // );
      },
      (error) => {
        console.error("모델 로드 오류:", error);
      }
    );

    return null;
  };

  useEffect(() => {
    createTree();

    return () => {
      if (treeRef.current && scene) {
        scene.remove(treeRef.current);
        treeRef.current = null;
      }
    };
  }, [scene]);

  useEffect(() => {
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

    if (treeRef.current) {
      treeRef.current.traverse((child) => {
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
  }, [isNight]);

  return null;
};

export default Tree;
