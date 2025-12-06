import { heroui } from "@heroui/theme";

/** @type {import('tailwindcss').Config} */
export default {
	content: [
		"./index.html",
		"./src/layouts/**/*.{js,ts,jsx,tsx,mdx}",
		"./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
		"./src/components/**/*.{js,ts,jsx,tsx,mdx}",
		"./node_modules/@heroui/theme/dist/**/*.{js,ts,jsx,tsx}",
	],
	theme: {
		extend: {
			fontFamily: {
				sans: ["Inter", "sans-serif"],
				mono: ["JetBrains Mono", "monospace"],
			},
		},
	},
	darkMode: "class",
	plugins: [
		heroui({
			themes: {
				dark: {
					colors: {
						// BAZA (Ciemny Grafit - bez zmian)
						background: "#09090b",
						foreground: "#ECEDEE", // Twardy biały/szary dla tekstu (żeby nie był żółty!)

						// WARSTWY (Zinc - bez zmian)
						content1: "#18181b",
						content2: "#27272a",
						content3: "#3f3f46",
						content4: "#52525b",

						// AKCENT: INDUSTRIAL ORANGE 🟠
						primary: {
							// 50-900: Pełna paleta dla stanów hover/active
							50: "#fff7ed",
							100: "#ffedd5",
							200: "#fed7aa",
							300: "#fdba74",
							400: "#fb923c",
							500: "#f97316", // Nasz główny kolor
							600: "#ea580c",
							700: "#c2410c",
							800: "#9a3412",
							900: "#7c2d12",
							DEFAULT: "#f97316", // To jest domyślny kolor przycisków
							foreground: "#ffffff", // Tekst na pomarańczowym przycisku (biały)
						},

						// FOCUS RING (Musi pasować do akcentu)
						focus: "#f97316",

						// Reszta semantyki
						danger: { DEFAULT: "#f31260", foreground: "#ffffff" },
						success: { DEFAULT: "#17c964", foreground: "#000000" },
						warning: { DEFAULT: "#f5a524", foreground: "#000000" }, // Warning może zostać żółty

						default: { DEFAULT: "#3f3f46", foreground: "#ecedee" },
					},
				},
			},
		}),
	],
};
