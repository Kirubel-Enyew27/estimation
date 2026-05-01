package main

import (
	"context"
	"estimation/service"
	"estimation/store"
	"fmt"
	"log"
	"os"
)

const defaultMaterialCatalogPath = "data/materials.json"

func main() {
	catalogPath := defaultMaterialCatalogPath
	if len(os.Args) > 1 {
		catalogPath = os.Args[1]
	}

	catalog, err := store.LoadMaterialCatalog(catalogPath)
	if err != nil {
		log.Fatalf("failed to load material catalog: %v", err)
	}

	calculator := service.NewCalculatorWithMaterialStore(catalog)
	materials, err := catalog.List(context.Background())
	if err != nil {
		log.Fatalf("failed to inspect material catalog: %v", err)
	}

	fmt.Printf("Estimation service loaded %d materials from %s\n", len(materials), catalogPath)
	_ = calculator
}
