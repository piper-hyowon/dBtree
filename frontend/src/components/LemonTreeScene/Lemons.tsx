import React, { useCallback, useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { LEMONS } from "./constants/lemon.constant";
import { mockApi, QuizQuestion } from "../../services/mockApi";
import { useTheme } from "../../hooks/useTheme";
import { useLemonTreeScene } from "../../contexts/LemonTreeSceneContext";

export interface AvailableLemon {
  id: number;
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
}

interface LemonsProps {
  lemons: AvailableLemon[];
  setLemons: React.Dispatch<React.SetStateAction<AvailableLemon[]>>;
  lemonsLoaded: boolean;
  setLemonsLoaded: React.Dispatch<React.SetStateAction<boolean>>;
  addLemonToBasket: (id: number) => Promise<boolean>;
}

const Lemons: React.FC<LemonsProps> = ({
  lemons,
  setLemons,
  lemonsLoaded,
  setLemonsLoaded,
  addLemonToBasket,
}) => {
  const { scene, camera, renderer, controls } = useLemonTreeScene();

  const lemonModelRef = useRef<THREE.Group | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // 퀴즈 게임 상태
  const [quizQuestions, setQuizQuestions] = useState<QuizQuestion[]>([]);
  const [activeQuiz, setActiveQuiz] = useState<{
    question: QuizQuestion;
    lemonId: number;
  } | null>(null);

  const [currentTargetLemonId, setCurrentTargetLemonId] = useState<
    number | null
  >(null);

  const [showTarget, setShowTarget] = useState(false);
  const [loadingQuiz, setLoadingQuiz] = useState(false);
  const animationFrameRef = useRef<number | null>(null);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  // 서버에서 퀴즈 데이터 로드(10개)
  useEffect(() => {
    const loadQuizQuestions = async () => {
      try {
        const response = await mockApi.getQuizQuestions(10);
        if (response.success && response.data?.questions) {
          setQuizQuestions(response.data.questions);
        }
      } catch (error) {
        console.error("퀴즈 데이터 로드 오류:", error);
      }
    };

    loadQuizQuestions();
  }, []);

  const handleLemonClick = useCallback(
    async (lemonId: number) => {
      // 퀴즈는 동시에 1개만 활성화 가능
      if (activeQuiz || loadingQuiz) return;

      // orbit 컨트롤 비활성화
      if (controls) {
        controls.enabled = false;
      }

      try {
        setLoadingQuiz(true);

        let randomQuestion;
        if (quizQuestions.length > 0) {
          const randomIndex = Math.floor(Math.random() * quizQuestions.length);
          randomQuestion = quizQuestions[randomIndex];
        } else {
          const response = await mockApi.getQuizQuestions(1);
          if (!response.success || !response.data?.questions.length) {
            throw new Error("퀴즈를 가져오지 못했습니다");
          }
          randomQuestion = response.data.questions[0];
        }

        // 퀴즈 순서 랜덤
        const seed = (Date.now() % 1000) + Math.floor(Math.random() * 1000);
        const shuffledQuestion = { ...randomQuestion };
        const options = [...randomQuestion.options];
        const originalIndices = options.map((_, index) => index);

        // Fisher-Yates
        for (let i = options.length - 1; i > 0; i--) {
          const j = Math.floor((i + 1) * ((seed * (i + 1)) % 1) + 0.001);
          [options[i], options[j]] = [options[j], options[i]];
          [originalIndices[i], originalIndices[j]] = [
            originalIndices[j],
            originalIndices[i],
          ]; // 원본 인덱스도 함께 이동
        }

        // 셔플된 옵션 - 원본 인덱스 맵핑
        shuffledQuestion.options = options;
        shuffledQuestion.originalIndices = originalIndices;

        setActiveQuiz({
          question: shuffledQuestion,
          lemonId: lemonId,
        });
      } catch (error) {
        console.error("퀴즈 가져오기 오류:", error);
        alert("퀴즈를 가져오는데 오류가 발생했습니다.");
        if (controls) {
          controls.enabled = true;
        }
      } finally {
        setLoadingQuiz(false);
      }
    },
    [activeQuiz, controls, loadingQuiz, quizQuestions]
  );

  const handleQuizAnswer = useCallback(
    async (selectedIndex: number) => {
      if (!activeQuiz) return;

      const { question, lemonId } = activeQuiz;
      const originalSelectedIndex = question.originalIndices
        ? question.originalIndices[selectedIndex]
        : selectedIndex;

      const response = await mockApi.submitQuizAnswer(
        question.id,
        originalSelectedIndex
      );

      if (response.success && response.data?.correct) {
        setActiveQuiz(null); // 퀴즈 UI 닫기
        setCurrentTargetLemonId(lemonId);
        setShowTarget(true);

        timerRef.current = setTimeout(() => {
          console.log("타이머 완료: 시간 초과");
          setShowTarget(false);
          setCurrentTargetLemonId(null);
          if (controls) controls.enabled = true;
          alert("시간이 초과되었습니다! 다시 시도해주세요.");
        }, 5000);
      } else {
        setActiveQuiz(null);
        if (controls) controls.enabled = true;
        alert(
          `틀렸습니다! 정답은 "${
            question.options[question.correctOptionIndex]
          }" 입니다.`
        );
      }
    },
    [activeQuiz, controls]
  );

  const handleHTMLTargetClick = useCallback(() => {
    console.log("handleHTMLTargetClick");
    console.log(currentTargetLemonId);
    if (currentTargetLemonId === null || currentTargetLemonId === undefined) {
      console.error("타겟 클릭: 레몬 ID가 없음");
      return;
    }

    console.log(`타겟 클릭: 레몬 ID ${currentTargetLemonId} 처리 중`);

    setShowTarget(false); // 타겟 게임 종료

    // 타이머 취소
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }

    // 레몬 수확 처리
    addLemonToBasket(currentTargetLemonId)
      .then((success) => {
        console.log(
          `레몬 ID ${currentTargetLemonId} 수확 ${success ? "성공" : "실패"}`
        );
        setCurrentTargetLemonId(null);
      })
      .catch((err) => {
        console.error("레몬 수확 중 오류:", err);
        setCurrentTargetLemonId(null);
      });

    // orbit control 활성화
    if (controls) controls.enabled = true;
  }, [currentTargetLemonId, addLemonToBasket, controls]);

  // 클릭 이벤트 처리
  useEffect(() => {
    if (!scene || !renderer) return;

    const raycaster = new THREE.Raycaster();
    const mouse = new THREE.Vector2();

    const onMouseClick = (event: MouseEvent) => {
      // 타겟이 표시 중이면 Three.js 클릭 이벤트는 무시
      if (showTarget) return;

      const canvasBounds = renderer.domElement.getBoundingClientRect();
      mouse.x =
        ((event.clientX - canvasBounds.left) / canvasBounds.width) * 2 - 1;
      mouse.y =
        -((event.clientY - canvasBounds.top) / canvasBounds.height) * 2 + 1;

      raycaster.setFromCamera(mouse, camera);
      const intersects = raycaster.intersectObjects(scene.children, true);

      if (intersects.length > 0) {
        let hitObject = intersects[0].object;

        // 레몬 클릭 처리
        let currentObj: THREE.Object3D | null = hitObject;
        while (currentObj) {
          if (currentObj.userData.isLemon) {
            // 퀴즈 활성화 상태 아닐떄만 레몬클릭 처리!
            if (!activeQuiz) {
              handleLemonClick(currentObj.userData.lemonId);
            }
            break;
          }
          currentObj = currentObj.parent;
        }
      }
    };

    renderer.domElement.addEventListener("click", onMouseClick);

    return () => {
      renderer.domElement.removeEventListener("click", onMouseClick);
      if (animationFrameRef.current !== null) {
        cancelAnimationFrame(animationFrameRef.current);
        animationFrameRef.current = null;
      }
    };
  }, [scene, camera, renderer, handleLemonClick, showTarget, activeQuiz]);
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
              model.userData.isLemon = true;
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
                  child.userData.isLemon = true;
                  child.userData.lemonId = lemon.id;
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

      {/* 퀴즈 UI */}
      {activeQuiz && (
        <div className="quiz-container">
          <h3 className="quiz-title">DB 퀴즈</h3>
          <p className="quiz-question">{activeQuiz.question.question}</p>
          <div className="quiz-options">
            {activeQuiz.question.options.map((option, index) => (
              <button
                key={index}
                onClick={() => handleQuizAnswer(index)}
                className="quiz-option"
              >
                {option}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* 로딩 표시 */}
      {loadingQuiz && (
        <div className="quiz-loading">
          <p>퀴즈 로딩 중</p>
        </div>
      )}

      {/* HTML 타겟 */}
      {showTarget && (
        <div className="html-target" onClick={handleHTMLTargetClick} />
      )}
    </>
  );
};

export default Lemons;
