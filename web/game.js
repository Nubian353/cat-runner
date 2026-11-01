/* Котик-бегун — веб-версия для iPad/Safari.
   Тот же забег, что в терминальной игре: лужи, три жизни, собака,
   финал с подушками, молоком, подгузниками, шапочками и сосочками. */
(() => {
  'use strict';

  const COLS = 80, ROWS = 25, ROAD = 5, RT = ROWS - ROAD;
  const MAX_LIVES = 3, MAX_STAGES = 2, CAT_W = 9, HITBOX = 5.5;
  const JUMP = 26, GRAVITY = 60;
  const GOALS = [650, 850], BASE_SPEED = [17, 26], MAX_SPEED = [30, 40], RAMP = [0.32, 0.48];
  const PILLOW_W = 11, PILLOW_GAP = 3;

  // ---- палитра xterm-256 ----
  const PAL = new Array(256);
  const BASE = [[0,0,0],[128,0,0],[0,128,0],[128,128,0],[0,0,128],[128,0,128],[0,128,128],[192,192,192],
                [128,128,128],[255,0,0],[0,255,0],[255,255,0],[0,0,255],[255,0,255],[0,255,255],[255,255,255]];
  for (let i = 0; i < 16; i++) PAL[i] = `rgb(${BASE[i][0]},${BASE[i][1]},${BASE[i][2]})`;
  for (let i = 16; i < 232; i++) {
    const n = i - 16, r = (n / 36) | 0, g = ((n % 36) / 6) | 0, b = n % 6;
    const f = v => v === 0 ? 0 : 55 + v * 40;
    PAL[i] = `rgb(${f(r)},${f(g)},${f(b)})`;
  }
  for (let i = 232; i < 256; i++) { const v = 8 + (i - 232) * 10; PAL[i] = `rgb(${v},${v},${v})`; }

  // ---- спрайты ----
  const CAT_IDLE = ['  /\\_/\\  ', ' ( o.o ) ', '  > ^ <  ', ' /|   |\\ ', '  ^   ^  '];
  const CAT_RUN_B = ['  /\\_/\\  ', ' ( o.o ) ', '  > ^ <  ', '  |   |\\ ', '  ^    ^ '];
  const CAT_JUMP = ['  /\\_/\\  ', ' ( o.o ) ', '  > ^ <  ', '  /| |\\  ', '  ^   ^  '];
  const CAT_SCARED = ['  /\\_/\\  ', ' ( O.O ) ', '  > ^ <  ', ' /|   |\\ ', '  ^   ^  '];
  const CAT_SLEEP = ['  /\\_/\\  ', ' ( -.- ) ', '  >   <  ', ' /_____\\ '];
  const DOG_A = ['  ▐▄▄▄▄▄▌  ', '  ████████  ', '  █ o  o █  ', '  █ ▄██▄ █  ', '   ▀████▀   ', '   █    █   '];
  const DOG_B = ['  ▐▄▄▄▄▄▌  ', '  ████████  ', '  █ o  o █  ', '  █ ▄██▄ █  ', '   ▀████▀   ', '   █   █    '];

  const ORANGE = { body: 208, ear: 214, eye: 231, nose: 217 };
  const GRAY = { body: 250, ear: 249, eye: 39, nose: 218 };

  // ---- холст ----
  const canvas = document.getElementById('game');
  const ctx = canvas.getContext('2d', { alpha: false });
  let CW = 10, CH = 18, OX = 0, OY = 0, FONT = 16, DPR = 1;

  function resize() {
    const w = window.innerWidth, h = window.innerHeight;
    DPR = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = Math.floor(w * DPR);
    canvas.height = Math.floor(h * DPR);
    canvas.style.width = w + 'px';
    canvas.style.height = h + 'px';
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
    CW = Math.max(4, Math.floor(Math.min(w / COLS, h / (ROWS * 1.62))));
    CH = Math.floor(CW * 1.62);
    OX = Math.floor((w - CW * COLS) / 2);
    OY = Math.floor((h - CH * ROWS) / 2);
    FONT = Math.max(8, Math.floor(CH * 0.94));
    ctx.font = `${FONT}px "SF Mono", Menlo, Consolas, monospace`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
  }
  window.addEventListener('resize', resize);
  window.addEventListener('orientationchange', () => setTimeout(resize, 120));

  function px(cx) { return OX + cx * CW; }
  function py(cy) { return OY + cy * CH; }
  // Блочные символы рисуем прямоугольниками: в шрифте у них есть боковые
  // отступы, и «█» не стыкуется с соседом, а в терминале — стыкуется.
  const BLOCK = {
    '█': [0, 0, 1, 1], '▀': [0, 0, 1, .5], '▄': [0, .5, 1, .5],
    '▐': [.5, 0, .5, 1], '▌': [0, 0, .5, 1],
    '▖': [0, .5, .5, .5], '▗': [.5, .5, .5, .5],
    '▘': [0, 0, .5, .5], '▝': [.5, 0, .5, .5],
    '▁': [0, .86, 1, .14], '━': [0, .44, 1, .12],
    '▒': [.12, .12, .76, .76], '▓': [.06, .06, .88, .88],
    '·': [.34, .34, .32, .32],
  };

  function chr(ch, cx, cy, color) {
    if (ch === ' ' || !ch) return;
    ctx.fillStyle = PAL[color];
    const b = BLOCK[ch];
    if (b) {
      ctx.fillRect(px(cx) + b[0] * CW, py(cy) + b[1] * CH, b[2] * CW, b[3] * CH);
      return;
    }
    ctx.fillText(ch, px(cx) + CW / 2, py(cy) + CH / 2 + 1);
  }
  function rect(cx, cy, w, h, color) {
    ctx.fillStyle = PAL[color];
    ctx.fillRect(px(cx), py(cy), w * CW, h * CH);
  }
  function circle(cx, cy, r, color) {
    ctx.fillStyle = PAL[color];
    ctx.beginPath();
    ctx.arc(px(cx) + CW / 2, py(cy) + CH / 2, r, 0, Math.PI * 2);
    ctx.fill();
  }
  function label(text, cx, cy, color, size) {
    ctx.fillStyle = PAL[color];
    ctx.font = `600 ${size}px "SF Mono", Menlo, Consolas, monospace`;
    ctx.fillText(text, px(cx), py(cy) + CH / 2);
    ctx.font = `${FONT}px "SF Mono", Menlo, Consolas, monospace`;
  }
  function sprite(rows, x, base, pal, blink) {
    for (let j = 0; j < rows.length; j++) {
      const yy = base - (rows.length - 1) + j;
      for (let i = 0; i < rows[j].length; i++) {
        let ch = rows[j][i];
        if (ch === ' ') continue;
        if (ch === 'o' && blink) ch = '-';
        let color = pal.body;
        if ('oOx-'.includes(ch)) color = pal.eye;
        else if (ch === '^') color = pal.nose;
        else if ('/\\_'.includes(ch)) color = pal.ear;
        chr(ch, x + i, yy, color);
      }
    }
  }

  // ---- состояние ----
  const S = {
    st: 'start', t: 0, runTime: 0, stageTime: 0,
    distance: 0, stageDistance: 0, stageGoal: GOALS[0], level: 1, speed: BASE_SPEED[0],
    points: 0, score: 0, lives: MAX_LIVES, invuln: 0, flash: 0,
    catX: 11, catY: 0, vy: 0, onGround: true,
    puddles: [], spawnIn: 1.2, particles: [], floats: [],
    dogX: 0, catEaten: false, lossFor: 0, lossDone: 0, clearFor: 0, overFor: 0,
    finishFor: 0, finishDone: 0, bites: 0, biteTimer: 0, sausageX: COLS - 14, sausageW: 12,
    cat2Active: false, cat2X: -CAT_W, pillowShown: false, pillowX: 0, asleep: false, sleepFor: 0,
    victoryFor: 0, newBest: false,
    best: Number(localStorage.getItem('cat-runner-best') || 0) || 0,
  };

  const rnd = (a, b) => a + Math.random() * (b - a);
  const clamp = (v, a, b) => Math.max(a, Math.min(b, v));

  function reset() {
    S.st = 'start'; S.t = 0; S.runTime = 0; S.stageTime = 0;
    S.distance = 0; S.stageDistance = 0; S.stageGoal = GOALS[0]; S.level = 1; S.speed = BASE_SPEED[0];
    S.points = 0; S.score = 0; S.lives = MAX_LIVES; S.invuln = 0; S.flash = 0;
    S.catX = clamp(COLS * 0.14, 8, 20); S.catY = 0; S.vy = 0; S.onGround = true;
    S.puddles = []; S.spawnIn = 1.2; S.particles = []; S.floats = [];
    S.dogX = 0; S.catEaten = false; S.lossFor = 0; S.lossDone = 0; S.clearFor = 0; S.overFor = 0;
    S.finishFor = 0; S.finishDone = 0; S.bites = 0; S.biteTimer = 0; S.sausageW = 12;
    S.cat2Active = false; S.cat2X = -CAT_W; S.pillowShown = false; S.pillowX = 0;
    S.asleep = false; S.sleepFor = 0; S.victoryFor = 0; S.newBest = false;
  }

  function startGame() {
    reset();
    S.st = 'running';
    S.catY = 2; S.vy = 0; S.onGround = false;
    spawnPuddle();
  }

  function jump() {
    if (!S.onGround || S.st !== 'running') return;
    S.onGround = false; S.vy = JUMP;
    addFloat('', S.catX + 3, RT - 1, 223, 0.3);
  }

  function spawnPuddle() {
    let w = 3 + Math.random() * 3;
    w += Math.min(2.2, S.stageTime * 0.05);
    w += (S.level - 1) * 0.6;
    S.puddles.push({ x: COLS + 2, w: Math.round(Math.min(9, w)), hit: false, passed: false });
  }
  function nextDelay() { return S.level >= 2 ? rnd(1.05, 1.8) : rnd(1.2, 2.05); }

  function addFloat(text, x, y, color, life) {
    S.floats.push({ text, x, y, color, life, max: life });
  }
  function dust() {
    for (let i = 0; i < 4; i++) S.particles.push({ x: S.catX + 2 + Math.random() * 4, y: RT - 0.2, vx: rnd(-3, 3), vy: rnd(-3, -1), life: 0.4, max: 0.4, color: 223, ch: '+' });
  }
  function splash(x) {
    for (let i = 0; i < 10; i++) S.particles.push({ x, y: RT - 0.3, vx: rnd(-8, 8), vy: rnd(-9, -2), life: 0.5, max: 0.5, color: 45, ch: '•' });
  }

  function updateEffects(dt, scale = 1) {
    for (const p of S.particles) {
      p.life -= dt * scale; p.x += p.vx * dt; p.y += p.vy * dt; p.vy += 26 * dt;
    }
    S.particles = S.particles.filter(p => p.life > 0);
    for (const f of S.floats) { f.life -= dt * scale; f.y -= dt * 2.2; }
    S.floats = S.floats.filter(f => f.life > 0);
  }

  function collide() {
    const catL = S.catX + 2.5, catR = catL + HITBOX;
    for (const p of S.puddles) {
      if (p.passed) continue;
      if (p.x + p.w < catL) {
        p.passed = true;
        if (!p.hit) {
          S.points += 15; S.score = Math.floor(S.points);
          addFloat('+15', p.x + p.w / 2, RT - 3, 220, 0.75);
        }
        continue;
      }
      if (p.hit || catR <= p.x || catL >= p.x + p.w) continue;
      if (S.catY > 1.45) continue;
      if (S.invuln > 0) { p.hit = true; continue; }
      p.hit = true; S.lives--; S.invuln = 1.5; S.flash = 1;
      splash(p.x + p.w / 2);
      addFloat('-1 ♥', S.catX + 4, RT - 6, 203, 1);
      if (S.lives <= 0) { beginDog(); return; }
    }
  }

  function beginStageClear() {
    S.st = 'stageclear'; S.clearFor = 0;
    S.points += 100; S.score = Math.floor(S.points);
    S.puddles = []; S.particles = []; S.floats = []; S.invuln = 0; S.flash = 0;
    S.catY = 0; S.vy = 0; S.onGround = true;
    addFloat('+100', COLS / 2, RT - 8, 220, 2);
  }

  function beginStage(level) {
    S.level = level; S.st = 'running'; S.stageTime = 0; S.stageDistance = 0;
    S.stageGoal = GOALS[level - 1]; S.puddles = []; S.spawnIn = 1.2;
    S.catY = 0; S.vy = 0; S.onGround = true;
  }

  function beginFinish() {
    S.st = 'finish'; S.finishFor = 0; S.finishDone = 0; S.bites = 0; S.biteTimer = 0;
    S.cat2Active = false; S.cat2X = -CAT_W; S.pillowShown = false; S.asleep = false;
    S.puddles = []; S.particles = []; S.floats = [];
  }

  function beginDog() {
    S.st = 'dog'; S.catEaten = false; S.lossFor = 0; S.lossDone = 0;
    S.dogX = COLS + 2; S.invuln = 0;
  }

  function finishRun() {
    S.st = 'victory'; S.victoryFor = 0;
    S.score = Math.floor(S.points);
    S.newBest = S.score > S.best;
    if (S.newBest) { S.best = S.score; localStorage.setItem('cat-runner-best', String(S.best)); }
  }

  function updateFinish(dt) {
    S.finishFor += dt; S.runTime += dt;
    if (!S.pillowShown) {
      if (!S.cat2Active) {
        const target = S.sausageX - CAT_W;
        if (S.catX < target) { S.catX = Math.min(target, S.catX + 42 * dt); }
        else { S.cat2Active = true; S.cat2X = COLS + 1; }
      } else {
        const target = S.sausageX + S.sausageW + 1;
        if (S.cat2X > target) S.cat2X = Math.max(target, S.cat2X - 42 * dt);
        else {
          S.biteTimer += dt;
          if (S.biteTimer >= 0.55 && S.bites < 3) {
            S.biteTimer = 0; S.bites++;
            S.sausageW = Math.max(4, S.sausageW - 2);
            S.points += 15; S.score = Math.floor(S.points);
            addFloat('НЯМ!', S.sausageX + 4, RT - 7, 220, 0.7);
          }
          if (S.bites >= 3) {
            S.finishDone += dt;
            if (S.finishDone >= 0.8) {
              S.pillowShown = true;
              S.pillowX = clamp(S.catX - 6, 5, COLS - 30);
              addFloat('СПАТЬ!', S.pillowX + 12, RT - 8, 220, 1.2);
            }
          }
        }
      }
    } else {
      if (!S.asleep) {
        const t1 = S.pillowX + 1, t2 = S.pillowX + PILLOW_W + PILLOW_GAP + 1;
        S.catX = S.catX < t1 ? Math.min(t1, S.catX + 38 * dt) : Math.max(t1, S.catX - 38 * dt);
        S.cat2X = S.cat2X < t2 ? Math.min(t2, S.cat2X + 38 * dt) : Math.max(t2, S.cat2X - 38 * dt);
        if (Math.abs(S.catX - t1) < 0.1 && Math.abs(S.cat2X - t2) < 0.1) {
          S.catX = t1; S.cat2X = t2; S.asleep = true; S.sleepFor = 0;
          addFloat('МУР-МУР!', S.pillowX + 13, RT - 9, 220, 1.6);
        }
      } else {
        S.sleepFor += dt;
        if (S.sleepFor >= 1.8) { finishRun(); return; }
      }
    }
    updateEffects(dt);
  }

  function updateDog(dt) {
    S.lossFor += dt;
    const target = S.catX - 3;
    if (!S.catEaten) {
      if (S.dogX > target) S.dogX = Math.max(target, S.dogX - 72 * dt);
      if (S.dogX <= target + 0.1) {
        S.catEaten = true; S.dogX = target; S.flash = 1;
        addFloat('ХВАТЬ!', S.catX + 4, RT - 6, 220, 1.1);
      }
    } else {
      S.lossDone += dt;
      if (S.lossDone > 1.25) { S.st = 'over'; S.overFor = 0; }
    }
    updateEffects(dt);
  }

  function update(dt) {
    dt = Math.min(dt, 0.06);
    S.t += dt;
    if (S.flash > 0) S.flash = Math.max(0, S.flash - dt * 2.4);

    switch (S.st) {
      case 'start':
        S.runTime += dt;
        break;
      case 'running': {
        S.runTime += dt; S.stageTime += dt;
        const i = clamp(S.level - 1, 0, MAX_STAGES - 1);
        S.speed = Math.min(MAX_SPEED[i], BASE_SPEED[i] + S.stageTime * RAMP[i]);
        S.stageDistance += S.speed * dt; S.distance += S.speed * dt;
        S.points += S.speed * dt * 1.25; S.score = Math.floor(S.points);
        if (S.invuln > 0) S.invuln = Math.max(0, S.invuln - dt);
        S.spawnIn -= dt;
        if (S.spawnIn <= 0) { spawnPuddle(); S.spawnIn = nextDelay(); }
        if (!S.onGround) {
          S.vy -= GRAVITY * dt; S.catY += S.vy * dt;
          if (S.catY <= 0) { S.catY = 0; S.vy = 0; S.onGround = true; dust(); }
        }
        for (const p of S.puddles) p.x -= S.speed * dt;
        S.puddles = S.puddles.filter(p => p.x + p.w > -6);
        collide();
        updateEffects(dt);
        if (S.st === 'running' && S.stageDistance >= S.stageGoal) {
          if (S.level < MAX_STAGES) beginStageClear(); else beginFinish();
        }
        break;
      }
      case 'stageclear':
        S.clearFor += dt; updateEffects(dt);
        if (S.clearFor >= 2.2) beginStage(2);
        break;
      case 'finish': updateFinish(dt); break;
      case 'dog': updateDog(dt); break;
      case 'victory': S.victoryFor += dt; updateEffects(dt, 0.6); break;
      case 'over': S.overFor += dt; updateEffects(dt, 0.6); break;
    }
  }

  // ---- отрисовка сцены ----
  function hash2(x, y) { let h = x * 374761393 + y * 668265263; h = (h ^ (h >> 13)) * 1274126177; return Math.abs(h ^ (h >> 16)); }

  function drawSky() {
    rect(0, 0, COLS, RT, 17);
    for (let y = 2; y < RT - 5; y++) {
      for (let x = 1; x < COLS - 1; x += 2) {
        const h = hash2(x + 7, y * 13 + 3);
        if (h % 23 !== 0) continue;
        const tw = 0.55 + 0.45 * Math.sin(S.t * 1.8 + (h % 628) / 100);
        const color = tw > 0.78 ? 255 : tw > 0.42 ? 246 : 239;
        chr(h % 7 === 0 ? '+' : '·', x, y, color);
      }
    }
    // луна: полумесяц
    const mx = px(COLS - 11) + CW / 2, my = py(4) + CH / 2, R = CH * 1.25;
    ctx.fillStyle = PAL[230];
    ctx.beginPath(); ctx.arc(mx, my, R, 0, Math.PI * 2); ctx.fill();
    ctx.fillStyle = PAL[17];
    ctx.beginPath(); ctx.arc(mx + R * 0.55, my - R * 0.3, R * 0.92, 0, Math.PI * 2); ctx.fill();
  }

  function drawSkyline() {
    const off = Math.floor(S.distance * 0.12), bw = 6;
    for (let x = 0; x < COLS; x++) {
      const wx = x + off, bi = Math.floor(wx / bw);
      const h = 1 + (hash2(bi, 7) % 3);
      const top = Math.max(1, RT - h);
      rect(x, top, 1, RT - top, 235);
      if (h >= 2 && wx % 2 === 0 && hash2(bi, 31) % 4 !== 0) chr('·', x, top + 1, 180);
    }
  }

  function drawRoad() {
    rect(0, RT, COLS, ROAD, 237);
    for (let x = 0; x < COLS; x++) chr('▁', x, RT, 240);
    const off = Math.floor(S.distance);
    for (let x = 0; x < COLS; x++) {
      if ((((x + off) % 16) + 16) % 16 < 7) chr('━', x, RT + 2, 178);
    }
  }

  function drawPuddles() {
    for (const p of S.puddles) {
      const x = Math.round(p.x);
      for (let i = 0; i < p.w; i++) {
        const cx = x + i;
        if (cx < -1 || cx > COLS) continue;
        chr('▄', cx, RT - 1, p.hit ? 24 : 33);
        chr('▀', cx, RT, p.hit ? 18 : 25);
        if (!p.hit && (i + Math.floor(S.t * 6)) % 4 === 0) chr('·', cx, RT - 1, 45);
      }
    }
  }

  function drawSausage() {
    if (S.st !== 'finish' || S.pillowShown) return;
    const x = Math.round(S.sausageX), w = Math.round(S.sausageW), y = RT - 1;
    rect(x, y + 1, w, 1, 240);           // тарелка
    rect(x + 1, y, w - 2, 2, 130);       // сосиска
    for (let i = 1; i < w - 1; i += 3) rect(x + i, y, 1, 1, 173);
  }

  function drawScene() {
    drawSky();
    drawSkyline();
    drawRoad();
    drawPuddles();
    drawSausage();
  }

  function drawCat() {
    if (S.catEaten) return;
    if (S.st !== 'over' && S.invuln > 0 && Math.floor(S.t * 16) % 2 === 0) return;
    const base = RT - Math.round(S.catY);
    let rows = CAT_IDLE;
    if (S.st === 'dog') rows = CAT_SCARED;
    else if (!S.onGround) rows = CAT_JUMP;
    else if (S.st === 'running') rows = (Math.floor(S.runTime * 9) % 2 === 0) ? CAT_IDLE : CAT_RUN_B;
    sprite(rows, Math.round(S.catX), base, ORANGE, false);
  }

  function drawDog() {
    if (S.st !== 'dog') return;
    const rows = Math.floor(S.lossFor * 9) % 2 === 0 ? DOG_A : DOG_B;
    sprite(rows, Math.round(S.dogX), RT, { body: 100, ear: 94, eye: 231, nose: 217 }, false);
  }

  function drawPillows() {
    const x = Math.round(S.pillowX);
    for (const px0 of [x, x + PILLOW_W + PILLOW_GAP]) {
      chr('▖', px0 + 1, RT - 2, 253);
      chr('▗', px0 + PILLOW_W - 2, RT - 2, 253);
      for (let i = 2; i < PILLOW_W - 2; i++) chr('▄', px0 + i, RT - 2, 253);
      for (let i = 0; i < PILLOW_W; i++) chr('█', px0 + i, RT - 1, 230);
      chr('▓', px0 + 3, RT - 1, 223);
      chr('▓', px0 + 4, RT - 1, 223);
      chr('▘', px0 + 1, RT, 181);
      chr('▝', px0 + PILLOW_W - 2, RT, 181);
      for (let i = 2; i < PILLOW_W - 2; i++) chr('▀', px0 + i, RT, 181);
    }
  }

  function drawBottle(x) {
    rect(x + 0.35, RT, 1.3, 0.18, 181);        // тень
    rect(x + 0.3, RT - 2, 1.4, 2, 255);        // корпус
    rect(x + 0.45, RT - 1.85, 0.25, 1.6, 253); // блик
    rect(x + 0.75, RT - 3, 0.5, 1, 255);       // горлышко
    rect(x + 0.6, RT - 3.9, 0.8, 0.9, 195);    // крышка
  }

  function drawBottles() {
    const x = Math.round(S.pillowX);
    drawBottle(x - 4);
    drawBottle(x + PILLOW_W * 2 + PILLOW_GAP + 1);
  }

  function drawSleeping(x, pal) {
    const base = RT - 3;
    sprite(CAT_SLEEP, x, base, pal, false);
    // подгузник
    chr('▖', x + 2, RT - 4, 195);
    for (let i = 3; i <= 5; i++) chr('▄', x + i, RT - 4, 255);
    chr('▗', x + 6, RT - 4, 195);
    chr('▐', x + 2, RT - 3, 195);
    chr('█', x + 3, RT - 3, 255);
    chr('▓', x + 4, RT - 3, 195);
    chr('█', x + 5, RT - 3, 255);
    chr('▌', x + 6, RT - 3, 195);
    // шапочка: конус и помпон
    rect(x + 1, RT - 7, 7, 1, 218);
    rect(x + 3, RT - 8, 3, 1, 218);
    circle(x + 4.5, RT - 9, CH * 0.3, 255);
    // сосочка
    chr('o', x + 4, RT - 5, 218);
  }

  function drawFinale() {
    if (!S.pillowShown) return;
    drawPillows();
    drawBottles();
    if (S.asleep) {
      drawSleeping(Math.round(S.cat2X), GRAY);
      drawSleeping(Math.round(S.catX), ORANGE);
      for (const cx of [Math.round(S.catX), Math.round(S.cat2X)]) {
        for (let i = 0; i < 2; i++) {
          const zx = cx + 8 + i;
          const zy = RT - 8 - i * 2 - (Math.floor(S.t * 2) % 2);
          chr('z', zx, zy, 246);
        }
      }
      const hx = Math.round(S.pillowX) + PILLOW_W + 1;
      for (let i = 0; i < 2; i++) chr('♥', hx + i * 2, RT - 8 - i, 197);
    } else {
      drawCat();
    }
  }

  function drawFloats() {
    for (const f of S.floats) {
      const a = Math.min(1, f.life / f.max);
      ctx.globalAlpha = a;
      label(f.text, Math.round(f.x), Math.round(f.y), f.color, Math.floor(FONT * 0.9));
      ctx.globalAlpha = 1;
    }
  }

  function drawParticles() {
    for (const p of S.particles) chr(p.ch, Math.round(p.x), Math.round(p.y), p.color);
  }

  function drawHud() {
    label('ОЧКИ ' + S.score, 3, 1, 231, Math.floor(FONT * 1.0));
    for (let i = 0; i < S.lives; i++) chr('♥', 36 + i * 2, 1, 197);
    label('ПРИЗ', COLS - 4, 1, 231, Math.floor(FONT * 0.9));
    label('ДО СОСИСКИ', 16, 3, 246, Math.floor(FONT * 0.8));
    const total = GOALS[0] + GOALS[1];
    const done = Math.min(1, S.distance / total);
    const bx = 30, bw = 34;
    rect(bx, 3.25, bw, 0.5, 238);
    rect(bx, 3.25, Math.max(0, bw * done), 0.5, 220);
    label(Math.round(done * 100) + '%', bx + bw + 4, 3, 246, Math.floor(FONT * 0.8));
  }

  function panel(lines) {
    const w = Math.min(COLS - 6, 52), h = lines.length + 4;
    const x = Math.round((COLS - w) / 2), y = Math.round((RT - h) / 2);
    rect(x, y, w, h, 236);
    rect(x + 1, y + 1, w - 2, h - 2, 232);
    for (let i = 0; i < lines.length; i++) {
      const [text, color, size] = lines[i];
      label(text, COLS / 2, y + 2 + i, color, size || Math.floor(FONT * 1.05));
    }
  }

  function drawOverlay() {
    if (S.st === 'start') {
      panel([
        ['КОТИК-БЕГУН', 220, Math.floor(FONT * 1.6)],
        ['', 231],
        ['Коснитесь экрана — прыжок', 253],
        ['Два этапа, три жизни, сосиска', 246],
        ['', 231],
        ['КОСНИТЕСЬ, ЧТОБЫ НАЧАТЬ', 220, Math.floor(FONT * 1.0)],
      ]);
    } else if (S.st === 'stageclear') {
      panel([
        ['ЭТАП ' + S.level + ' ПРОЙДЕН', 220, Math.floor(FONT * 1.4)],
        ['', 231],
        ['+100 очков, дальше быстрее', 253],
      ]);
    } else if (S.st === 'victory') {
      panel([
        ['СОСИСКА СЪЕДЕНА!', 220, Math.floor(FONT * 1.4)],
        ['', 231],
        ['Очки: ' + S.score + (S.newBest ? '   НОВЫЙ РЕКОРД!' : '   Рекорд: ' + S.best), 253],
        ['', 231],
        ['Коснитесь, чтобы сыграть ещё', 246],
      ]);
    } else if (S.st === 'over' && S.overFor > 0.6) {
      panel([
        ['КОТИКА СЪЕЛИ', 203, Math.floor(FONT * 1.4)],
        ['', 231],
        ['Очки: ' + S.score + '   Рекорд: ' + S.best, 253],
        ['', 231],
        ['Коснитесь, чтобы сыграть ещё', 246],
      ]);
    }
  }

  function render() {
    ctx.fillStyle = '#070722';
    ctx.fillRect(0, 0, canvas.width / DPR, canvas.height / DPR);
    rect(0, 0, COLS, ROWS, 17);
    drawScene();
    if (S.flash > 0) {
      ctx.globalAlpha = S.flash * 0.25;
      rect(0, 0, COLS, ROWS, 9);
      ctx.globalAlpha = 1;
    }
    if (S.st === 'finish' && S.pillowShown) drawFinale();
    else { drawCat(); drawDog(); }
    drawParticles();
    drawFloats();
    drawHud();
    drawOverlay();
  }

  // ---- ввод ----
  function tap() {
    if (S.st === 'start') { startGame(); return; }
    if (S.st === 'running') { jump(); return; }
    if (S.st === 'victory' || (S.st === 'over' && S.overFor > 0.6)) { reset(); startGame(); return; }
    if (S.st === 'finish' && !S.asleep) { S.sleepFor += 0.15; }
  }
  canvas.addEventListener('pointerdown', (e) => { e.preventDefault(); tap(); });
  window.addEventListener('keydown', (e) => {
    if (e.code === 'Space' || e.code === 'Enter' || e.code === 'ArrowUp' || e.code === 'KeyW') { e.preventDefault(); tap(); }
  });
  document.addEventListener('gesturestart', e => e.preventDefault());
  document.addEventListener('dblclick', e => e.preventDefault());

  // ---- цикл ----
  let last = 0;
  function frame(ts) {
    if (!last) last = ts;
    let dt = (ts - last) / 1000;
    last = ts;
    if (dt > 0.25) dt = 0.25;
    update(dt);
    render();
    requestAnimationFrame(frame);
  }

  resize();
  requestAnimationFrame(frame);

  // отладочный доступ к состоянию (для тестов и превью)
  window.__catRunner = S;
})();
