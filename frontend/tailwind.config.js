/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Muted stone palette: calm enough for long-running operations.
        primary: {
          50: '#f5f4f0',
          100: '#e8e7e1',
          200: '#d6d5cc',
          300: '#b8b8ad',
          400: '#92948a',
          500: '#73776e',
          600: '#5d625a',
          700: '#4a4f48',
          800: '#373b36',
          900: '#292c29',
          950: '#1d1f1d'
        },
        accent: {
          50: '#faf9f6',
          100: '#f1f0eb',
          200: '#e3e1d9',
          300: '#ceccc2',
          400: '#aaa9a0',
          500: '#85857d',
          600: '#686a63',
          700: '#50524d',
          800: '#393b37',
          900: '#272925',
          950: '#191a18'
        },
        dark: {
          50: '#f7f7f3',
          100: '#e8e8e2',
          200: '#d0d1c8',
          300: '#b0b2a8',
          400: '#8f9188',
          500: '#70736b',
          600: '#585b54',
          700: '#42453f',
          800: '#30332f',
          900: '#222421',
          950: '#171817'
        }
      },
      fontFamily: {
        sans: [
          'LXGW WenKai',
          '霞鹜文楷',
          'cursive'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: '0 2px 12px rgba(36, 37, 34, 0.06)',
        'glass-sm': '0 1px 6px rgba(36, 37, 34, 0.05)',
        glow: '0 0 0 rgba(0, 0, 0, 0)',
        'glow-lg': '0 0 0 rgba(0, 0, 0, 0)',
        card: '0 1px 2px rgba(36, 37, 34, 0.04)',
        'card-hover': '0 4px 14px rgba(36, 37, 34, 0.07)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.05)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': 'linear-gradient(135deg, #73776e 0%, #5d625a 100%)',
        'gradient-dark': 'linear-gradient(135deg, #30332f 0%, #171817 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.06) 0%, rgba(255,255,255,0.02) 100%)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(20, 184, 166, 0.25)' },
          '100%': { boxShadow: '0 0 30px rgba(20, 184, 166, 0.4)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
