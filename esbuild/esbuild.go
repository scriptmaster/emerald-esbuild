package esbuild

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func BuildApp() error {
	var dir = "app"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("No app dir, creating...")
		WriteApp()
	}
	if _, err := os.Stat(path.Join(dir, "./main.tsx")); os.IsNotExist(err) {
		fmt.Println("No app main.tsx, creating...")
		WriteApp()
	}
	if _, err := os.Stat("./importmap.json"); os.IsNotExist(err) {
		fmt.Println("Cannot find importmap.json, creating...")
		WriteImportMap("./importmap.json")
	}
	if _, err := os.Stat("./deno.json"); os.IsNotExist(err) {
		fmt.Println("Cannot find deno.json, linking...")
		if err = os.Symlink("importmap.json", "deno.json"); err != nil {
			WriteImportMap("./deno.json")
		}
	}
	fmt.Println("Bundling with esbuild...")
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{dir + "/main.tsx"},
		Outdir:      "dist/",
		GlobalName:  "Emerald",

		AllowOverwrite:    true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		// Engines: []api.Engine{
		// 	{api.EngineChrome, "58"},
		// 	{api.EngineFirefox, "57"},
		// 	{api.EngineSafari, "11"},
		// 	{api.EngineEdge, "16"},
		// },
		// Alias:    GetImportsFromImportMap(),
		// Plugins:  []api.Plugin{envPlugin(), importmapPlugin(), esmUrlResolverPlugin()},
		Plugins:  []api.Plugin{envPlugin(), esmUrlResolverPlugin()},
		External: []string{"Alpine"},
		// JSXFactory:  "m",
		// JSXFragment: "m.Fragment",
		Write: true,
	})

	if len(result.Errors) > 0 {
		fmt.Printf("esbuild Errors: %v", result.Errors)
		return errors.New("esbuild api error: " + result.Errors[0].Text)
	}

	fmt.Println("esbuild Done")

	return nil
}

func envPlugin() api.Plugin {
	return api.Plugin{
		Name: "env",
		Setup: func(build api.PluginBuild) {
			// Intercept import paths called "env" so esbuild doesn't attempt
			// to map them to a file system location. Tag them with the "env-ns"
			// namespace to reserve them for this plugin.
			build.OnResolve(api.OnResolveOptions{Filter: `^env$`},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					// if (kind === "entry-point" && options?.fromEntryFile) {
					// 	return { path, namespace: "esbuild-svelte-direct-import" };
					// }
					return api.OnResolveResult{
						Path:      args.Path,
						Namespace: "env-ns",
					}, nil
				})

			// Load paths tagged with the "env-ns" namespace and behave as if
			// they point to a JSON file containing the environment variables.
			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "env-ns"},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					mappings := make(map[string]string)
					for _, item := range os.Environ() {
						if equals := strings.IndexByte(item, '='); equals != -1 {
							mappings[item[:equals]] = item[equals+1:]
						}
					}
					bytes, err := json.Marshal(mappings)
					if err != nil {
						return api.OnLoadResult{}, err
					}
					contents := string(bytes)
					return api.OnLoadResult{
						Contents: &contents,
						Loader:   api.LoaderJSON,
					}, nil
				})
		},
	}
}

func esmUrlResolverPlugin() api.Plugin {
	var importmap, err = os.ReadFile("importmap.json")
	if err != nil {
		fmt.Println("Error in reading importmap.json: ", err.Error())
		return api.Plugin{}
	}
	importMap := &ImportMap{}
	json.Unmarshal(importmap, &importMap)

	var esmPlugin = api.Plugin{
		Name: "esm",
		Setup: func(build api.PluginBuild) {
			// Intercept import paths called "env" so esbuild doesn't attempt
			// to map them to a file system location. Tag them with the "env-ns"
			// namespace to reserve them for this plugin.
			// filter := `^https\:\/\/esm.sh\/.+$`
			// filter := `.+`
			/*
				build.OnResolve(api.OnResolveOptions{Filter: `.*`, Namespace: "esm-ns"},
					func(args api.OnResolveArgs) (api.OnResolveResult, error) {
						// return nil, nil
						fmt.Println("esm Url Resolver Plugin / args.Path:", args.Path, ", ns:", args.Namespace, ",importer:", args.Importer, ", kind:", args.Kind)

						joinedPath := path.Join(path.Dir(args.Importer), args.Path)
						// fmt.Println(joinedPath)
						absPath, err := filepath.Abs(joinedPath)
						if err != nil {
							fmt.Println("Error in resolving absolute path from:", args.Path)
						}
						return api.OnResolveResult{
							Path:      absPath,
							Namespace: args.Namespace,
							// Importer:  args.Importer,
							// Kind:      args.Kind,
							// Namespace: "esm-ns",
						}, nil
					})
				build.OnResolve(api.OnResolveOptions{Filter: `https\:\/\/`, Namespace: "esm-ns"},
					func(args api.OnResolveArgs) (api.OnResolveResult, error) {
						// return nil, nil
						fmt.Println("esm Url Resolver Plugin https:// args.Path:", args.Path, ", ns:", args.Namespace, ",importer:", args.Importer, ", kind:", args.Kind)
						return api.OnResolveResult{
							// Path: args.Path,
							// Namespace: args.Namespace,
							// Namespace: "esm-ns",
						}, nil
					})
			*/

			build.OnResolve(api.OnResolveOptions{Filter: `.*`},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					// fmt.Println("esm Plugin: args.Path:", args.Path, ", ns:", args.Namespace, ",importer:", args.Importer, ", kind:", args.Kind)

					if url, ok := importMap.Imports[args.Path]; ok {
						return api.OnResolveResult{
							Path:      url,
							Namespace: "esm-ns",
						}, nil
					} else if args.Namespace == "esm-ns" {
						// fmt.Println("esm Plugin: args.Path:", args.Path, ", ns:", args.Namespace, ",importer:", args.Importer, ", kind:", args.Kind)
						if strings.HasPrefix(args.Path, ".") {
							joinedPath := path.Join(path.Dir(args.Importer), args.Path)
							absPath, err := filepath.Abs(joinedPath)
							if err != nil {
								fmt.Println("Error.1 in resolving absolute path from:", args.Path)
							}
							fmt.Println("Resolved ", args.Path, " to: ", absPath)
							return api.OnResolveResult{
								Path:      absPath,
								Namespace: args.Namespace,
							}, nil
						}
						return api.OnResolveResult{
							Path:      args.Path,
							Namespace: args.Namespace,
						}, nil
					} else if args.Namespace == "file" {
						// fmt.Println("args.Path", args.Path, ", importer:", args.Importer)
						joinedPath := path.Join(path.Dir(args.Importer), args.Path)
						// fmt.Println(joinedPath)
						absPath, err := filepath.Abs(joinedPath)
						if err != nil {
							fmt.Println("Error.2 in resolving absolute path from:", args.Path)
						}
						return api.OnResolveResult{
							Path:      absPath,
							Namespace: "file",
							// Namespace: "importmap-ns",
						}, nil
					} else {
						return api.OnResolveResult{
							Path:      args.Path,
							Namespace: args.Namespace,
							// Importer:  args.Importer,
							// Kind:      args.Kind,
						}, nil
					}
				})

			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "esm-ns"},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					// fmt.Println("esm-ns:", args.Path, args.Namespace, args.With, args.Suffix)
					url := args.Path
					if !strings.HasPrefix(url, "https://") {
						url = "https:/" + path.Join("/esm.sh", url)
					}
					// fmt.Println("GET", url, args.Namespace)
					res, err := http.Get(url)
					if err != nil {
						return api.OnLoadResult{}, err
					}
					defer res.Body.Close()
					bytes, err := io.ReadAll(res.Body)
					if err != nil {
						return api.OnLoadResult{}, err
					}
					contents := string(bytes)
					// return api.OnLoadResult{Contents: &contents, Loader: api.LoaderTSX}, nil
					return api.OnLoadResult{Contents: &contents}, nil
				})
		},
	}

	return esmPlugin
}

type ImportMap struct {
	Imports map[string]string `json:"imports"`
}

func GetImportsFromImportMap() map[string]string {
	var importmap, err = os.ReadFile("importmap.json")
	if err != nil {
		fmt.Println("Error in reading importmap.json: ", err.Error())
		return map[string]string{}
	}
	importMap := &ImportMap{}
	json.Unmarshal(importmap, &importMap)
	return importMap.Imports
}

/*
func importmapPlugin() api.Plugin {
	var importmap, err = os.ReadFile("importmap.json")
	if err != nil {
		fmt.Println("Error in reading importmap.json: ", err.Error())
		return api.Plugin{}
	}
	importMap := &ImportMap{}
	json.Unmarshal(importmap, &importMap)

	var importmapPlugin = api.Plugin{
		Name: "importmap",
		Setup: func(build api.PluginBuild) {
			// Intercept import paths called "env" so esbuild doesn't attempt
			// to map them to a file system location. Tag them with the "env-ns"
			// namespace to reserve them for this plugin.
			build.OnResolve(api.OnResolveOptions{Filter: `.*`},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					if url, ok := importMap.Imports[args.Path]; ok {
						return api.OnResolveResult{
							Path:      url,
							Namespace: "esm-ns",
						}, nil
					} else if args.Namespace == "file" {
						// fmt.Println("args.Path", args.Path, ", importer:", args.Importer)
						joinedPath := path.Join(path.Dir(args.Importer), args.Path)
						// fmt.Println(joinedPath)
						absPath, err := filepath.Abs(joinedPath)
						if err != nil {
							fmt.Println("Error in resolving absolute path from:", args.Path)
						}
						return api.OnResolveResult{
							Path:      absPath,
							Namespace: "file",
							// Namespace: "importmap-ns",
						}, nil
					} else {
						return api.OnResolveResult{
							Path:      args.Path,
							Namespace: args.Namespace,
							// Importer:  args.Importer,
							// Kind:      args.Kind,
						}, nil
					}
				})

		},
	}

	return importmapPlugin
}
*/
