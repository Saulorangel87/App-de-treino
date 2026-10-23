'use client';

import { useEffect, useRef } from 'react';
import { prefersReducedMotion } from '@/lib/motion';
import { buildRider, WHEEL_R } from './ride-scene-rider';

// Cena 3D da tela de acesso: um ciclista do Cadência numa estrada sobre relevo
// real, com as curvas de nível da carta topográfica sobrepostas ao terreno e a
// faixa central vermelha da trilha. Céu físico com nuvens, luz do sol coerente
// com ele, sombras, oclusão de ambiente (GTAO), capim instanciado e, no
// computador, profundidade de campo de lente. O Three.js só é baixado aqui, no
// navegador; sem WebGL fica a folha com relevo do CSS e, com movimento reduzido,
// a cena aparece parada.

// Direção do sol (radianos): azimute a partir de +z rumo a +x e elevação.
const SUN_AZIMUTH = 1.15;
const SUN_ELEVATION = 0.42;
const HAZE = '#bccbd2';

// Relevo natural.
function hills(x: number, z: number) {
  return (
    0.55 *
      (0.9 * Math.sin(x * 0.35 + 1.3) * Math.cos(z * 0.28) +
        0.5 * Math.sin(x * 0.8 + z * 0.5) +
        0.35 * Math.cos(z * 0.9 - x * 0.3)) +
    0.04 * Math.sin(x * 2.3 + z * 1.7) * Math.cos(z * 2.9 - x * 0.7)
  );
}

function pathZ(x: number) {
  return 0.8 * Math.sin(x * 0.12);
}

// Greide da estrada: só ondulações longas, para as duas rodas tocarem o asfalto
// juntas. É a única altura sob a bicicleta.
function roadY(x: number) {
  return 0.3 * Math.sin(x * 0.21 + 0.7) + 0.16 * Math.sin(x * 0.083 + 2.1) - 0.05;
}

const ROAD_HALF = 0.3;

function smoothstep(a: number, b: number, t: number) {
  const k = Math.min(1, Math.max(0, (t - a) / (b - a)));
  return k * k * (3 - 2 * k);
}

// O relevo se acomoda ao greide perto da estrada (aterro e corte suaves) e,
// na faixa da estrada, fica plano na transversal.
function ground(x: number, z: number) {
  const d = Math.abs(z - pathZ(x));
  const fit = Math.exp(-(d * d) / 2.6);
  const natural = hills(x, z) + (roadY(x) - hills(x, pathZ(x))) * fit;
  const road = 1 - smoothstep(ROAD_HALF + 0.05, ROAD_HALF + 0.6, d);
  return natural + (roadY(x) - natural) * road;
}

const SPEED = 1.6; // unidades por segundo
const GEAR = 2.6; // voltas da roda por volta do pedal
const REAR_X = -0.5; // contato das rodas com o chão, no eixo da bicicleta
const FRONT_X = 0.53;
const WIDTH = 26;
const DEPTH = 18;
// Faixa de capim ao redor do ciclista (o resto some na névoa).
const GRASS_SPAN = 14;
const GRASS_DEPTH = 9;

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

    Promise.all([
      import('three'),
      import('three/examples/jsm/objects/Sky.js'),
      import('three/examples/jsm/postprocessing/EffectComposer.js'),
      import('three/examples/jsm/postprocessing/RenderPass.js'),
      import('three/examples/jsm/postprocessing/GTAOPass.js'),
      import('three/examples/jsm/postprocessing/BokehPass.js'),
      import('three/examples/jsm/postprocessing/OutputPass.js'),
    ])
      .then(([THREE, { Sky }, { EffectComposer }, { RenderPass }, { GTAOPass }, { BokehPass }, { OutputPass }]) => {
        if (disposed) return;
        const still = prefersReducedMotion();
        const mobile = window.matchMedia('(max-width: 959px)').matches;
        const random = (() => {
          let seed = 7;
          return () => {
            seed = (seed * 16807) % 2147483647;
            return (seed - 1) / 2147483646;
          };
        })();

        const renderer = new THREE.WebGLRenderer({ antialias: false, powerPreference: 'high-performance' });
        const pixelRatio = Math.min(window.devicePixelRatio, mobile ? 1.5 : 2);
        renderer.setPixelRatio(pixelRatio);
        renderer.toneMapping = THREE.NeutralToneMapping;
        renderer.toneMappingExposure = 0.95;
        renderer.shadowMap.enabled = true;
        renderer.shadowMap.type = THREE.PCFShadowMap;
        renderer.domElement.setAttribute('aria-hidden', 'true');
        host.appendChild(renderer.domElement);

        const scene = new THREE.Scene();
        const camera = new THREE.PerspectiveCamera(28, 1, 0.1, 80);
        const sunDirection = new THREE.Vector3(
          Math.sin(SUN_AZIMUTH) * Math.cos(SUN_ELEVATION),
          Math.sin(SUN_ELEVATION),
          Math.cos(SUN_AZIMUTH) * Math.cos(SUN_ELEVATION),
        );

        // ---------- Céu físico (espalhamento da atmosfera e nuvens) ----------
        // Desenhado uma vez num cubo: vira o fundo e a luz ambiente da cena.
        const sky = new Sky();
        sky.scale.setScalar(10);
        const skyUniforms = sky.material.uniforms;
        skyUniforms.turbidity.value = 3;
        skyUniforms.rayleigh.value = 2;
        skyUniforms.mieCoefficient.value = 0.004;
        skyUniforms.mieDirectionalG.value = 0.82;
        skyUniforms.cloudCoverage.value = 0.38;
        skyUniforms.cloudDensity.value = 0.5;
        skyUniforms.sunPosition.value.copy(sunDirection);
        const skyScene = new THREE.Scene();
        skyScene.add(sky);
        const skyTarget = new THREE.WebGLCubeRenderTarget(mobile ? 256 : 512, { type: THREE.HalfFloatType });
        const skyCamera = new THREE.CubeCamera(0.1, 100, skyTarget);
        skyCamera.update(renderer, skyScene);
        scene.background = skyTarget.texture;
        const pmrem = new THREE.PMREMGenerator(renderer);
        const environment = pmrem.fromScene(skyScene, 0.02).texture;
        scene.environment = environment;
        scene.environmentIntensity = 0.55;
        scene.fog = new THREE.Fog(HAZE, 8, 21);

        const sun = new THREE.DirectionalLight('#fff0d8', 3.4);
        sun.castShadow = true;
        const shadowSize = mobile ? 1024 : 2048;
        sun.shadow.mapSize.set(shadowSize, shadowSize);
        Object.assign(sun.shadow.camera, { left: -3, right: 3, top: 3, bottom: -3, near: 0.5, far: 20 });
        sun.shadow.bias = -0.0003;
        sun.shadow.normalBias = 0.02;
        sun.shadow.radius = 3;
        scene.add(sun, sun.target);
        scene.add(new THREE.HemisphereLight('#cfe0ee', '#4d5a36', 0.35));

        // ---------- Terreno: relevo iluminado com a carta sobreposta ----------
        const terrainGeometry = new THREE.PlaneGeometry(WIDTH, DEPTH, mobile ? 110 : 180, mobile ? 80 : 130);
        terrainGeometry.rotateX(-Math.PI / 2);
        const positions = terrainGeometry.attributes.position;
        const baseX = Float32Array.from({ length: positions.count }, (_, i) => positions.getX(i));
        const baseZ = Float32Array.from({ length: positions.count }, (_, i) => positions.getZ(i));

        const offsetUniform = { value: 0 };
        const terrainMaterial = new THREE.MeshStandardMaterial({ color: '#ffffff', roughness: 0.95, metalness: 0 });
        terrainMaterial.onBeforeCompile = (shader) => {
          shader.uniforms.uOffset = offsetUniform;
          shader.vertexShader = shader.vertexShader
            .replace('#include <common>', '#include <common>\nuniform float uOffset;\nvarying vec3 vMap;\nvarying vec3 vUp;')
            .replace(
              '#include <begin_vertex>',
              '#include <begin_vertex>\nvMap = vec3(position.x + uOffset, position.y, position.z);\nvUp = normal;',
            );
          shader.fragmentShader = shader.fragmentShader
            .replace(
              '#include <common>',
              `#include <common>
              varying vec3 vMap;
              varying vec3 vUp;
              float pathZ(float x) { return 0.8 * sin(x * 0.12); }
              float hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }
              float noise(vec2 p) {
                vec2 i = floor(p); vec2 f = fract(p); f = f * f * (3.0 - 2.0 * f);
                return mix(mix(hash(i), hash(i + vec2(1, 0)), f.x), mix(hash(i + vec2(0, 1)), hash(i + vec2(1, 1)), f.x), f.y);
              }
              float lineMask(float value, float width) {
                float d = abs(fract(value - 0.5) - 0.5);
                return 1.0 - smoothstep(0.0, fwidth(value) * width, d);
              }`,
            )
            .replace(
              'vec4 diffuseColor = vec4( diffuse, opacity );',
              `vec2 p = vMap.xz;
              float h = vMap.y;
              float n = noise(p * 1.7) * 0.6 + noise(p * 7.0) * 0.4;
              float fine = noise(p * 31.0);
              float slope = 1.0 - clamp(vUp.y, 0.0, 1.0);
              // Pasto com manchas de capim seco e terra exposta nas encostas.
              vec3 color = mix(vec3(0.27, 0.38, 0.17), vec3(0.47, 0.52, 0.26), smoothstep(-0.4, 0.7, h + (n - 0.5) * 0.7));
              color = mix(color, vec3(0.62, 0.58, 0.36), smoothstep(0.62, 0.8, noise(p * 0.9 + 4.0)) * 0.55);
              color *= 0.82 + fine * 0.3;
              color = mix(color, vec3(0.5, 0.42, 0.3), smoothstep(0.2, 0.45, slope + (n - 0.5) * 0.2));
              color = mix(color, vec3(0.22, 0.36, 0.44), smoothstep(-0.62, -0.7, h) * 0.85);
              // Curvas de nível da carta: finas e mestras (a cada cinco).
              float k = h * 7.0;
              color = mix(color, vec3(0.3, 0.22, 0.12), lineMask(k, 1.0) * 0.22);
              color = mix(color, vec3(0.3, 0.22, 0.12), lineMask(k / 5.0, 1.8) * 0.42);
              // Estrada: acostamento de cascalho, asfalto, bordas brancas e a faixa vermelha.
              float d = abs(p.y - pathZ(p.x));
              float aa = fwidth(d);
              float shoulder = 1.0 - smoothstep(${(ROAD_HALF + 0.09).toFixed(2)}, ${(ROAD_HALF + 0.09).toFixed(2)} + aa * 2.0 + n * 0.04, d);
              float asphalt = 1.0 - smoothstep(${ROAD_HALF.toFixed(2)}, ${ROAD_HALF.toFixed(2)} + aa * 2.0, d);
              color = mix(color, vec3(0.5, 0.46, 0.39) * (0.8 + fine * 0.4), shoulder);
              vec3 tarmac = vec3(0.14, 0.145, 0.15) * (0.82 + noise(p * 40.0) * 0.36);
              // Trilha das rodas mais lisa e escura no centro de cada mão.
              tarmac *= 1.0 - 0.12 * (1.0 - smoothstep(0.02, 0.09, abs(d - 0.14)));
              color = mix(color, tarmac, asphalt);
              float edge = 1.0 - smoothstep(0.012, 0.012 + aa * 1.5, abs(d - ${(ROAD_HALF - 0.04).toFixed(2)}));
              color = mix(color, vec3(0.82, 0.82, 0.78) * (0.85 + fine * 0.2), edge * asphalt);
              float dash = step(0.42, fract(p.x * 0.9));
              color = mix(color, vec3(0.72, 0.17, 0.1), (1.0 - smoothstep(0.014, 0.014 + aa * 1.5, d)) * dash);
              // As cores acima estão em sRGB; a luz física trabalha em linear.
              vec4 diffuseColor = vec4(pow(color, vec3(2.2)), opacity);`,
            )
            .replace(
              '#include <roughnessmap_fragment>',
              // Asfalto e faixa pintada refletem um pouco mais que o pasto.
              '#include <roughnessmap_fragment>\nroughnessFactor = mix(roughnessFactor, 0.72, asphalt);',
            );
        };
        const terrainMesh = new THREE.Mesh(terrainGeometry, terrainMaterial);
        terrainMesh.receiveShadow = true;
        scene.add(terrainMesh);

        function updateTerrain(offset: number) {
          for (let i = 0; i < positions.count; i += 1) {
            positions.setY(i, ground(baseX[i] + offset, baseZ[i]));
          }
          positions.needsUpdate = true;
          terrainGeometry.computeVertexNormals();
          offsetUniform.value = offset;
        }

        // ---------- Capim: tufos instanciados que balançam com o vento ----------
        const blades: number[] = [];
        const shades: number[] = [];
        for (let b = 0; b < 7; b += 1) {
          const angle = random() * Math.PI;
          const lean = (random() - 0.5) * 0.5;
          const height = 0.08 + random() * 0.09;
          const width = 0.012 + random() * 0.008;
          const cx = (random() - 0.5) * 0.06;
          const cz = (random() - 0.5) * 0.06;
          const ax = Math.cos(angle);
          const az = Math.sin(angle);
          const point = (side: number, t: number) => {
            const w = width * (1 - t) * side;
            return [cx + ax * w + lean * t * t * height * az, t * height, cz + az * w - lean * t * t * height * ax];
          };
          // Folha em dois segmentos, afinando até a ponta.
          const quad = [point(-1, 0), point(1, 0), point(-1, 0.5), point(1, 0.5), point(0, 1)];
          [
            [0, 1, 2],
            [1, 3, 2],
            [2, 3, 4],
          ].forEach((tri) =>
            tri.forEach((index) => {
              blades.push(...quad[index]);
              shades.push(quad[index][1] / height);
            }),
          );
        }
        const grassGeometry = new THREE.BufferGeometry();
        grassGeometry.setAttribute('position', new THREE.Float32BufferAttribute(blades, 3));
        grassGeometry.setAttribute(
          'normal',
          new THREE.Float32BufferAttribute(Array.from({ length: blades.length / 3 }, () => [0, 1, 0]).flat(), 3),
        );
        grassGeometry.setAttribute('aTip', new THREE.Float32BufferAttribute(shades, 1));
        const timeUniform = { value: 0 };
        const grassMaterial = new THREE.MeshStandardMaterial({ roughness: 0.85, side: THREE.DoubleSide });
        grassMaterial.onBeforeCompile = (shader) => {
          shader.uniforms.uTime = timeUniform;
          shader.vertexShader = shader.vertexShader
            .replace('#include <common>', '#include <common>\nuniform float uTime;\nattribute float aTip;\nvarying float vTip;')
            .replace(
              '#include <begin_vertex>',
              `#include <begin_vertex>
              vTip = aTip;
              vec4 root = instanceMatrix * vec4(0.0, 0.0, 0.0, 1.0);
              float gust = sin(uTime * 1.7 + root.x * 1.3 + root.z * 0.7) * 0.5 + sin(uTime * 3.1 + root.x * 2.9) * 0.2;
              transformed.x += gust * 0.035 * aTip * aTip;`,
            );
          shader.fragmentShader = shader.fragmentShader
            .replace('#include <common>', '#include <common>\nvarying float vTip;')
            // Base escura (sombra entre as folhas) e ponta clara, translúcida ao sol.
            .replace('vec4 diffuseColor = vec4( diffuse, opacity );', 'vec4 diffuseColor = vec4(diffuse * mix(0.35, 1.25, vTip), opacity);');
        };
        const grassCount = mobile ? 2200 : 5200;
        const grass = new THREE.InstancedMesh(grassGeometry, grassMaterial, grassCount);
        grass.receiveShadow = true;
        grass.frustumCulled = false;
        const grassMap = new Float32Array(grassCount * 2);
        const grassMatrix = new THREE.Matrix4();
        const grassColor = new THREE.Color();
        const dry = new THREE.Color('#a39a5a');
        for (let i = 0; i < grassCount; i += 1) {
          grassMap[i * 2] = random() * GRASS_SPAN;
          grassMap[i * 2 + 1] = (random() - 0.5) * 2 * GRASS_DEPTH;
          const scale = 0.7 + random() * 0.8;
          grassMatrix.compose(
            new THREE.Vector3(),
            new THREE.Quaternion().setFromAxisAngle(new THREE.Vector3(0, 1, 0), random() * Math.PI * 2),
            new THREE.Vector3(scale, scale * (0.8 + random() * 0.5), scale),
          );
          grass.setMatrixAt(i, grassMatrix);
          grassColor.setHSL(0.2 + random() * 0.06, 0.45 + random() * 0.2, 0.2 + random() * 0.1, THREE.SRGBColorSpace);
          if (random() < 0.25) grassColor.lerp(dry, 0.6);
          grass.setColorAt(i, grassColor);
        }
        const grassBase = Float32Array.from(grass.instanceMatrix.array);
        scene.add(grass);

        // Cada tufo tem um lugar fixo no mapa; a faixa visível corre com o
        // ciclista e o que sai por trás volta pela frente.
        function updateGrass(offset: number, z: number) {
          const array = grass.instanceMatrix.array as Float32Array;
          for (let i = 0; i < grassCount; i += 1) {
            const x = ((((grassMap[i * 2] - offset) % GRASS_SPAN) + GRASS_SPAN) % GRASS_SPAN) - GRASS_SPAN * 0.45;
            const mapX = x + offset;
            const mapZ = grassMap[i * 2 + 1] + z;
            const onRoad = Math.abs(mapZ - pathZ(mapX)) < ROAD_HALF + 0.14;
            const k = onRoad ? 0 : 1;
            const o = i * 16;
            for (let e = 0; e < 11; e += 1) array[o + e] = grassBase[o + e] * k;
            array[o + 12] = x;
            array[o + 13] = ground(mapX, mapZ);
            array[o + 14] = mapZ;
          }
          grass.instanceMatrix.needsUpdate = true;
        }

        // ---------- Ciclista (components/ride-scene-rider.ts) ----------
        const rider = buildRider(THREE);
        const bike = rider.root;
        scene.add(bike);

        // ---------- Pós-processamento ----------
        // Oclusão de ambiente para assentar ciclista e capim no chão, lente com
        // profundidade de campo no computador e saída com o tone mapping físico.
        const target = new THREE.WebGLRenderTarget(1, 1, { type: THREE.HalfFloatType, samples: 4 });
        const composer = new EffectComposer(renderer, target);
        composer.setPixelRatio(pixelRatio);
        composer.addPass(new RenderPass(scene, camera));
        const gtao = new GTAOPass(scene, camera, 1, 1);
        gtao.updateGtaoMaterial({
          radius: 0.35,
          distanceExponent: 1.4,
          thickness: 1,
          scale: 1.1,
          samples: mobile ? 8 : 16,
        });
        gtao.blendIntensity = 0.9;
        composer.addPass(gtao);
        const bokeh = mobile
          ? null
          : new BokehPass(scene, camera, { focus: 3.4, aperture: 0.002, maxblur: 0.006 });
        if (bokeh) composer.addPass(bokeh);
        composer.addPass(new OutputPass());

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
          composer.setSize(width, height);
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
          timeUniform.value = elapsed;
          updateTerrain(offset);

          // O ciclista fica no x = 0 e o mundo corre por baixo dele. Cada roda
          // toca o greide no seu ponto de contato: a inclinação sai exata das
          // duas alturas e a bicicleta nunca afunda nem flutua.
          const z = pathZ(offset);
          const yaw = -Math.atan(pathZ(offset + 0.5) - pathZ(offset - 0.5));
          const rearY = roadY(offset + REAR_X * Math.cos(yaw));
          const frontY = roadY(offset + FRONT_X * Math.cos(yaw));
          const sinPitch = (frontY - rearY) / (FRONT_X - REAR_X);
          const pitch = Math.asin(sinPitch);
          bike.position.set(0, rearY - REAR_X * sinPitch, z);
          bike.rotation.set(0, yaw, pitch, 'YZX');
          rider.pose(crankAngle, wheelAngle);
          updateGrass(offset, z);

          sun.position.copy(bike.position).addScaledVector(sunDirection, 8);
          sun.target.position.copy(bike.position);

          pointer.x += (pointer.tx - pointer.x) * 0.04;
          pointer.y += (pointer.ty - pointer.y) * 0.04;
          const narrow = camera.aspect < 1.2;
          // Vista de três quartos, baixa como a de um fotógrafo na beira da estrada.
          const orbit = 0.42 + Math.sin(elapsed * 0.1) * 0.22 + pointer.x * 0.25;
          const distance = narrow ? 4.4 : 3.4;
          camera.position.set(
            Math.sin(orbit) * distance,
            bike.position.y + 0.85 - pointer.y * 0.2,
            z + Math.cos(orbit) * distance,
          );
          lookAt.set(narrow ? 0.35 : 0.05, bike.position.y + 0.7, z);
          camera.lookAt(lookAt);
          if (bokeh) (bokeh.uniforms as { focus: { value: number } }).focus.value = camera.position.distanceTo(lookAt);
          composer.render(dt);
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
          grassGeometry.dispose();
          grassMaterial.dispose();
          grass.dispose();
          sky.geometry.dispose();
          sky.material.dispose();
          skyTarget.dispose();
          environment.dispose();
          pmrem.dispose();
          rider.dispose();
          composer.dispose();
          gtao.dispose();
          bokeh?.dispose();
          renderer.dispose();
          renderer.domElement.remove();
        };
      })
      .catch(() => undefined);

    return () => {
      disposed = true;
      cleanup();
    };
  }, []);

  return <div ref={hostRef} className="ride-scene" role="img" aria-label={label} />;
}
