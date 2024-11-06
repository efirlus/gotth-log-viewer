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
    extend: {}
  },
  plugins: [],
}