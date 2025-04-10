import React, { useCallback, useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { AvailableLemon } from "../LemonTree";
import { LEMONS } from "../LemonTree/constants/lemon.constant";
import { mockApi } from "../../services/mockApi";
import { useTheme } from "../../hooks/useTheme";
import { useLemonTreeScene } from "../../contexts/LemonTreeSceneContext";

interface LemonsProps {
  lemons: AvailableLemon[];
  setLemons: React.Dispatch<React.SetStateAction<AvailableLemon[]>>;
  lemonsLoaded: boolean;
  setLemonsLoaded: React.Dispatch<React.SetStateAction<boolean>>;
}

const Lemons: React.FC<LemonsProps> = ({
  lemons,
  setLemons,
  lemonsLoaded,
  setLemonsLoaded,
}) => {
  const { scene} = useLemonTreeScene();

  const lemonModelRef = useRef<THREE.Group | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const { isNight } = useTheme();

  useEffect(() => {
    fetchLemonData();

    return () => {
      if (lemonModelRef.current && scene) {
        scene.remove(lemonModelRef.current);
        lemonModelRef.current = null;
      }
    };
  }, [scene]);

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

              model.traverse((child) => {
                if (child instanceof THREE.Mesh) {
                  if (Array.isArray(child.material)) {
                    child.material.forEach((mat) => {
                      if (mat instanceof THREE.MeshStandardMaterial) {
                        mat.userData.originalColor = mat.color.clone();
               
                        if (isNight) {
                          mat.emissive.copy(mat.userData.originalColor);
                          mat.emissiveIntensity = 0.8;
                        } else {
                          mat.emissive.set(0, 0, 0);
                          mat.emissiveIntensity = 0;
                        }
                        mat.needsUpdate = true;
                      }
                    });
                  } else if (
                    child.material instanceof THREE.MeshStandardMaterial
                  ) {
                    child.material.userData.originalColor =
                      child.material.color.clone();
                

                    if (isNight) {
                      child.material.emissive.copy(
                        child.material.userData.originalColor
                      );
                      child.material.emissiveIntensity = 0.8;
                    } else {
                      child.material.emissive.set(0, 0, 0);
                      child.material.emissiveIntensity = 0;
                    }
                    child.material.needsUpdate = true;
                  }
                }
              });

              scene.add(model);
            },
            undefined,
            (error) => console.error(`레몬 ${lemon.id} 로드 오류:`, error)
          );
        });
      }
      setLemonsLoaded(true); 
    } catch (err) {
      console.error("레몬 데이터 로드 오류:", err);
    }
  }, [lemonsLoaded]);

  const updateLemonMaterials = useCallback(
    (nightMode: boolean) => {
      if (!scene) return;

      scene.traverse((object) => {
        if (object.name && object.name.startsWith("lemon-")) {
          object.traverse((child) => {
            if (child instanceof THREE.Mesh) {

              const processMaterial = (material: THREE.Material) => {
                if (material instanceof THREE.MeshStandardMaterial) {
                  if (!material.userData.originalColor) {
                    material.userData.originalColor = material.color.clone();
               
                  }

                  if (nightMode) {
                    material.emissive.copy(material.userData.originalColor);
                    material.emissiveIntensity = 0.8;
             
                  } else {
                    material.emissive.set(0, 0, 0);
                    material.emissiveIntensity = 0;
                  }
                  material.needsUpdate = true;
                }
              };

              if (Array.isArray(child.material)) {
                child.material.forEach(processMaterial);
              } else if (child.material) {
                processMaterial(child.material);
              }
            }
          });
        }
      });
    },
    [scene]
  );

  useEffect(() => {
    updateLemonMaterials(isNight);
  }, [isNight, updateLemonMaterials]);

  useEffect(() => {
    if (lemonsLoaded && scene) {
      updateLemonMaterials(isNight);
    }
  }, [lemonsLoaded, updateLemonMaterials, isNight, scene]);
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
