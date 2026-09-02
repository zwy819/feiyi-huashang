#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""一次性入库脚本：把 8 个压缩后的 GLB 模型登记到 models 表并关联 garments"""
import os
import shutil
import sqlite3

SERVER_DIR = os.path.dirname(os.path.abspath(__file__))
DB_PATH = os.path.join(SERVER_DIR, 'data', 'nonlegacy.db')
SRC_DIR = os.path.normpath(os.path.join(SERVER_DIR, '..', 'client', 'public', 'models'))
DST_ROOT = os.path.join(SERVER_DIR, 'data', 'models')

# garment_id -> 模型名（不含 .glb）
MODELS = {
    179: '哈萨克族绣花大衣',
    182: '瑶族刺绣盘王衣',
    183: '土家族织锦衣裙',
    184: '苗族花带裙',
    185: '蒙古族长袍',
    188: '汉族缂丝吉服',
    189: '汉族杭绣礼服',
    190: '畲族凤凰装',
}

conn = sqlite3.connect(DB_PATH)
cur = conn.cursor()

for gid, name in sorted(MODELS.items()):
    fname = name + '.glb'
    src = os.path.join(SRC_DIR, fname)
    if not os.path.exists(src):
        print(f'[跳过] {name}: 源文件不存在 {src}')
        continue

    file_size = os.path.getsize(src)
    dst_dir = os.path.join(DST_ROOT, str(gid))
    os.makedirs(dst_dir, exist_ok=True)
    dst = os.path.join(dst_dir, fname)
    shutil.copy2(src, dst)

    file_path = 'data/models/{}/{}.glb'.format(gid, name)

    cur.execute('SELECT model_id FROM garments WHERE id = ?', (gid,))
    row = cur.fetchone()
    existing_mid = row[0] if row else None

    if existing_mid:
        # 记录旧 file_path，用于清理错误残留文件
        cur.execute('SELECT file_path FROM models WHERE id = ?', (existing_mid,))
        old = cur.fetchone()
        old_path = old[0] if old else ''
        cur.execute(
            'UPDATE models SET filename=?, original_name=?, file_size=?, file_path=?, garment_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?',
            (fname, fname, file_size, file_path, gid, existing_mid),
        )
        # 清理旧文件（如果指向不同文件名）
        if old_path and old_path != file_path:
            old_abs = os.path.join(SERVER_DIR, old_path.replace('\\', os.sep))
            if os.path.exists(old_abs) and os.path.basename(old_abs) != fname:
                try:
                    os.remove(old_abs)
                    print(f'  [清理] 删除旧文件: {old_path}')
                except OSError as e:
                    print(f'  [警告] 清理旧文件失败: {e}')
        print(f'[更新] {name} (garment {gid}): model {existing_mid} -> {fname} ({file_size} bytes)')
    else:
        cur.execute(
            "INSERT INTO models (filename, original_name, file_size, mime_type, file_path, garment_id) VALUES (?, ?, ?, 'model/gltf-binary', ?, ?)",
            (fname, fname, file_size, file_path, gid),
        )
        mid = cur.lastrowid
        cur.execute('UPDATE garments SET model_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?', (mid, gid))
        print(f'[新增] {name} (garment {gid}): model {mid} ({file_size} bytes)')

conn.commit()
conn.close()
print('完成')
