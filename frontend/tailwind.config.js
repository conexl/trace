/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        canvas: '#050505',
        surface: '#0A0A0A',
        'surface-elevated': '#111111',
        border: {
          DEFAULT: 'rgba(255, 255, 255, 0.10)',
          focus: 'rgba(255, 255, 255, 0.26)',
          glow: 'rgba(255, 255, 255, 0.22)',
        },
        accent: {
          DEFAULT: '#FFFFFF',
          muted: 'rgba(255, 255, 255, 0.07)',
          glow: 'rgba(255, 255, 255, 0.14)',
          strong: 'rgba(255, 255, 255, 0.28)',
        },
        muted: {
          DEFAULT: '#8A8A8A',
          soft: '#C7C7C7',
        },
        active: '#FFFFFF',
        amber: {
          muted: '#A3A3A3',
          soft: '#D4D4D4',
        },
      },
      fontFamily: {
        sans: ['Space Grotesk', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['IBM Plex Mono', 'ui-monospace', 'monospace'],
      },
      letterSpacing: {
        tighter: '-0.04em',
        tight: '-0.02em',
      },
      boxShadow: {
        'accent-glow': '0 0 26px rgba(255, 255, 255, 0.12)',
        'accent-glow-strong': '0 0 38px rgba(255, 255, 255, 0.20)',
      },
      animation: {
        'pulse-slow': 'pulse-ring 2s cubic-bezier(0, 0, 0.2, 1) infinite',
        flash: 'flash 0.3s ease-out',
      },
      keyframes: {
        'pulse-ring': {
          '0%': { transform: 'scale(1)', opacity: '0.5' },
          '100%': { transform: 'scale(2.2)', opacity: '0' },
        },
        flash: {
          '0%': { opacity: '0.6' },
          '100%': { opacity: '1' },
        },
      },
      backgroundImage: {
        noise:
          "url(\"data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.75' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)' opacity='0.028'/%3E%3C/svg%3E\")",
      },
    },
  },
  plugins: [],
};
