import React, {
    createContext,
    useContext,
    useRef,
    useState,
    useEffect,
} from "react";
import * as THREE from "three";
import {OrbitControls} from "three/examples/jsm/controls/OrbitControls";
import {useTheme} from "../hooks/useTheme";

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
    lemons: AvailableLemon[];
}

const LemonTreeSceneContext = createContext<
    LemonTreeSceneContextType | undefined
>(undefined);

export const LemonTreeSceneProvider: React.FC<{
    children: React.ReactNode;
}> = ({children}) => {
    const {isNight} = useTheme();
    const containerRef = useRef<HTMLDivElement>(null);

    const [isInitialized, setIsInitialized] = useState(false);
    const [lemons, setLemons] = useState<AvailableLemon[]>([]);

    const sceneRef = useRef<THREE.Scene>(new THREE.Scene());
    const cameraRef = useRef<THREE.PerspectiveCamera>(
        new THREE.PerspectiveCamera(45, 1, 0.1, 1000)
    );
    const rendererRef = useRef<THREE.WebGLRenderer | null>(null);
    const controlsRef = useRef<OrbitControls | null>(null);

    const animationFrameIdRef = useRef<number | null>(null);

    const ambientLightRef = useRef<THREE.AmbientLight>(
        new THREE.AmbientLight(0xffffff)
    );
    const keyLightRef = useRef<THREE.DirectionalLight>(
        new THREE.DirectionalLight(0xffffff)
    );

    // 씬 초기화
    useEffect(() => {
        if (isInitialized) return;
        if (!containerRef.current || isInitialized) return;

        const container = containerRef.current;
        // console.log("컨테이너 확인:", container);

        // console.log("Scene initialization started, container size:", {
        //     width: container.clientWidth,
        //     height: container.clientHeight,
        // });

        while (container.firstChild) {
            container.removeChild(container.firstChild);
        }

        const scene = sceneRef.current;
        scene.background = null;
        const camera = cameraRef.current;
        camera.position.set(0, 3, 7);
        camera.aspect = container.clientWidth / container.clientHeight;
        camera.updateProjectionMatrix();

        try {
            const renderer = new THREE.WebGLRenderer({
                antialias: true,
                powerPreference: "high-performance",
                alpha: true,
                premultipliedAlpha: true,
            });
            renderer.setSize(container.clientWidth, container.clientHeight);
            renderer.setPixelRatio(window.devicePixelRatio);
            renderer.outputColorSpace = THREE.SRGBColorSpace;
            renderer.shadowMap.enabled = true;
            renderer.shadowMap.type = THREE.PCFSoftShadowMap;

            container.appendChild(renderer.domElement);
            rendererRef.current = renderer;

            // console.log("렌더러 생성 및 DOM에 추가 완료:", renderer.domElement);

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
                window.removeEventListener("resize", handleResize);

                if (animationFrameIdRef.current !== null) {
                    cancelAnimationFrame(animationFrameIdRef.current);
                    animationFrameIdRef.current = null;
                }

                if (renderer && container.contains(renderer.domElement)) {
                    container.removeChild(renderer.domElement);
                }

                if (rendererRef.current) {
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

        scene.background = null

        if (isNight) {
            ambientLight.intensity = 0.01;
            keyLight.intensity = 0.2;
        } else {
            ambientLight.intensity = 1.5;
            keyLight.intensity = 1.6;
        }
    }, [isNight, isInitialized]);

    const contextValue: LemonTreeSceneContextType = {
        scene: sceneRef.current,
        camera: cameraRef.current,
        renderer: rendererRef.current,
        controls: controlsRef.current,
        containerRef,
        lemons,
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
