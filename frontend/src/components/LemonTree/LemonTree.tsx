import { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import "./LemonTree.css";
import { useTheme } from "../../hooks/useTheme";

interface LemonTreeProps {}

const LemonTree: React.FC<LemonTreeProps> = ({}) => {
  const mountRef = useRef<HTMLDivElement | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const { theme } = useTheme();
  const isNight = theme === "dark";

  useEffect(() => {
    if (!mountRef.current) return;
    const currentMount = mountRef.current;

    const scene = new THREE.Scene();
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

    const camera = new THREE.PerspectiveCamera(
      45,
      currentMount.clientWidth / currentMount.clientHeight,
      0.1,
      1000
    );
    camera.position.set(5, 5, 5);

    const renderer = new THREE.WebGLRenderer({
      antialias: true,
      powerPreference: "high-performance",
      alpha: false,
    });

    renderer.setSize(currentMount.clientWidth, currentMount.clientHeight);
    renderer.setPixelRatio(window.devicePixelRatio);
    renderer.outputColorSpace = THREE.SRGBColorSpace;

    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    currentMount.appendChild(renderer.domElement);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.maxPolarAngle = Math.PI / 2;
    controls.minDistance = 10 * 0.2;
    controls.maxDistance = 10 * 1.5;
    controls.zoomSpeed = 0.6;
    controls.enablePan = false;

    const ambientLight = new THREE.AmbientLight(0xffffff);
    const keyLight = new THREE.DirectionalLight(0xffffff);
    keyLight.position.set(1, 0, 1);
    keyLight.castShadow = true;
    keyLight.shadow.mapSize.width = 2048;
    keyLight.shadow.mapSize.height = 2048;
    keyLight.shadow.bias = -0.001;

    if (isNight) {
      ambientLight.intensity = 0.01;
      keyLight.intensity = 0.2;
    } else {
      ambientLight.intensity = 1.5;
      keyLight.intensity = 1.6;
    }
    scene.add(ambientLight);
    scene.add(keyLight);

    const ground = new THREE.Mesh(
      new THREE.BoxGeometry(3, 3, 0.2),
      isNight
        ? new THREE.MeshStandardMaterial({
            color: "#dbf4d8",
            roughness: 1,
            emissive: "#dbf4d8",
            emissiveIntensity: 0.4,
          })
        : new THREE.MeshStandardMaterial({ color: "#dbf4d8", roughness: 1 })
    );
    ground.rotation.x = -Math.PI / 2;
    ground.position.y = -1.45;
    ground.receiveShadow = true;
    scene.add(ground);

    const loader = new GLTFLoader();
    loader.load(
      "/models/greentree.gltf",
      (gltf) => {
        const model = gltf.scene;

        model.traverse((child) => {
          if (child instanceof THREE.Mesh) {
            if (!child.name.includes("body")) {
              console.log(child.name);

              const material = child.material as THREE.MeshStandardMaterial;
              material.emissive = new THREE.Color(material.color);
              material.emissiveIntensity = isNight ? 0.8 : 0;

              material.needsUpdate = true;
            }
          }
        });

        scene.add(model);
        model.rotation.y = Math.PI;
        camera.lookAt(model.position);

        setIsLoading(false);
      },
      undefined,
      (error) => {
        console.error("모델 로드 오류:", error);
        setIsLoading(false);
      }
    );

    function animate() {
      requestAnimationFrame(animate);
      controls.update();
      renderer.render(scene, camera);
    }
    animate();

    // 창 크기 변경 처리
    function handleResize() {
      if (!currentMount) return;
      camera.aspect = currentMount.clientWidth / currentMount.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(currentMount.clientWidth, currentMount.clientHeight);
    }
    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      renderer.dispose();
      if (currentMount?.contains(renderer.domElement)) {
        currentMount.removeChild(renderer.domElement);
      }
    };
  }, [isNight]);

  return (
    <div className="gltf-container">
      <div ref={mountRef} className="gltf-mount" />

      {isLoading && <div className="gltf-loading">모델 로딩 중...</div>}
    </div>
  );
};

export default LemonTree;
