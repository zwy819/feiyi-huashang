/**
 * GLB 模型分析工具
 *
 * 读取 GLB 文件，输出详细的资源使用报告：
 * - 顶点/三角形数量、网格数量
 * - 纹理数量、格式、大小
 * - 预估面数 + 纹理优化后的体积
 *
 * 运行: cd tools && npm install && node analyze.js
 */

import { readFileSync } from 'fs';
import { resolve } from 'path';
import { NodeIO } from '@gltf-transform/core';
import { ALL_EXTENSIONS } from '@gltf-transform/extensions';

const MODELS_DIR = resolve('../client/public/models');
const MODELS = [
  '回族绣花盖头.glb',
  '广东潮绣婚服.glb',
  '朝鲜族短袄裙.glb',
  '湖北土家族摆手舞衣.glb',
  '满族旗袍.glb',
  '福建惠安女服饰.glb',
  '藏族安多袍.glb',
  '裕固族红缨帽.glb',
  '赫哲族鱼皮衣.glb',
  '陕西肚兜.glb',
  '高山族贝珠衣.glb',
  '黎族筒裙.glb',
];

const io = new NodeIO().registerExtensions(ALL_EXTENSIONS);

for (const modelFile of MODELS) {
  const inputPath = resolve(MODELS_DIR, modelFile);
  if (!readFileSync(inputPath, { flag: 'r' }).length) {
    console.warn(`⚠ 跳过不存在: ${modelFile}`);
    continue;
  }

  console.log(`\n${'='.repeat(55)}`);
  console.log(`  ${modelFile}`);
  console.log(`${'='.repeat(55)}`);

  // ----- 文件大小 -----
  const fileSize = readFileSync(inputPath).length;
  console.log(`  文件大小   ${(fileSize / 1024 / 1024).toFixed(2)} MB`);

  // ----- 模型结构 -----
  const doc = await io.read(inputPath);
  const root = doc.getRoot();

  const meshes = root.listMeshes();
  let totalVerts = 0;
  let totalTris = 0;
  let totalIndices = 0;

  for (const mesh of meshes) {
    for (const prim of mesh.listPrimitives()) {
      const pos = prim.getAttribute('POSITION');
      const indices = prim.getIndices();
      if (pos) totalVerts += pos.getCount();
      if (indices) {
        const idxCnt = indices.getCount();
        totalIndices += idxCnt;
        totalTris += idxCnt / 3;
      }
    }
  }

  console.log(`  网格数     ${meshes.length}`);
  console.log(`  顶点数     ${totalVerts.toLocaleString()}`);
  console.log(`  三角形     ${Math.round(totalTris).toLocaleString()}`);
  console.log(`  索引数     ${totalIndices.toLocaleString()}`);

  // ----- 纹理 -----
  const textures = root.listTextures();
  console.log(`  纹理数     ${textures.length}`);
  let texTotal = 0;
  for (const tex of textures) {
    const name = tex.getName() || 'unnamed';
    const img = tex.getImage();
    const size = img ? img.byteLength : 0;
    const mime = tex.getMimeType();
    texTotal += size;
    console.log(`    - ${name}: ${(size / 1024 / 1024).toFixed(2)} MB (${mime})`);
  }
  console.log(`  纹理合计   ${(texTotal / 1024 / 1024).toFixed(2)} MB (占 ${((texTotal / fileSize) * 100).toFixed(0)}%)`);
  console.log(`  几何数据   ${((fileSize - texTotal) / 1024 / 1024).toFixed(2)} MB (占 ${(((fileSize - texTotal) / fileSize) * 100).toFixed(0)}%)`);

  // ----- 预估优化后 -----
  const estGeoTarget = Math.min(totalTris, 100000);
  const geoRatio = estGeoTarget / totalTris;
  const estGeoSize = ((fileSize - texTotal) * geoRatio * 0.8).toFixed(1); // ~80% 额外效率
  // 纹理 JPEG 压缩 + 缩小 → 约原大小 5-15%
  const estTexJpeg = (texTotal * 0.08).toFixed(1);
  const estTotal = (Number(estGeoSize) + Number(estTexJpeg)).toFixed(1);

  console.log(`\n  预估优化后`);
  console.log(`    目标面数     ${Math.round(estGeoTarget).toLocaleString()} (${(geoRatio * 100).toFixed(0)}%)`);
  console.log(`    几何体       ~${estGeoSize} MB`);
  console.log(`    纹理(JPEG)   ~${estTexJpeg} MB`);
  console.log(`    ** 合计 **   ~${estTotal} MB`);
}

console.log(`\n运行 'node optimize.js' 执行优化。\n`);
