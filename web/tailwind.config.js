/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: '#0f172a',
        surface: '#1e293b',
        border: '#334155',
        'status-working': '#3b82f6',
        'status-waiting': '#eab308',
        'status-error': '#ef4444',
        'status-dead': '#64748b',
        'status-idle': '#10b981',
      },
    },
  },
  plugins: [],
}

