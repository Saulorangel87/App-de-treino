'use client';

import { useEffect, useRef } from 'react';
import { prefersReducedMotion } from '@/lib/motion';
import { buildRider, WHEEL_R } from './ride-scene-rider';

// Cena 3D da tela de acesso: um ciclista pedalando sobre uma carta
// topográfica (curvas de nível, mata, água e a trilha vermelha). O Three.js só
// é baixado aqui, no navegador; sem WebGL fica a folha com relevo do CSS e,
// com movimento reduzido, a cena aparece parada.

const COLORS = {
  paper: '#eef1ea',
  forest: '#bdd1ad',
  contour: '#a98559',
  water: '#2f6f9e',
  trail: '#c3321f',
};

// Mesmo relevo no shader e no JavaScript (o ciclista segue o chão).
const HEIGHT_GLSL = `
float terrain(vec2 p) {
  return 0.55 * (0.9 * sin(p.x * 0.35 + 1.3) * cos(p.y * 0.28)
    + 0.5 * sin(p.x * 0.8 + p.y * 0.5)
    + 0.35 * cos(p.y * 0.9 - p.x * 0.3));
}
float pathZ(float x) { return 0.8 * sin(x * 0.12); }
`;

function terrain(x: number, z: number) {
  return (
    0.55 *
    (0.9 * Math.sin(x * 0.35 + 1.3) * Math.cos(z * 0.28) +
      0.5 * Math.sin(x * 0.8 + z * 0.5) +
      0.35 * Math.cos(z * 0.9 - x * 0.3))
  );
}

function pathZ(x: number) {
  return 0.8 * Math.sin(x * 0.12);
}

const SPEED = 1.5; // unidades por segundo
const GEAR = 2.6; // voltas da roda por volta do pedal
const TERRAIN_RADIUS = 9;

function hasWebGL() {
  try {
    const canvas = document.createElement('canvas');
    return Boolean(canvas.getContext('webgl2') || canvas.getContext('webgl'));
  } catch {
    return false;
  }
}

export function RideScene({ label }: { label: string }) {
  const hostRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host || !hasWebGL()) return;
    let disposed = false;
    let cleanup = () => {};

    import('three').then((THREE) => {
      if (disposed) return;
      const still = prefersReducedMotion();

      const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
      renderer.setClearColor(COLORS.paper);
      renderer.domElement.setAttribute('aria-hidden', 'true');
      host.appendChild(renderer.domElement);

      const scene = new THREE.Scene();
      const camera = new THREE.PerspectiveCamera(30, 1, 0.1, 60);

      scene.add(new THREE.HemisphereLight('#ffffff', COLORS.forest, 1.1));
      const sun = new THREE.DirectionalLight('#ffffff', 2.4);
      sun.position.set(3, 7, 6);
      scene.add(sun);

      // ---------- Terreno: carta topográfica em shader ----------
      const terrainGeometry = new THREE.PlaneGeometry(TERRAIN_RADIUS * 2, TERRAIN_RADIUS * 2, 160, 160);
      terrainGeometry.rotateX(-Math.PI / 2);
      const terrainMaterial = new THREE.ShaderMaterial({
        uniforms: {
          uOffset: { value: 0 },
          uRadius: { value: TERRAIN_RADIUS },
          uPaper: { value: new THREE.Color(COLORS.paper) },
          uForest: { value: new THREE.Color(COLORS.forest) },
          uContour: { value: new THREE.Color(COLORS.contour) },
          uWater: { value: new THREE.Color(COLORS.water) },
          uTrail: { value: new THREE.Color(COLORS.trail) },
        },
        vertexShader: `
          uniform float uOffset;
          varying float vH;
          varying vec2 vW;
          varying float vDist;
          ${HEIGHT_GLSL}
          void main() {
            vec3 pos = position;
            vec2 world = vec2(pos.x + uOffset, pos.z);
            pos.y = terrain(world);
            vH = pos.y;
            vW = world;
            vDist = length(pos.xz);
            gl_Position = projectionMatrix * modelViewMatrix * vec4(pos, 1.0);
          }
        `,
        fragmentShader: `
          uniform float uRadius;
          uniform vec3 uPaper;
          uniform vec3 uForest;
          uniform vec3 uContour;
          uniform vec3 uWater;
          uniform vec3 uTrail;
          varying float vH;
          varying vec2 vW;
          varying float vDist;
          ${HEIGHT_GLSL}
          float lineMask(float value, float width) {
            float d = abs(fract(value - 0.5) - 0.5);
            return 1.0 - smoothstep(0.0, fwidth(value) * width, d);
          }
          void main() {
            vec3 color = uPaper;
            float forest = smoothstep(0.15, 0.3, sin(vW.x * 0.21) * sin(vW.y * 0.33 + 1.0));
            color = mix(color, uForest, forest * 0.9);
            float water = smoothstep(-0.58, -0.66, vH);
            color = mix(color, mix(uPaper, uWater, 0.35), water);
            float k = vH * 7.0;
            color = mix(color, uContour, lineMask(k, 1.1) * 0.55);
            color = mix(color, uContour, lineMask(k / 5.0, 2.2) * 0.8);
            float road = abs(vW.y - pathZ(vW.x));
            float dash = step(0.45, fract(vW.x * 0.9));
            float trail = (1.0 - smoothstep(0.035, 0.035 + fwidth(road) * 1.5, road)) * dash;
            color = mix(color, uTrail, trail);
            float fade = smoothstep(uRadius * 0.55, uRadius * 0.95, vDist);
            gl_FragColor = vec4(mix(color, uPaper, fade), 1.0);
          }
        `,
      });
      scene.add(new THREE.Mesh(terrainGeometry, terrainMaterial));

      // ---------- Ciclista (components/ride-scene-rider.ts) ----------
      const rider = buildRider(THREE);
      const bike = rider.root;
      scene.add(bike);

      // ---------- Câmera, tamanho e quadro a quadro ----------
      let ready = false;
      const pointer = { x: 0, y: 0, tx: 0, ty: 0 };
      function onPointer(event: PointerEvent) {
        const rect = host!.getBoundingClientRect();
        pointer.tx = ((event.clientX - rect.left) / rect.width - 0.5) * 2;
        pointer.ty = ((event.clientY - rect.top) / rect.height - 0.5) * 2;
      }
      if (!still) window.addEventListener('pointermove', onPointer, { passive: true });

      function resize() {
        const width = host!.clientWidth || 1;
        const height = host!.clientHeight || 1;
        renderer.setSize(width, height, false);
        camera.aspect = width / height;
        camera.updateProjectionMatrix();
        // Parada (movimento reduzido), a cena só é redesenhada ao mudar de tamanho.
        if (still && ready) frame(0);
      }
      const resizeObserver = new ResizeObserver(resize);
      resizeObserver.observe(host);
      resize();

      let offset = 3;
      let wheelAngle = 0;
      let crankAngle = 0.6;
      let elapsed = 0;
      const lookAt = new THREE.Vector3();

      function frame(dt: number) {
        elapsed += dt;
        offset += SPEED * dt;
        wheelAngle += (SPEED / WHEEL_R) * dt;
        crankAngle += (SPEED / WHEEL_R / GEAR) * dt;
        terrainMaterial.uniforms.uOffset.value = offset;

        // O ciclista fica no x = 0 e o mapa corre por baixo dele.
        const z = pathZ(offset);
        const slopeZ = pathZ(offset + 0.5) - pathZ(offset - 0.5);
        const backY = terrain(offset - 0.5, pathZ(offset - 0.5));
        const frontY = terrain(offset + 0.5, pathZ(offset + 0.5));
        bike.position.set(0, (backY + frontY) / 2 + 0.01, z);
        // Inclinação limitada: o relevo é ilustrativo, não uma parede.
        const pitch = Math.max(-0.16, Math.min(0.16, Math.atan2(frontY - backY, 1)));
        bike.rotation.set(0, -Math.atan2(slopeZ, 1), pitch);
        rider.pose(crankAngle, wheelAngle);

        pointer.x += (pointer.tx - pointer.x) * 0.04;
        pointer.y += (pointer.ty - pointer.y) * 0.04;
        const narrow = camera.aspect < 1.2;
        const orbit = 0.55 + Math.sin(elapsed * 0.12) * 0.25 + pointer.x * 0.25;
        const distance = narrow ? 4.8 : 3.9;
        camera.position.set(
          Math.sin(orbit) * distance,
          bike.position.y + 1.5 - pointer.y * 0.3,
          z + Math.cos(orbit) * distance,
        );
        lookAt.set(narrow ? 0.4 : 0.1, bike.position.y + 0.7, z);
        camera.lookAt(lookAt);
        renderer.render(scene, camera);
      }

      let raf = 0;
      let last = performance.now();
      let visible = true;
      function loop(now: number) {
        const dt = Math.min(0.05, (now - last) / 1000);
        last = now;
        frame(dt);
        raf = requestAnimationFrame(loop);
      }
      function start() {
        if (still || raf || !visible || document.hidden) return;
        last = performance.now();
        raf = requestAnimationFrame(loop);
      }
      function stop() {
        cancelAnimationFrame(raf);
        raf = 0;
      }
      // Fora da tela ou com a aba escondida, a cena para de gastar bateria.
      const intersection = new IntersectionObserver(([entry]) => {
        visible = entry.isIntersecting;
        if (visible) start();
        else stop();
      });
      intersection.observe(host);
      const onVisibility = () => (document.hidden ? stop() : start());
      document.addEventListener('visibilitychange', onVisibility);

      ready = true;
      frame(0);
      host.classList.add('ready');
      start();

      cleanup = () => {
        stop();
        intersection.disconnect();
        resizeObserver.disconnect();
        document.removeEventListener('visibilitychange', onVisibility);
        window.removeEventListener('pointermove', onPointer);
        terrainGeometry.dispose();
        terrainMaterial.dispose();
        rider.dispose();
        renderer.dispose();
        renderer.domElement.remove();
      };
    }).catch(() => undefined);

    return () => {
      disposed = true;
      cleanup();
    };
  }, []);

  return <div ref={hostRef} className="ride-scene" role="img" aria-label={label} />;
}
