/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,vue}',
  ],
  theme: {
    extend: {
      colors: {
        // ===== 莫高窟色系 (Mogao Grottoes Palette) =====
        mogao: {
          // 主色 — 朱砂红
          cinnabar:   { DEFAULT: '#C43A31', 50: '#FBE9E7', 100: '#F5CBC7', 200: '#E9988F', 300: '#DC6E63', 400: '#D05045', 500: '#C43A31', 600: '#A42E27', 700: '#83231E', 800: '#651A16', 900: '#4A120F' },
          // 石绿 — 壁画中常用的矿物绿
          malachite:  { DEFAULT: '#2E7D5A', 50: '#E8F3ED', 100: '#C5E2D3', 200: '#9DCFB6', 300: '#74BB9A', 400: '#54AB83', 500: '#2E7D5A', 600: '#266848', 700: '#1F5339', 800: '#17402C', 900: '#0F2E1F' },
          // 土黄 — 敦煌大漠底色
          ochre:      { DEFAULT: '#C89B5E', 50: '#FBF5EF', 100: '#F5E8D8', 200: '#ECD2B5', 300: '#E0BB8B', 400: '#D4A96F', 500: '#C89B5E', 600: '#B0804A', 700: '#8F6639', 800: '#6F4E2B', 900: '#52391E' },
          // 青金石蓝 — 壁画中贵重的青金石颜料
          lapis:      { DEFAULT: '#2E4A7A', 50: '#E8ECF3', 100: '#C6D0E2', 200: '#9EAECF', 300: '#768CB9', 400: '#5071A5', 500: '#2E4A7A', 600: '#253D65', 700: '#1D3051', 800: '#15243E', 900: '#0E192C' },
          // 铅白 — 壁画底色
          leadwhite:  { DEFAULT: '#F5F0E8', 50: '#FEFDFC', 100: '#FCFAF7', 200: '#FAF5EF', 300: '#F7F0E5', 400: '#F5EBD8', 500: '#F5F0E8', 600: '#D4C9B8', 700: '#AFA290', 800: '#8A7E6E', 900: '#665D52' },
          // 金色 — 贴金/描金
          gold:       { DEFAULT: '#B8862D', 50: '#F9F2E4', 100: '#F0DEBC', 200: '#E3C68A', 300: '#D4AD5E', 400: '#C69A40', 500: '#B8862D', 600: '#9A7025', 700: '#7C5A1D', 800: '#5F4516', 900: '#453210' },
        },
      },
      fontFamily: {
        // 中文衬线：适合标题的中国风字体
        serif: ['"Noto Serif SC"', '"Source Han Serif SC"', '"Songti SC"', 'STSong', 'serif'],
        // 中文无衬线：正文
        sans: ['"Noto Sans SC"', '"Source Han Sans SC"', '"PingFang SC"', '"Microsoft YaHei"', 'sans-serif'],
      },
      backgroundImage: {
        'gradient-mogao': 'linear-gradient(135deg, #C43A31 0%, #2E7D5A 100%)',
        'gradient-paper': 'linear-gradient(to bottom, #FCFAF7 0%, #F5F0E8 100%)',
      },
      boxShadow: {
        'mogao': '0 4px 24px rgba(196, 58, 49, 0.12)',
        'mogao-lg': '0 8px 40px rgba(196, 58, 49, 0.18)',
      },
      borderRadius: {
        'mogao': '2px',
      },
    },
  },
  plugins: [],
};
