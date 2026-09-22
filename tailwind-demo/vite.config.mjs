import {defineConfig} from 'vite';
import react from '@vitejs/plugin-react';
import tailwind from '@tailwindcss/vite';
import jev from '../integration/jev-vite.mjs';
export default defineConfig({plugins:[react(),tailwind(),jev()]});
