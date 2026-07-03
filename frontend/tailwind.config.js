/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        canvas: '#06070A',
        surface: '#0B0D13',
        'surface-elevated': '#11131A',
        border: {
          DEFAULT: '#1A1F2C',
          focus: '#263147',
          glow: '#2B374E',
        },
        accent: {
          DEFAULT: '#00F576',
          muted: 'rgba(0, 245, 118, 0.08)',
          glow: 'rgba(0, 245, 118, 0.15)',
          strong: 'rgba(0, 245, 118, 0.35)',
        },
        muted: {
          DEFAULT: '#64748B',
          soft: '#94A3B8',
        },
        active: '#F8FAFC',
        amber: {
          muted: '#B45309',
          soft: '#F59E0B',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
      },
      letterSpacing: {
        tighter: '-0.04em',
        tight: '-0.02em',
      },
      boxShadow: {
        'accent-glow': '0 0 20px rgba(0, 245, 118, 0.15)',
        'accent-glow-strong': '0 0 28px rgba(0, 245, 118, 0.28)',
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
          "url(\"data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)' opacity='0.04'/%3E%3C/svg%3E\")",
      },
    },
  },
  plugins: [],
};
