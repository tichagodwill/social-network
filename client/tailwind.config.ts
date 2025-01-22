import type {Config} from 'tailwindcss';
import flowbitePlugin from 'flowbite/plugin';
import typography from '@tailwindcss/typography';
import forms from '@tailwindcss/forms';
import aspectRatio from '@tailwindcss/aspect-ratio';
import type {PluginAPI} from 'tailwindcss/types/config';

export default {
    content: ['./src/**/*.{html,js,svelte,ts}', './node_modules/flowbite-svelte/**/*.{html,js,svelte,ts}'],
    darkMode: 'selector',
    theme: {
        container: {
            center: true,
            padding: {
                DEFAULT: '1rem',
                sm: '2rem',
                lg: '4rem',
                xl: '5rem',
                '2xl': '6rem',
            },
        },
        fontFamily: {
            sans: ['Inter', '"PT Sans"', 'system-ui', 'sans-serif'],
            display: ['Inter', '"PT Sans"', 'system-ui', 'sans-serif'],
            mono: ['JetBrains Mono', 'monospace'],
        },
        extend: {
            colors: {
                // Primary colors with adjusted shades
                primary: {
                    50: '#eef2ff',
                    100: '#e0e7ff',
                    200: '#c7d2fe',
                    300: '#a5b4fc',
                    400: '#818cf8',
                    500: '#6366f1',  // Main Primary Color
                    600: '#4f46e5',
                    700: '#4338ca',
                    800: '#3730a3',
                    900: '#312e81',
                },
                // Social colors
                social: {
                    facebook: '#1877F2',
                    twitter: '#1DA1F2',
                    instagram: '#E4405F',
                    linkedin: '#0A66C2',
                },
                // Semantic colors
                success: {
                    light: '#D1FAE5',
                    DEFAULT: '#10B981',
                    dark: '#065F46',
                },
                warning: {
                    light: '#FEF3C7',
                    DEFAULT: '#F59E0B',
                    dark: '#92400E',
                },
                error: {
                    light: '#FEE2E2',
                    DEFAULT: '#EF4444',
                    dark: '#991B1B',
                },
                info: {
                    light: '#DBEAFE',
                    DEFAULT: '#3B82F6',
                    dark: '#1E40AF',
                },
            },
            spacing: {
                '128': '32rem',
                '144': '36rem',
            },
            borderRadius: {
                '4xl': '2rem',
            },
            fontSize: {
                '2xs': ['0.625rem', {lineHeight: '0.75rem'}],
                '3xl': ['2rem', {lineHeight: '2.25rem'}],
                '4xl': ['2.5rem', {lineHeight: '2.75rem'}],
                '5xl': ['3rem', {lineHeight: '3.25rem'}],
            },
            animation: {
                'bounce-slow': 'bounce 3s linear infinite',
                'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
                'spin-slow': 'spin 3s linear infinite',
                'ping-slow': 'ping 3s cubic-bezier(0, 0, 0.2, 1) infinite',
            },
            transitionDuration: {
                '400': '400ms',
            },
            boxShadow: {
                'inner-lg': 'inset 0 2px 4px 0 rgb(0 0 0 / 0.10)',
            },
            keyframes: {
                wiggle: {
                    '0%, 100%': {transform: 'rotate(-3deg)'},
                    '50%': {transform: 'rotate(3deg)'},
                },
                'fade-in-up': {
                    '0%': {
                        opacity: '0',
                        transform: 'translateY(10px)',
                    },
                    '100%': {
                        opacity: '1',
                        transform: 'translateY(0)',
                    },
                },
                'fade-out-down': {
                    '0%': {
                        opacity: '1',
                        transform: 'translateY(0)',
                    },
                    '100%': {
                        opacity: '0',
                        transform: 'translateY(10px)',
                    },
                },
            },
            backgroundImage: {
                'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
                'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
            },
        },
    },
    plugins: [
        flowbitePlugin,
        typography,
        forms,
        aspectRatio,
        ({addBase, theme}: PluginAPI) => {
            addBase({
                '*:focus-visible': {
                    outline: `2px solid ${theme('colors.primary.500')}`,
                    outlineOffset: '2px',
                },
                '*::-webkit-scrollbar': {
                    width: '6px',
                    height: '6px',
                },
                '*::-webkit-scrollbar-track': {
                    background: theme('colors.gray.100'),
                },
                '*::-webkit-scrollbar-thumb': {
                    background: theme('colors.gray.300'),
                    borderRadius: '3px',
                },
                '*::-webkit-scrollbar-thumb:hover': {
                    background: theme('colors.gray.400'),
                },
                'html': {
                    '-webkit-tap-highlight-color': 'transparent',
                },
                'body': {
                    '@apply antialiased text-gray-800 bg-gray-50': {},
                }, ':root': {
                    '--primary-50': theme('colors.primary.50'),
                    '--primary-100': theme('colors.primary.100'),
                    '--primary-200': theme('colors.primary.200'),
                    '--primary-300': theme('colors.primary.300'),
                    '--primary-400': theme('colors.primary.400'),
                    '--primary-500': theme('colors.primary.500'),
                    '--primary-600': theme('colors.primary.600'),
                    '--primary-700': theme('colors.primary.700'),
                    '--primary-800': theme('colors.primary.800'),
                    '--primary-900': theme('colors.primary.900'),
                }
            });
        },
    ],
} satisfies Config;