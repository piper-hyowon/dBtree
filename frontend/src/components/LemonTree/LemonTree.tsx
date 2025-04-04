// src/components/LemonTree/LemonTree.tsx

import React, { useRef, useEffect, useState } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls";
import "./LemonTree.css";

interface LemonTreeProps {
  onHarvest: (amount: number) => void;
  cooldownRemaining?: number;
  isCoolingDown?: boolean;
}

const LemonTree: React.FC<LemonTreeProps> = ({
  onHarvest,
  cooldownRemaining = 0,
  isCoolingDown = false,
}) => {
  const mountRef = useRef<HTMLDivElement>(null);
  const [isShaking, setIsShaking] = useState(false);
  const [showHarvestEffect, setShowHarvestEffect] = useState(false);
  const [harvestedAmount, setHarvestedAmount] = useState(0);

  useEffect(() => {
    if (!mountRef.current) return;

    // Scene setup
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xd8f3dc);

    // Camera setup
    const camera = new THREE.PerspectiveCamera(
      75,
      mountRef.current.clientWidth / mountRef.current.clientHeight,
      0.1,
      1000
    );
    camera.position.set(0, 2, 5);

    // Renderer setup
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(
      mountRef.current.clientWidth,
      mountRef.current.clientHeight
    );
    renderer.shadowMap.enabled = true;
    mountRef.current.appendChild(renderer.domElement);

    // Orbit controls
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.maxPolarAngle = Math.PI / 2;

    // Lighting
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.5);
    scene.add(ambientLight);

    const directionalLight = new THREE.DirectionalLight(0xffffff, 0.8);
    directionalLight.position.set(5, 10, 5);
    directionalLight.castShadow = true;
    scene.add(directionalLight);

    // Ground
    const groundGeometry = new THREE.CircleGeometry(3, 32);
    const groundMaterial = new THREE.MeshStandardMaterial({
      color: 0x90a955,
      roughness: 0.8,
      metalness: 0.2,
    });
    const ground = new THREE.Mesh(groundGeometry, groundMaterial);
    ground.rotation.x = -Math.PI / 2;
    ground.receiveShadow = true;
    scene.add(ground);

    // Tree trunk
    const trunkGeometry = new THREE.CylinderGeometry(0.2, 0.3, 1.5, 8);
    const trunkMaterial = new THREE.MeshStandardMaterial({
      color: 0x6a4c3b,
      roughness: 1.0,
      metalness: 0.0,
    });
    const trunk = new THREE.Mesh(trunkGeometry, trunkMaterial);
    trunk.position.y = 0.75;
    trunk.castShadow = true;
    scene.add(trunk);

    // Tree foliage (parent group)
    const foliageGroup = new THREE.Group();
    foliageGroup.position.y = 1.8;
    scene.add(foliageGroup);

    // Create multiple leaf clusters for a fuller tree
    const leafMaterial = new THREE.MeshStandardMaterial({
      color: 0x3a5a40,
      roughness: 0.8,
      metalness: 0.0,
    });

    const createLeafCluster = (
      x: number,
      y: number,
      z: number,
      scale: number
    ) => {
      const leafGeometry = new THREE.SphereGeometry(0.8, 10, 10);
      const leafCluster = new THREE.Mesh(leafGeometry, leafMaterial);
      leafCluster.position.set(x, y, z);
      leafCluster.scale.set(scale, scale, scale);
      leafCluster.castShadow = true;
      return leafCluster;
    };

    // Add several leaf clusters for a more natural look
    foliageGroup.add(createLeafCluster(0, 0, 0, 1));
    foliageGroup.add(createLeafCluster(0.6, -0.2, 0.5, 0.8));
    foliageGroup.add(createLeafCluster(-0.5, 0.1, 0.3, 0.9));
    foliageGroup.add(createLeafCluster(0.2, 0.4, -0.4, 0.85));
    foliageGroup.add(createLeafCluster(-0.3, -0.3, -0.6, 0.7));

    // Lemons
    const lemonMaterial = new THREE.MeshStandardMaterial({
      color: 0xffd700,
      roughness: 0.5,
      metalness: 0.1,
    });

    const lemons: THREE.Mesh[] = [];
    const lemonPositions = [
      [0.8, -0.3, 0.7],
      [-0.7, 0.2, 0.5],
      [0.5, 0.5, -0.2],
      [-0.4, -0.4, -0.6],
      [0.1, 0.7, 0.3],
      [-0.2, 0.0, 0.9],
    ];

    lemonPositions.forEach((pos) => {
      const lemonGeometry = new THREE.SphereGeometry(0.15, 16, 16);
      const lemon = new THREE.Mesh(lemonGeometry, lemonMaterial);
      lemon.position.set(pos[0], pos[1], pos[2]);
      lemon.castShadow = true;
      foliageGroup.add(lemon);
      lemons.push(lemon);
    });

    // Animation
    let initialFoliageRotation = foliageGroup.rotation.y;
    let shakingOffset = 0;

    // Handle window resize
    const handleResize = () => {
      if (!mountRef.current) return;

      camera.aspect =
        mountRef.current.clientWidth / mountRef.current.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(
        mountRef.current.clientWidth,
        mountRef.current.clientHeight
      );
    };

    window.addEventListener("resize", handleResize);

    // Animation loop
    const animate = () => {
      requestAnimationFrame(animate);

      // Gentle movement of the tree
      foliageGroup.rotation.y =
        initialFoliageRotation + Math.sin(Date.now() * 0.0005) * 0.05;

      // Tree shaking animation when harvesting
      if (isShaking) {
        shakingOffset += 0.2;
        foliageGroup.rotation.x = Math.sin(shakingOffset) * 0.1;
        foliageGroup.rotation.z = Math.cos(shakingOffset) * 0.1;

        // Bounce lemons slightly
        lemons.forEach((lemon, i) => {
          lemon.position.y += Math.sin(shakingOffset + i) * 0.01;
        });
      } else {
        // Slowly return to normal position
        foliageGroup.rotation.x *= 0.9;
        foliageGroup.rotation.z *= 0.9;
        shakingOffset = 0;
      }

      controls.update();
      renderer.render(scene, camera);
    };

    animate();

    // Cleanup function
    return () => {
      window.removeEventListener("resize", handleResize);
      mountRef.current?.removeChild(renderer.domElement);
      renderer.dispose();
    };
  }, [isShaking]);

  const handleHarvest = () => {
    if (isCoolingDown) return;

    setIsShaking(true);

    // Calculate harvested amount (5 is base amount from your Go code)
    const amount = 5;
    setHarvestedAmount(amount);

    // Show harvest effect
    setShowHarvestEffect(true);

    // Stop shaking after 1 second
    setTimeout(() => {
      setIsShaking(false);

      // Hide harvest effect after another second
      setTimeout(() => {
        setShowHarvestEffect(false);
        onHarvest(amount);
      }, 1000);
    }, 1000);
  };

  // Format time remaining
  const formatTimeRemaining = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (hours > 0) {
      return `${hours}시간 ${minutes}분`;
    } else {
      return `${minutes}분 ${seconds % 60}초`;
    }
  };

  return (
    <div className="lemon-tree-container">
      <div ref={mountRef} className="lemon-tree-canvas" />

      {showHarvestEffect && (
        <div className="harvest-effect">+{harvestedAmount} 레몬 수확!</div>
      )}

      <button
        onClick={handleHarvest}
        disabled={isCoolingDown}
        className={`harvest-button ${isCoolingDown ? "cooling-down" : ""}`}
      >
        {isCoolingDown
          ? `${formatTimeRemaining(cooldownRemaining)} 후 수확 가능`
          : "레몬 수확하기"}
      </button>
    </div>
  );
};

export default LemonTree;
