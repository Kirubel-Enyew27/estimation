package main

import (
	"context"
	"estimation/handler"
	"estimation/service"
	"estimation/store"
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultMaterialCatalogPath = "data/materials.json"
const defaultServerAddress = ":8080"

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
	api := handler.New(calculator, catalog)
	materials, err := catalog.List(context.Background())
	if err != nil {
		log.Fatalf("failed to inspect material catalog: %v", err)
	}

	address := os.Getenv("ESTIMATOR_ADDR")
	if address == "" {
		address = defaultServerAddress
	}

	fmt.Printf("Estimation service loaded %d materials from %s\n", len(materials), catalogPath)

    fmt.Printf("Estimation service listening on %s\n", address)
	if err := http.ListenAndServe(address, api.Routes()); err != nil {
		log.Fatalf("estimation service stopped: %v", err)
	}
}
