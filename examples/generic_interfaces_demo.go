package main

import (
	"fmt"

	"github.com/arkan/poxxy"
)

// Cet exemple démontre l'amélioration apportée par les interfaces génériques
// Avant : type switching fragile dans validators.go (lignes 60-90)
// Après : interfaces génériques robustes avec ValidatorsAppender et DefaultValueSetter

func main() {
	fmt.Println("=== Démonstration des interfaces génériques améliorées ===")

	// 1. Différents types de champs avec validators - tous utilisent la même interface
	var name string
	var age int
	var scores []float64
	var settings map[string]string

	schema := poxxy.NewSchema(
		// ValueField
		poxxy.Value("name", &name,
			poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(2))),

		// ValueField avec type différent
		poxxy.Value("age", &age,
			poxxy.WithValidators(poxxy.Required(), poxxy.Min(18), poxxy.Max(120))),

		// SliceField
		poxxy.Slice("scores", &scores,
			poxxy.WithValidators(poxxy.Required(), poxxy.Each(poxxy.Min(0.0), poxxy.Max(100.0)))),

		// MapField
		poxxy.Map("settings", &settings,
			poxxy.WithValidators(poxxy.Required())),
	)

	data := map[string]interface{}{
		"name":   "Alice",
		"age":    25,
		"scores": []float64{85.5, 92.0, 78.5},
		"settings": map[string]interface{}{
			"theme": "dark",
			"lang":  "en",
		},
	}

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Tous les champs validés avec succès!\n")
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Age: %d\n", age)
	fmt.Printf("Scores: %v\n", scores)
	fmt.Printf("Settings: %v\n\n", settings)

	// 2. Démonstration des valeurs par défaut avec interface générique
	var title string
	var count int
	var active bool

	defaultSchema := poxxy.NewSchema(
		poxxy.Value("title", &title, poxxy.WithDefault("Untitled")),
		poxxy.Value("count", &count, poxxy.WithDefault(0)),
		poxxy.Value("active", &active, poxxy.WithDefault(true)),
	)

	// Données vides - les valeurs par défaut seront utilisées
	emptyData := map[string]interface{}{}

	if err := defaultSchema.Apply(emptyData); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Valeurs par défaut appliquées automatiquement!\n")
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Count: %d\n", count)
	fmt.Printf("Active: %t\n\n", active)

	// 3. Démonstration de l'extensibilité
	fmt.Println("Pour ajouter un nouveau type de champ, il suffit d'implémenter:")
	fmt.Println("- ValidatorsAppender pour les validators")
	fmt.Println("- DefaultValueSetter[T] pour les valeurs par défaut")
	fmt.Println("- L'interface Field pour l'intégration complète")

	fmt.Println("=== Avantages de la nouvelle approche ===")
	fmt.Println("✅ Plus de type switching fragile")
	fmt.Println("✅ Extensibilité facile - il suffit d'implémenter les interfaces")
	fmt.Println("✅ Maintenance réduite - pas besoin d'ajouter de nouveaux cas")
	fmt.Println("✅ Type safety amélioré avec les génériques")
	fmt.Println("✅ Code plus lisible et maintenable")
}
