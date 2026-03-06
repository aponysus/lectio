/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        canvas: '#f4efe4',
        ink: '#1f2933',
        accent: '#8b5e34',
        pine: '#27473b',
      },
      fontFamily: {
        display: ['Georgia', 'serif'],
        body: ['"Trebuchet MS"', '"Segoe UI"', 'sans-serif'],
      },
      boxShadow: {
        card: '0 20px 50px -24px rgba(31, 41, 51, 0.35)',
      },
    },
  },
  plugins: [],
}
