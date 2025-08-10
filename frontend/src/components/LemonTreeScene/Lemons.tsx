import React, {useCallback, useEffect, useRef, useState} from "react";
import * as THREE from "three";
import {GLTFLoader} from "three/examples/jsm/loaders/GLTFLoader";
import {LEMONS} from "./constants/lemon.constant";
import {DEMO_QUIZ} from "../../services/mockApi";
import {useTheme} from "../../hooks/useTheme";
import {useLemonTreeScene} from "../../contexts/LemonTreeSceneContext";
import api from "../../services/api";
import {useAuth} from "../../contexts/AuthContext";

export interface AvailableLemon {
    id: number;
    position: { x: number; y: number; z: number };
    rotation: { x: number; y: number; z: number };
}

interface LemonsProps {
    setLemons: React.Dispatch<React.SetStateAction<AvailableLemon[]>>;
    lemonsLoaded: boolean;
    setLemonsLoaded: React.Dispatch<React.SetStateAction<boolean>>;
    setAvailableLemonCount: React.Dispatch<React.SetStateAction<number>>;
    setNextGrowthTime: React.Dispatch<React.SetStateAction<string | null>>;
}

const Lemons: React.FC<LemonsProps> = ({
                                           setLemons,
                                           lemonsLoaded,
                                           setLemonsLoaded,
                                           setAvailableLemonCount,
                                           setNextGrowthTime,
                                       }) => {
    const {isLoggedIn} = useAuth();

    const {scene, camera, renderer, controls} = useLemonTreeScene();

    const lemonModelRef = useRef<THREE.Group | null>(null);
    const [isLoading, setIsLoading] = useState(false);

    // 퀴즈 게임 상태
    const [activeQuiz, setActiveQuiz] = useState<{
        question: string;
        options: string[];
        lemonId: number;
        attemptID: number;
    } | null>(null);

    const [currentTargetLemonId, setCurrentTargetLemonId] = useState<
        number | null
    >(null);
    const [currentAttemptId, setCurrentAttemptId] = useState<number | null>(null);

    const [canHarvestStatus, setCanHarvestStatus] = useState<{
        canHarvest: boolean;
        waitSeconds: number;
    } | null>(null);

    const [showTarget, setShowTarget] = useState(false);
    const [loadingQuiz, setLoadingQuiz] = useState(false);
    const animationFrameRef = useRef<number | null>(null);
    const timerRef = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        const checkHarvestAvailability = async () => {
            if (!isLoggedIn) {
                // 비로그인 상태는 항상 수확 가능
                setCanHarvestStatus({canHarvest: true, waitSeconds: 0});
                return;
            }

            try {
                const status = await api.quiz.canHarvest();
                setCanHarvestStatus(status);
                console.log("status: ", status)

                if (!status.canHarvest && status.waitSeconds) {
                    console.log(`${status.waitSeconds} 초 후 가능`);
                }
            } catch (error) {
                console.error("수확 가능 여부 체크 실패:", error);
            }
        };

        checkHarvestAvailability();
    }, []);

    const handleLemonClick = async (lemonId: number) => {
        if (activeQuiz || loadingQuiz) return;

        // 쿨다운 체크
        if (isLoggedIn && canHarvestStatus && !canHarvestStatus.canHarvest) {
            alert(`아직 수확할 수 없습니다. ${canHarvestStatus.waitSeconds}초 후 가능`);
            return;
        }

        if (controls) {
            controls.enabled = false;
        }

        setLoadingQuiz(true);

        try {
            if (isLoggedIn) {
                const response = await api.quiz.getQuizQuestions(lemonId);
                setActiveQuiz({
                    question: response.question,
                    options: response.options,
                    lemonId: lemonId,
                    attemptID: response.attemptID,
                });
            } else {
                await new Promise(resolve => setTimeout(resolve, 500));
                setActiveQuiz({
                    question: DEMO_QUIZ.question,
                    options: DEMO_QUIZ.options,
                    lemonId: lemonId,
                    attemptID: DEMO_QUIZ.attemptID,
                });
            }
        } catch (error) {
            console.error("퀴즈 가져오기 오류:", error);
            alert("퀴즈를 가져오는데 오류가 발생했습니다.");
            if (controls) {
                controls.enabled = true;
            }
        } finally {
            setLoadingQuiz(false);
        }
    };

    const handleQuizAnswer = async (selectedIndex: number) => {
        if (!activeQuiz) return;

        try {
            let isCorrect = false;
            let correctOption = 0;

            if (isLoggedIn) {
                const response = await api.quiz.submitQuizAnswer(
                    selectedIndex,
                    activeQuiz.attemptID
                );
                isCorrect = response.isCorrect;
                correctOption = response.correctOption;

                setActiveQuiz(null);

                if (isCorrect && response.harvestEnabled) {
                    setCurrentTargetLemonId(activeQuiz.lemonId);
                    setCurrentAttemptId(activeQuiz.attemptID);
                    setShowTarget(true);

                    const timeoutMs = new Date(response.harvestTimeoutAt).getTime() - Date.now();
                    timerRef.current = setTimeout(() => {
                        setShowTarget(false);
                        setCurrentTargetLemonId(null);
                        setCurrentAttemptId(null);
                        if (controls) controls.enabled = true;
                        alert("시간이 초과되었습니다!");
                    }, timeoutMs);
                } else {
                    if (controls) controls.enabled = true;
                    alert(`틀렸습니다! 정답은 "${response.correctOption + 1}번" 입니다.`);
                }
            } else {
                // 데모 모드
                isCorrect = selectedIndex === DEMO_QUIZ.correctIndex;
                correctOption = DEMO_QUIZ.correctIndex;

                setActiveQuiz(null);

                if (isCorrect) {
                    setCurrentTargetLemonId(activeQuiz.lemonId);
                    setCurrentAttemptId(DEMO_QUIZ.attemptID);
                    setShowTarget(true);

                    // 데모는 5초 타임아웃
                    timerRef.current = setTimeout(() => {
                        setShowTarget(false);
                        setCurrentTargetLemonId(null);
                        setCurrentAttemptId(null);
                        if (controls) controls.enabled = true;
                        alert("시간이 초과되었습니다!");
                    }, 5000);
                } else {
                    if (controls) controls.enabled = true;
                    alert(`틀렸습니다!`);
                }
            }
        } catch (error) {
            console.error("답변 제출 오류:", error);
            alert("답변 제출 중 오류가 발생했습니다.");
            setActiveQuiz(null);
            if (controls) controls.enabled = true;
        }
    };

    const handleHTMLTargetClick = useCallback(async () => {
        if (currentTargetLemonId === null || currentAttemptId === null) {
            console.error("타겟 클릭: 필요한 정보가 없음");
            return;
        }

        setShowTarget(false);

        if (timerRef.current) {
            clearTimeout(timerRef.current);
            timerRef.current = null;
        }

        if (isLoggedIn) {
            try {
                const response = await api.quiz.harvestLemon(
                    currentTargetLemonId,
                    currentAttemptId
                );

                if (response) {
                    alert(`축하합니다! ${response.harvestAmount} 크레딧을 획득했습니다!\n현재 잔액: ${response.newBalance}`);

                    const newStatus = await api.quiz.canHarvest();
                    setCanHarvestStatus(newStatus);
                }
            } catch (error) {
                console.error("수확 오류:", error);
                alert("수확 중 오류가 발생했습니다.");
            }
        } else {
            alert("🎉 수확 성공! \n\n로그인하면 실제로 크레딧을 얻을 수 있습니다.\n지금 로그인하시겠습니까?");
            // TODO: 확인 시 로그인 페이지로 이동?
        }

        setCurrentTargetLemonId(null);
        setCurrentAttemptId(null);
        if (controls) controls.enabled = true;
    }, [currentTargetLemonId, currentAttemptId, controls, isLoggedIn]);

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
    const {isNight} = useTheme();

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
            setIsLoading(true); // 로딩 시작
            const response = await api.home.getLemonTreeStatus();
            console.log("API 응답:", response);

            // state 업데이트
            const lemonCount = response?.availablePositions?.length ?? 0;
            setAvailableLemonCount(lemonCount);
            setNextGrowthTime(response?.nextRegrowthTime ?? response?.nextRegrowthTime ?? null);

            if (response?.availablePositions?.length) {
                const lemonData: AvailableLemon[] = response.availablePositions.map((e) => ({
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
            setIsLoading(false); // 로딩 완료
        } catch (err) {
            console.error("레몬 데이터 로드 오류:", err);
            setIsLoading(false); // 에러 시에도 로딩 종료
        }
    }, [lemonsLoaded, setAvailableLemonCount, setNextGrowthTime, setLemons, setLemonsLoaded, scene, isNight]);

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
                    <p className="quiz-question">{activeQuiz.question}</p>
                    <div className="quiz-options">
                        {activeQuiz.options.map((option, index) => (
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
                <div className="html-target" onClick={handleHTMLTargetClick}/>
            )}

            {isLoggedIn && canHarvestStatus && !canHarvestStatus.canHarvest && (
                <div style={{
                    position: "absolute",
                    top: "50px",
                    left: "50%",
                    transform: "translateX(-50%)",
                    background: "rgba(255, 100, 100, 0.9)",
                    color: "white",
                    padding: "10px 20px",
                    borderRadius: "8px",
                    zIndex: 500,
                }}>
                    {Math.floor(canHarvestStatus.waitSeconds / 60)}분 {canHarvestStatus.waitSeconds % 60}초 후 수확 가능
                </div>
            )}
        </>
    );
};

export default Lemons;