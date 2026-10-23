// Ciclista e bicicleta da cena de acesso (components/ride-scene.tsx), em
// estilo de ilustração: sombreamento em três faixas e contorno em tinta, como
// numa carta impressa. Recebe o módulo do Three.js já carregado, para que o
// pacote só seja baixado na tela de acesso.
//
// Eixos locais: +x é a frente da bicicleta, +y para cima, z para os lados.

type Three = typeof import('three');
type Vec3 = import('three').Vector3;
type Obj3D = import('three').Object3D;
type Geometry = import('three').BufferGeometry;
type Material = import('three').Material;

export const WHEEL_R = 0.34;
export const CRANK_R = 0.17;

const COLORS = {
  ink: '#1f231c',
  trail: '#c3321f',
  trailDeep: '#9c2616',
  sheet: '#f8f9f5',
  skin: '#c69c76',
  water: '#2f6f9e',
};

// Contorno: uma casca um pouco maior, só com as faces de trás, em tinta.
const OUTLINE = 0.011;
const THIGH = 0.45;
const SHIN = 0.44;
const UPPER_ARM = 0.29;
const FOREARM = 0.28;

export type Rider = {
  root: Obj3D;
  pose: (crankAngle: number, wheelAngle: number) => void;
  dispose: () => void;
};

export function buildRider(THREE: Three): Rider {
  const disposables: Array<{ dispose(): void }> = [];
  const keep = <T extends { dispose(): void }>(item: T) => {
    disposables.push(item);
    return item;
  };

  const tones = new Uint8Array([110, 190, 255]);
  const gradientMap = keep(new THREE.DataTexture(tones, tones.length, 1, THREE.RedFormat));
  gradientMap.minFilter = THREE.NearestFilter;
  gradientMap.magFilter = THREE.NearestFilter;
  gradientMap.needsUpdate = true;
  const toon = (color: string) => keep(new THREE.MeshToonMaterial({ color, gradientMap }));
  const ink = toon(COLORS.ink);
  const red = toon(COLORS.trail);
  const redDeep = toon(COLORS.trailDeep);
  const sheet = toon(COLORS.sheet);
  const skin = toon(COLORS.skin);
  const outline = keep(new THREE.MeshBasicMaterial({ color: COLORS.ink, side: THREE.BackSide }));
  const lineInk = keep(new THREE.LineBasicMaterial({ color: COLORS.ink }));

  const v = (x: number, y: number, z = 0) => new THREE.Vector3(x, y, z);
  const up = v(0, 1, 0);

  // Peça com contorno: a malha e a casca usam geometrias do mesmo formato.
  function inked(geometry: Geometry, shell: Geometry | null, material: Material) {
    const group = new THREE.Group();
    if (shell) group.add(new THREE.Mesh(keep(shell), outline));
    group.add(new THREE.Mesh(keep(geometry), material));
    return group;
  }
  const capsule = (radius: number, length: number, material: Material, line = OUTLINE) =>
    inked(
      new THREE.CapsuleGeometry(radius, length, 4, 12),
      line ? new THREE.CapsuleGeometry(radius + line, length, 4, 12) : null,
      material,
    );
  const ball = (radius: number, material: Material, line = OUTLINE) =>
    inked(
      new THREE.SphereGeometry(radius, 18, 12),
      line ? new THREE.SphereGeometry(radius + line, 18, 12) : null,
      material,
    );

  // Coloca uma peça alongada (eixo y) entre dois pontos.
  function between(part: Obj3D, a: Vec3, b: Vec3) {
    const dir = new THREE.Vector3().subVectors(b, a);
    part.position.copy(a).addScaledVector(dir, 0.5);
    part.quaternion.setFromUnitVectors(up, dir.normalize());
    return part;
  }
  function tube(parent: Obj3D, a: Vec3, b: Vec3, radius: number, material: Material, line = OUTLINE * 0.6) {
    const part = between(capsule(radius, a.distanceTo(b), material, line), a, b);
    parent.add(part);
    return part;
  }
  function add(parent: Obj3D, part: Obj3D) {
    parent.add(part);
    return part;
  }
  function at(parent: Obj3D, part: Obj3D, point: Vec3) {
    part.position.copy(point);
    parent.add(part);
    return part;
  }

  const root = new THREE.Group();

  // Sombra macia no chão, sob a bicicleta.
  const shadowMaterial = keep(
    new THREE.ShaderMaterial({
      transparent: true,
      depthWrite: false,
      uniforms: { uColor: { value: new THREE.Color(COLORS.ink) } },
      vertexShader: 'varying vec2 vUv; void main() { vUv = uv; gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0); }',
      fragmentShader:
        'uniform vec3 uColor; varying vec2 vUv; void main() { float d = length((vUv - 0.5) * 2.0); gl_FragColor = vec4(uColor, (1.0 - smoothstep(0.2, 1.0, d)) * 0.28); }',
    }),
  );
  const shadow = new THREE.Mesh(keep(new THREE.PlaneGeometry(1.7, 0.62)), shadowMaterial);
  shadow.rotation.x = -Math.PI / 2;
  shadow.position.set(0.02, 0.015, 0);
  shadow.renderOrder = -1;
  root.add(shadow);

  // ---------- Bicicleta ----------
  const bike = new THREE.Group();
  root.add(bike);
  const rear = v(-0.5, WHEEL_R);
  const front = v(0.53, WHEEL_R);
  const bb = v(0, 0.29);
  const seatCluster = v(-0.14, 0.78);
  const headTop = v(0.39, 0.84);
  const headLow = v(0.43, 0.69);
  const bar = v(0.46, 0.92);

  // Rodas de perfil alto: pneu, aro largo, raios e cubo.
  const spokeVertices: number[] = [];
  const SPOKES = 18;
  for (let i = 0; i < SPOKES; i += 1) {
    const angle = (i / SPOKES) * Math.PI * 2;
    const z = i % 2 ? 0.025 : -0.025;
    spokeVertices.push(0, 0, z, Math.cos(angle) * (WHEEL_R - 0.07), Math.sin(angle) * (WHEEL_R - 0.07), 0);
  }
  const spokes = keep(new THREE.BufferGeometry());
  spokes.setAttribute('position', new THREE.Float32BufferAttribute(spokeVertices, 3));
  const wheels = [rear, front].map((center) => {
    const wheel = new THREE.Group();
    wheel.position.copy(center);
    wheel.add(inked(new THREE.TorusGeometry(WHEEL_R, 0.027, 10, 64), null, ink));
    const rim = inked(
      new THREE.TorusGeometry(WHEEL_R - 0.05, 0.03, 8, 64),
      new THREE.TorusGeometry(WHEEL_R - 0.05, 0.03 + 0.006, 8, 64),
      sheet,
    );
    rim.scale.z = 0.45;
    wheel.add(rim);
    wheel.add(new THREE.LineSegments(spokes, lineInk));
    const hub = inked(new THREE.CylinderGeometry(0.028, 0.028, 0.09, 12), null, sheet);
    hub.rotation.x = Math.PI / 2;
    wheel.add(hub);
    bike.add(wheel);
    return wheel;
  });

  // Quadro vermelho com contorno, garfo e canote em tinta.
  tube(bike, bb, seatCluster, 0.022, red);
  tube(bike, seatCluster, headTop, 0.02, red);
  tube(bike, bb, headLow, 0.027, red);
  tube(bike, headLow, headTop, 0.029, redDeep);
  for (const side of [-0.055, 0.055]) {
    tube(bike, v(rear.x, rear.y, side), v(seatCluster.x - 0.01, seatCluster.y - 0.04, side * 0.3), 0.012, red);
    tube(bike, v(rear.x, rear.y, side), v(bb.x - 0.02, bb.y, side * 0.5), 0.014, red);
    tube(bike, v(headLow.x, headLow.y, side * 0.4), v(front.x, front.y, side), 0.013, ink, 0);
  }
  const seatTop = v(-0.16, 0.9);
  tube(bike, seatCluster, seatTop, 0.013, ink, 0);
  const saddle = at(bike, ball(1, ink, 0), v(-0.17, 0.915));
  saddle.scale.set(0.14, 0.022, 0.055);

  // Guidão drop: mesa, barra e as curvas de cada lado, com manetes.
  tube(bike, headTop, bar, 0.016, ink, 0);
  tube(bike, v(bar.x, bar.y, -0.2), v(bar.x, bar.y, 0.2), 0.013, ink, 0);
  for (const side of [-0.2, 0.2]) {
    const curve = new THREE.CatmullRomCurve3([
      v(bar.x, bar.y, side),
      v(bar.x + 0.11, bar.y - 0.01, side),
      v(bar.x + 0.13, bar.y - 0.09, side),
      v(bar.x + 0.07, bar.y - 0.15, side),
      v(bar.x - 0.02, bar.y - 0.15, side),
    ]);
    bike.add(inked(new THREE.TubeGeometry(curve, 24, 0.012, 8), null, ink));
    const hood = at(bike, ball(1, ink, 0), v(bar.x + 0.115, bar.y + 0.02, side));
    hood.scale.set(0.04, 0.03, 0.02);
  }

  // Caramanhola no tubo inferior.
  const bottleAt = v(0.17, 0.52, 0);
  const bottle = at(bike, capsule(0.032, 0.12, toon(COLORS.water), OUTLINE * 0.6), bottleAt);
  bottle.rotation.z = -0.95;
  bottle.position.add(v(-0.03, 0.045));

  // Transmissão: coroa, corrente, pedivela e pedais.
  const ringGroup = inked(new THREE.TorusGeometry(0.1, 0.012, 6, 36), null, ink);
  ringGroup.position.set(bb.x, bb.y, 0.07);
  bike.add(ringGroup);
  const chainPoints = [
    v(bb.x, bb.y + 0.1, 0.07),
    v(rear.x, rear.y + 0.045, 0.07),
    v(rear.x - 0.03, rear.y, 0.07),
    v(rear.x, rear.y - 0.045, 0.07),
    v(bb.x, bb.y - 0.1, 0.07),
  ];
  bike.add(new THREE.Line(keep(new THREE.BufferGeometry().setFromPoints(chainPoints)), lineInk));
  const crank = new THREE.Group();
  crank.position.copy(bb);
  bike.add(crank);
  [0.09, -0.09].forEach((z, index) => {
    const arm = new THREE.Group();
    arm.position.z = z;
    arm.rotation.z = index * Math.PI;
    tube(arm, v(0, 0), v(CRANK_R, 0), 0.014, sheet, 0);
    crank.add(arm);
  });

  // ---------- Ciclista ----------
  const rider = new THREE.Group();
  bike.add(rider);
  const hip = v(-0.17, 0.97);
  const shoulder = v(0.19, 1.27);
  tube(rider, hip, shoulder, 0.1, red, OUTLINE);
  at(rider, ball(0.1, ink), hip).scale.set(1.1, 0.95, 1.25);
  // Faixa branca nas costas da camisa.
  tube(rider, v(-0.12, 1.07, 0), v(0.12, 1.25, 0), 0.103, sheet, 0).scale.set(1, 0.25, 1);
  const head = v(0.33, 1.39);
  tube(rider, v(shoulder.x + 0.04, shoulder.y + 0.03), v(head.x - 0.03, head.y - 0.05), 0.035, skin, OUTLINE * 0.7);
  at(rider, ball(0.075, skin), head);
  const helmet = at(rider, ball(0.08, sheet), v(head.x - 0.025, head.y + 0.035));
  helmet.scale.set(1.45, 0.82, 1.1);
  helmet.rotation.z = 0.25;
  const glasses = at(rider, ball(1, ink, 0), v(head.x + 0.07, head.y + 0.005));
  glasses.scale.set(0.012, 0.02, 0.075);

  const legs = [0.1, -0.1].map((z) => ({
    z,
    thigh: add(rider, capsule(0.056, THIGH, ink)),
    shin: add(rider, capsule(0.041, SHIN, skin)),
    shoe: at(rider, ball(1, ink, 0), v(0, 0)),
  }));
  legs.forEach((leg) => leg.shoe.scale.set(0.085, 0.034, 0.042));
  const pedals = legs.map(() => {
    const pedal = at(bike, ball(1, sheet, 0), v(0, 0));
    pedal.scale.set(0.05, 0.012, 0.04);
    return pedal;
  });
  const arms = [0.16, -0.16].map((z) => ({
    z,
    upper: add(rider, capsule(0.04, UPPER_ARM, red)),
    fore: add(rider, capsule(0.032, FOREARM, skin)),
    glove: at(rider, ball(0.035, ink, 0), v(0, 0)),
  }));

  // Dois segmentos entre a e b, com a articulação para o lado indicado
  // (joelho para a frente, cotovelo para baixo), no plano xy.
  const joint = new THREE.Vector3();
  function solve(a: Vec3, b: Vec3, first: number, second: number, bend: 1 | -1) {
    const dx = b.x - a.x;
    const dy = b.y - a.y;
    const dist = Math.min(Math.hypot(dx, dy), first + second - 0.001);
    const base = Math.atan2(dy, dx);
    const cos = (first * first + dist * dist - second * second) / (2 * first * dist);
    const angle = base + bend * Math.acos(Math.max(-1, Math.min(1, cos)));
    return joint.set(a.x + Math.cos(angle) * first, a.y + Math.sin(angle) * first, (a.z + b.z) / 2);
  }

  function pose(crankAngle: number, wheelAngle: number) {
    wheels.forEach((wheel) => {
      wheel.rotation.z = -wheelAngle;
    });
    crank.rotation.z = -crankAngle;
    // O corpo acompanha a pedalada: quadril balança e o tronco sobe e desce de leve.
    const rock = Math.sin(crankAngle * 2);
    rider.rotation.x = rock * 0.025;
    legs.forEach((leg, index) => {
      const angle = -crankAngle + index * Math.PI;
      const foot = v(bb.x + Math.cos(angle) * CRANK_R, bb.y + Math.sin(angle) * CRANK_R, leg.z);
      pedals[index].position.copy(foot);
      const hipPoint = v(hip.x, hip.y + rock * 0.006, leg.z * 0.85);
      const knee = solve(hipPoint, foot, THIGH, SHIN, 1).clone();
      between(leg.thigh, hipPoint, knee);
      between(leg.shin, knee, foot);
      leg.shoe.position.copy(foot).add(v(0.015, 0.03));
      leg.shoe.rotation.z = Math.sin(angle) * 0.25;
    });
    arms.forEach((arm) => {
      const from = v(shoulder.x, shoulder.y + rock * 0.004, arm.z * 0.75);
      const hand = v(bar.x + 0.11, bar.y + 0.035, arm.z * 1.25);
      const elbow = solve(from, hand, UPPER_ARM, FOREARM, -1).clone();
      elbow.z = arm.z * 1.15;
      between(arm.upper, from, elbow);
      between(arm.fore, elbow, hand);
      arm.glove.position.copy(hand);
    });
  }

  return {
    root,
    pose,
    dispose: () => disposables.forEach((item) => item.dispose()),
  };
}
