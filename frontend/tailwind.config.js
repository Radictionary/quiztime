/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./html/*.html", "./css/*.css", "./js/*.js", "./node_modules/flowbite/**/*.js"],
  theme: {
    extend: {},
  },
  plugins: [require("flowbite/plugin"), require("@tailwindcss/forms"), require("tailwindcss-animated")],
  daisyui: {
    themes: false,
  },
};

