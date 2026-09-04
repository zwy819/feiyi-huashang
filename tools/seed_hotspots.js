/**
 * seed_hotspots.js — 计算示例热点标注的模型表面坐标
 *
 * 用 @gltf-transform 读取 GLB，重建世界空间三角形网格后，
 * 从 +Z/-Z 方向对指定高度做射线求交，得到可直接写入 seed 的表面点。
 *
 * 用法：cd tools && node seed_hotspots.js
 * 输出：JSON（每件服饰若干个表面建议点）
 */
import { NodeIO } from '@gltf-transform/core'

const TARGETS = [
  { file: '../client/public/models/苗族百褶裙.glb', heights: [0.75, 0.45, 0.1], name: '苗族百褶裙' },
  { file: '../client/public/models/苏绣旗袍.glb', heights: [0.85, 0.55, 0.2], name: '苏绣旗袍' },
  { file: '../client/public/models/藏族氆氇袍.glb', heights: [0.8, 0.45, 0.15], name: '藏族氆氇袍' },
]

// ---------- 精简 vec3 / mat4 工具 ----------
const v3 = (x = 0, y = 0, z = 0) => [x, y, z]
const sub = (a, b) => [a[0] - b[0], a[1] - b[1], a[2] - b[2]]
const cross = (a, b) => [
  a[1] * b[2] - a[2] * b[1],
  a[2] * b[0] - a[0] * b[2],
  a[0] * b[1] - a[1] * b[0],
]
const dot = (a, b) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2]

// 列主序 4x4：平移 * 旋转(四元数) * 缩放
function composeTRS(t, q, s) {
  const [x, y, z, w] = q
  const x2 = x + x, y2 = y + y, z2 = z + z
  const xx = x * x2, xy = x * y2, xz = x * z2
  const yy = y * y2, yz = y * z2, zz = z * z2
  const wx = w * x2, wy = w * y2, wz = w * z2
  return [
    (1 - (yy + zz)) * s[0], (xy + wz) * s[0], (xz - wy) * s[0], 0,
    (xy - wz) * s[1], (1 - (xx + zz)) * s[1], (yz + wx) * s[1], 0,
    (xz + wy) * s[2], (yz - wx) * s[2], (1 - (xx + yy)) * s[2], 0,
    t[0], t[1], t[2], 1,
  ]
}
function mul4(a, b) {
  const out = new Array(16).fill(0)
  for (let c = 0; c < 4; c++)
    for (let r = 0; r < 4; r++)
      for (let k = 0; k < 4; k++) out[c * 4 + r] += a[k * 4 + r] * b[c * 4 + k]
  return out
}
function transformPoint(m, p) {
  return [
    m[0] * p[0] + m[4] * p[1] + m[8] * p[2] + m[12],
    m[1] * p[0] + m[5] * p[1] + m[9] * p[2] + m[13],
    m[2] * p[0] + m[6] * p[1] + m[10] * p[2] + m[14],
  ]
}
const I4 = [1,0,0,0, 0,1,0,0, 0,0,1,0, 0,0,0,1]

// Möller–Trumbore 射线-三角形求交，返回 t 或 -1
function rayTriangle(orig, dir, tri) {
  const [v0, v1, v2] = tri
  const e1 = sub(v1, v0)
  const e2 = sub(v2, v0)
  const p = cross(dir, e2)
  const det = dot(e1, p)
  if (Math.abs(det) < 1e-9) return -1
  const invDet = 1 / det
  const tvec = sub(orig, v0)
  const u = dot(tvec, p) * invDet
  if (u < 0 || u > 1) return -1
  const q = cross(tvec, e1)
  const v = dot(dir, q) * invDet
  if (v < 0 || u + v > 1) return -1
  const t = dot(e2, q) * invDet
  return t > 1e-6 ? t : -1
}

// ---------- GLB → 世界空间三角形列表 ----------
async function loadTriangles(path) {
  const io = new NodeIO()
  const doc = await io.read(path)
  const tris = []

  for (const scene of doc.getRoot().listScenes()) {
    const walk = (node, parentMat) => {
      const world = mul4(parentMat, composeTRS(node.getTranslation(), node.getRotation(), node.getScale()))
      const mesh = node.getMesh()
      if (mesh) {
        for (const prim of mesh.listPrimitives()) {
          const pos = prim.getAttribute('POSITION')
          if (!pos) continue
          const idx = prim.getIndices()
          const count = idx ? idx.getCount() : pos.getCount()
          const el = [0, 0, 0]
          const pt = (i) => transformPoint(world, pos.getElement(i, el))
          for (let i = 0; i + 2 < count; i += 3) {
            const ia = idx ? idx.getScalar(i) : i
            const ib = idx ? idx.getScalar(i + 1) : i + 1
            const ic = idx ? idx.getScalar(i + 2) : i + 2
            tris.push([pt(ia), pt(ib), pt(ic)])
          }
        }
      }
      for (const child of node.listChildren()) walk(child, world)
    }
    for (const node of scene.listChildren()) walk(node, I4)
  }
  return tris
}

function computeBBox(tris) {
  const min = v3(Infinity, Infinity, Infinity)
  const max = v3(-Infinity, -Infinity, -Infinity)
  for (const tri of tris)
    for (const v of tri)
      for (let i = 0; i < 3; i++) {
        min[i] = Math.min(min[i], v[i])
        max[i] = Math.max(max[i], v[i])
      }
  return { min, max }
}

// 与前端 ModelLoader 一致：模型平移 -center 居中
function centerTriangles(tris, center) {
  for (const tri of tris) for (const v of tri) {
    v[0] -= center[0]; v[1] -= center[1]; v[2] -= center[2]
  }
}

// 从 +Z 优先（初始相机方向）找指定高度的表面点
function surfacePointAtY(tris, y) {
  const radius = 50
  for (const dirZ of [1, -1]) {
    const orig = v3(0, y, dirZ * radius)
    const dir = v3(0, 0, -dirZ)
    let bestT = Infinity
    for (const tri of tris) {
      const t = rayTriangle(orig, dir, tri)
      if (t > 0 && t < bestT) bestT = t
    }
    if (bestT < Infinity) {
      const p = [orig[0] + dir[0] * bestT, orig[1] + dir[1] * bestT, orig[2] + dir[2] * bestT]
      return p
    }
  }
  return null
}

const result = {}
for (const target of TARGETS) {
  const tris = await loadTriangles(target.file)
  if (!tris.length) {
    console.error(`[skip] ${target.file}: 无三角形`)
    continue
  }
  const { min, max } = computeBBox(tris)
  const center = [(min[0] + max[0]) / 2, (min[1] + max[1]) / 2, (min[2] + max[2]) / 2]
  centerTriangles(tris, center)
  const size = sub(max, min)

  const points = []
  for (const ratio of target.heights) {
    const y = -size[1] / 2 + size[1] * ratio
    const p = surfacePointAtY(tris, y)
    points.push({
      ratio,
      y: Number(y.toFixed(4)),
      position: p ? p.map((n) => Number(n.toFixed(4))) : null,
    })
  }
  result[target.name] = {
    triangles: tris.length,
    size: size.map((n) => Number(n.toFixed(3))),
    points,
  }
}

console.log(JSON.stringify(result, null, 2))
