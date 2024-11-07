/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/view/**/*.html",
    "./internal/view/**/*.js",
    "./internal/view/**/*.templ",
    "./internal/view/**/*.go",
    "./cmd/server/*.go",
  ],
  theme: {
    colors: {
      // Base surfaces 
      base: '#1e1e2e',
      mantle: '#181825',
      crust: '#11111b',
      
      // Text
      text: '#cdd6f4',
      subtext: '#a6adc8',
      
      // Surface variants
      surface0: '#313244',
      surface1: '#45475a',
      
      // Accent colors
      blue: '#89b4fa',
      lavender: '#b4befe',
      red: '#f38ba8',
      peach: '#fab387',
      yellow: '#f9e2af',
      green: '#a6e3a1',
      mauve: '#cba6f7',
      
      // Keep some standard colors if needed
      transparent: 'transparent',
      current: 'currentColor',
      white: '#ffffff',
      black: '#000000',
    },
  },
  plugins: [],
}