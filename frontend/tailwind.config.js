/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        canvas: '#071013',
        surface: '#0D171B',
        'surface-elevated': '#132229',
        border: {
          DEFAULT: 'rgba(178, 218, 224, 0.14)',
          focus: 'rgba(104, 225, 253, 0.35)',
          glow: 'rgba(104, 225, 253, 0.36)',
        },
        accent: {
          DEFAULT: '#68E1FD',
          muted: 'rgba(104, 225, 253, 0.08)',
          glow: 'rgba(104, 225, 253, 0.16)',
          strong: 'rgba(104, 225, 253, 0.35)',
        },
        muted: {
          DEFAULT: '#78929A',
          soft: '#B6C8CC',
        },
        active: '#F6FBF8',
        amber: {
          muted: '#B7791F',
          soft: '#FFB454',
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
        'accent-glow': '0 0 26px rgba(104, 225, 253, 0.18)',
        'accent-glow-strong': '0 0 38px rgba(104, 225, 253, 0.32)',
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
