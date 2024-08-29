package esbuild

import (
	"fmt"
	"os"
)

func WriteImportMap(filepath string) error {
	return os.WriteFile(filepath, []byte(`{
	"imports": {
		"alpinejs": "https://esm.sh/alpinejs",
		"react": "https://esm.sh/react",
		"react-dom": "https://esm.sh/react-dom",
		"moment": "https://esm.sh/moment"
	}
}`), os.FileMode(int(0777)))
}

func WriteApp() {
	if err := os.MkdirAll("./app", os.FileMode(int(0777))); err != nil {
		fmt.Println("error: ", err)
	}
	err := os.WriteFile("./app/main.tsx", []byte(`import Alpine from "alpinejs";

// const Alpine = window.Alpine;

/// --- DEFINE ALL OUR COMPONENTS AND PAGES ---
Alpine.data('content', () => ({}));

/// START ALPINE
Alpine.start();
`), os.FileMode(int(0777)))

	if err != nil {
		fmt.Println("error: ", err)
	}
}
