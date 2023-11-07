/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./html/*.html", "./css/*.css", "./js/*.js", "./node_modules/tw-elements/dist/js/**/*.js"],
  theme: {
    extend: {},
  },
  plugins: [require("tw-elements/dist/plugin.cjs"), require("@tailwindcss/forms"), require("tailwindcss-animated")],
};

