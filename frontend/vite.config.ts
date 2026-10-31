import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    babel({ presets: [reactCompilerPreset()] })
  ],
  server: {
    // Запросы к API идут через тот же origin, иначе браузер не сохранит куки (разные "сайты")
    proxy: {
      '/main/api': {
        target: 'http://192.168.0.30:30080',
        changeOrigin: true,
      },
    },
  },
})