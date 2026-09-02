/**
 * GLB 模型优化脚本
 *
 * 完整优化管线，将 Blender 直出的 80MB+ 模型压缩到 5-15MB。
 * 每步都有详细注释说明其作用和原理。
 *
 * ===== 优化管线概览 =====
 *
 * 1. 分析阶段     — 读取模型，统计顶点/三角形/纹理
 * 2. dedup()      — 去重：合并重复的 accessor（bufferView 引用去重）
 * 3. weld()       — 焊接：合并空间位置极近的顶点，减少冗余
 * 4. simplify()   — 减面：最关键的一步！用边缘坍缩算法减三角形数
 * 5. flatten()    — 展平：移除多余的中间节点，简化场景图
 * 6. join()       — 合并：将同材质的 mesh 合并为单个 mesh
 * 7. textureCompress() — 纹理压缩：PNG → JPEG + 缩小尺寸 + 降质
 * 8. prune()      — 清理：移除不再被引用的材质/纹理/buffer
 *
 * ===== 运行 =====
 *   cd tools
 *   npm install
 *   node optimize.js
 *
 * ===== 参数调整说明 =====
 *
 * simplifyRatio: 保留的三角形比例
 *   1.0  = 不减面
 *   0.5  = 减到 50%
 *   0.08 = 减到 8%（推荐起点，适合 150 万面 → 12 万面）
 *   增大 → 面数更多，细节更丰富，文件更大
 *   减小 → 面数更少，细节损失大，文件更小
 *
 * textureQuality: JPEG 质量 (1-100)
 *   95  = 近乎无损
 *   80  = 推荐值，肉眼几乎看不出区别
 *   60  = 有轻微色块，文件极小
 *
 * textureMaxSize: 纹理最大边长 (px)
 *   4096 = 4K，细节丰富
 *   2048 = 推荐值，适合 web
 *   1024 = 适合移动端
 */

import { readFileSync, writeFileSync, existsSync, renameSync, mkdirSync } from 'fs';
import { resolve, dirname } from 'path';
import { NodeIO } from '@gltf-transform/core';
import { ALL_EXTENSIONS } from '@gltf-transform/extensions';
import {
  dedup,
  flatten,
  join,
  weld,
  simplify,
  prune,
  textureCompress,
} from '@gltf-transform/functions';
import { MeshoptSimplifier } from 'meshoptimizer';
import sharp from 'sharp';

// ============================================================
// 用户可调参数 — 根据效果调整这里
// ============================================================
const OPTIONS = {
  /** 保留的三角形比例 (0-1)，越小越激进 */
  simplifyRatio: 0.08,
  /** 减面误差容忍度，越小越精确越慢 */
  simplifyError: 0.001,
  /** 纹理 JPEG 质量 (1-100) */
  textureQuality: 80,
  /** 纹理最大边长 (px) */
  textureMaxSize: 2048,
  /** 是否备份原始文件（*.glb.backup） */
  backup: true,
};

// ============================================================
// 模型路径
// ============================================================
const MODELS_DIR = resolve('../client/public/models');
const MODELS = [
  '哈萨克族绣花大衣.glb',
  '土家族织锦衣裙.glb',
  '汉族杭绣礼服.glb',
  '汉族缂丝吉服.glb',
  '瑶族刺绣盘王衣.glb',
  '畲族凤凰装.glb',
  '苗族花带裙.glb',
  '蒙古族长袍.glb',
];

// ============================================================
// 主流程
// ============================================================
const io = new NodeIO().registerExtensions(ALL_EXTENSIONS);

// Meshoptimizer 需要先初始化 WASM 二进制
console.log('初始化 meshoptimizer WASM...');
await MeshoptSimplifier.ready;
console.log('就绪\n');

if (!existsSync(MODELS_DIR)) {
  mkdirSync(MODELS_DIR, { recursive: true });
  console.log(`创建模型目录: ${MODELS_DIR}`);
}

for (const modelFile of MODELS) {
  const inputPath = resolve(MODELS_DIR, modelFile);

  if (!existsSync(inputPath)) {
    console.warn(`⚠ 跳过不存在的文件: ${modelFile}`);
    continue;
  }

  console.log(`${'='.repeat(55)}`);
  console.log(`  ${modelFile}`);
  console.log(`${'='.repeat(55)}`);

  const inputSize = readFileSync(inputPath).length;
  console.log(`原始: ${(inputSize / 1024 / 1024).toFixed(2)} MB`);

  try {
    // ===========================
    // 步骤 1：读取模型
    // ===========================
    process.stdout.write('读取模型... ');
    const doc = await io.read(inputPath);
    const root = doc.getRoot();
    console.log('OK');

    // ----- 统计原始数据 -----
    let origVerts = 0, origTris = 0;
    for (const mesh of root.listMeshes()) {
      for (const prim of mesh.listPrimitives()) {
        const pos = prim.getAttribute('POSITION');
        const idx = prim.getIndices();
        if (pos) origVerts += pos.getCount();
        if (idx) origTris += idx.getCount() / 3;
      }
    }
    console.log(`  原始: ${Math.round(origTris).toLocaleString()} 三角形, ` +
                `${origVerts.toLocaleString()} 顶点`);

    // ===========================
    // 步骤 2：dedup — 去重 accessor
    // ===========================
    // glTF 规范中多个 bufferView 可能指向 buffer 中的同一段数据。
    // dedup 找到完全相同的 accessor 并合并引用，减少冗余元数据。
    // 不影响渲染结果，纯粹减小文件体积。
    process.stdout.write('dedup  ... ');
    await doc.transform(dedup());
    console.log('OK');

    // ===========================
    // 步骤 3：weld — 焊接临近顶点
    // ===========================
    // 3D 建模软件导出时，同一空间位置的顶点常被拆成多个顶点
    //（如 UV 接缝、法线硬边）。weld 在容差范围内合并这些顶点。
    // tolerance: 世界空间单位，值越小越保守。典型值 0.0001-0.001
    process.stdout.write('weld   ... ');
    await doc.transform(weld({ tolerance: 0.0001 }));
    console.log('OK');

    // ===========================
    // 步骤 4：simplify — 减面（最核心步骤）
    // ===========================
    // 使用 meshoptimizer（基于 Quadric Error Metrics 的边缘坍缩算法）
    // 逐步移除对视觉影响最小的边，在保持外形的前提下大幅减少三角形。
    //
    // ratio:  最终三角形数 / 原始三角形数
    // error:  允许的最大误差（世界单位），越小越精确
    // lockBorder: true 时保持开放边界的形状不变
    //
    // 你的模型 150 万面 → 简化到 12 万面 (8%)，肉眼几乎看不出区别，
    // 因为高面数区域 (光滑曲面) 的面被优先合并，锐利边角的形状被保留。
    process.stdout.write(`simplify(r=${OPTIONS.simplifyRatio}) ... `);
    console.log('(需要 1-5 分钟)');
    await doc.transform(
      simplify({
        simplifier: MeshoptSimplifier,
        ratio: OPTIONS.simplifyRatio,
        error: OPTIONS.simplifyError,
      })
    );
    console.log('  simplify OK');

    // ===========================
    // 步骤 5：flatten — 展平场景图
    // ===========================
    // Blender 导出的 GLB 常有深层嵌套的节点树（如 Empty → Mesh → ...）。
    // flatten 把子节点变换 bake 到网格数据中并移除中间节点，
    // 减少文件中的 node/matrix 数量。不影响渲染结果。
    process.stdout.write('flatten ... ');
    await doc.transform(flatten());
    console.log('OK');

    // ===========================
    // 步骤 6：join — 合并同材质 mesh
    // ===========================
    // 如果多个 mesh 使用相同材质，join 把它们合并成一个。
    // 减少 draw call 数量 → 渲染更快 → 也略微减小文件。
    process.stdout.write('join    ... ');
    await doc.transform(join());
    console.log('OK');

    // ===========================
    // 步骤 7：textureCompress — 纹理压缩
    // ===========================
    // 你的模型纹理占 45MB（3 张 PNG）。优化策略：
    //
    //   a) 格式转换：PNG（无损）→ JPEG（有损）
    //      → 体积通常减到 5-15%
    //
    //   b) 尺寸缩小：原始纹理可能 4096×4096，resize 到 2048×2048
    //      → 像素只有 1/4，体积进一步减到 25%
    //
    //   c) 质量问题：普通凹凸纹理的正常贴图用 80 质量看不出区别；
    //      metallicRoughness 贴图可以更低 (60) 因为人眼对其不敏感
    //
    // encoder: 'sharp' — 使用 libvips (C 库) 极速编码
    process.stdout.write('texture compress (JPEG, q=80, 2048px) ... ');
    await doc.transform(
      textureCompress({
        encoder: sharp,
        targetFormat: 'jpeg',
        resize: [OPTIONS.textureMaxSize, OPTIONS.textureMaxSize],
        quality: OPTIONS.textureQuality,
      })
    );
    console.log('OK');

    // ===========================
    // 步骤 8：prune — 清理游离资源
    // ===========================
    // 前面步骤可能导致某些材质、纹理、accessor 不再被任何 mesh 引用。
    // prune 将其从文档中移除，进一步减小文件。
    process.stdout.write('prune   ... ');
    await doc.transform(prune());
    console.log('OK');

    // ===========================
    // 优化结果统计
    // ===========================
    let optVerts = 0, optTris = 0;
    for (const mesh of root.listMeshes()) {
      for (const prim of mesh.listPrimitives()) {
        const pos = prim.getAttribute('POSITION');
        const idx = prim.getIndices();
        if (pos) optVerts += pos.getCount();
        if (idx) optTris += idx.getCount() / 3;
      }
    }
    console.log(`  优化后: ${Math.round(optTris).toLocaleString()} 三角形, ` +
                `${optVerts.toLocaleString()} 顶点`);
    console.log(`  减面率: ${((1 - optTris / origTris) * 100).toFixed(1)}%`);

    // ===========================
    // 写入文件
    // ===========================
    const outputPath = inputPath.replace('.glb', '_optimized.glb');
    process.stdout.write(`写入 ${outputPath.split('/').pop()} ... `);
    const arrayBuffer = await io.writeBinary(doc);
    writeFileSync(outputPath, Buffer.from(arrayBuffer));
    const outputSize = readFileSync(outputPath).length;
    console.log(`OK (${(outputSize / 1024 / 1024).toFixed(2)} MB)`);

    const reduction = ((1 - outputSize / inputSize) * 100).toFixed(1);
    console.log(`\n  ${(inputSize / 1024 / 1024).toFixed(1)} MB → ${(outputSize / 1024 / 1024).toFixed(1)} MB  (${reduction}%)`);

    // ----- 备份 + 替换 -----
    if (OPTIONS.backup) {
      const backupPath = inputPath + '.backup';
      renameSync(inputPath, backupPath);
      renameSync(outputPath, inputPath);
      console.log(`  原始备份: ${modelFile}.backup`);
      console.log(`  已替换:   ${modelFile}`);
    }

  } catch (err) {
    console.error(`\n  ❌ 失败: ${err.message}`);
    console.error(err);
  }
}

console.log(`\n${'='.repeat(55)}`);
console.log('全部完成。');
console.log(`\n还原: rename *.glb.backup → *.glb`);
console.log(`调整: 修改 OPTIONS.simplifyRatio (当前 ${OPTIONS.simplifyRatio})`);
