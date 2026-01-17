# Product Requirements Document (PRD)
# Poxxy V2 - Améliorations DSL et Expérience Développeur

**Version:** 1.0
**Date:** 2026-01-17
**Auteur:** Engineering Team
**Status:** Draft

---

## Table des matières

1. [Executive Summary](#1-executive-summary)
2. [Contexte et Problématique](#2-contexte-et-problématique)
3. [Objectifs](#3-objectifs)
4. [Analyse Comparative Main vs V2](#4-analyse-comparative-main-vs-v2)
5. [Spécifications Fonctionnelles](#5-spécifications-fonctionnelles)
   - [5.1 Cross-field Validation](#51-cross-field-validation)
   - [5.2 Conditional Fields](#52-conditional-fields)
   - [5.3 Validation Groups](#53-validation-groups)
   - [5.4 Struct Tags Auto-Schema](#54-struct-tags-auto-schema)
   - [5.5 Fail-Fast Mode](#55-fail-fast-mode)
   - [5.6 Simplified Slice API](#56-simplified-slice-api)
   - [5.7 Discriminated Unions](#57-discriminated-unions)
   - [5.8 Sanitizers vs Transformers](#58-sanitizers-vs-transformers)
   - [5.9 Schema Composition](#59-schema-composition)
   - [5.10 HTTP Ergonomics](#510-http-ergonomics)
   - [5.11 Async Validation](#511-async-validation)
   - [5.12 Error Formatting](#512-error-formatting)
6. [Priorités et Roadmap](#6-priorités-et-roadmap)
7. [Métriques de Succès](#7-métriques-de-succès)
8. [Risques et Mitigations](#8-risques-et-mitigations)
9. [Annexes](#9-annexes)

---

## 1. Executive Summary

### Vision

Poxxy V2 vise à devenir **la bibliothèque de validation Go de référence** en offrant:
- Un DSL expressif et type-safe
- Une expérience développeur (DX) exceptionnelle
- Des fonctionnalités avancées absentes des alternatives (cross-field validation, conditional fields, i18n)

### Scope

Ce PRD couvre **12 améliorations majeures** pour Poxxy V2, organisées en 3 phases de développement. Ces améliorations introduisent des **breaking changes** par rapport à la V2 actuelle, justifiés par les gains significatifs en DX.

### Résumé des changements clés

| Catégorie | Fonctionnalités |
|-----------|-----------------|
| **Validation avancée** | Cross-field, Conditional fields, Validation groups, Fail-fast |
| **Réduction boilerplate** | Struct tags, Schema composition, Simplified slice API |
| **Types complexes** | Discriminated unions améliorés |
| **Sécurité** | Séparation Sanitizers/Transformers |
| **Intégration** | HTTP helpers, Async validation, Error formatting |

---

## 2. Contexte et Problématique

### 2.1 Situation actuelle

Poxxy existe en deux versions:

| Version | Status | Caractéristiques |
|---------|--------|------------------|
| **Main** | Production (buggy) | API verbeux, callbacks nested, pas d'i18n |
| **V2** | Development | Rewrite avec meilleur DSL, i18n, introspection |

### 2.2 Problèmes identifiés dans Main

1. **Verbosité excessive**
   ```go
   // 8 lignes pour un simple champ email
   poxxy.Value("email", &email,
       poxxy.WithTransformers(poxxy.TrimSpace(), poxxy.ToLower()),
       poxxy.WithValidators(poxxy.Required(), poxxy.Email()),
       poxxy.WithDescription("User email"),
   )
   ```

2. **Callbacks imbriqués illisibles**
   ```go
   poxxy.Struct("user", &user,
       poxxy.WithSubSchema(func(s *poxxy.Schema, u *User) {
           poxxy.WithSchema(s, poxxy.Value("name", &u.Name,
               poxxy.WithValidators(poxxy.Required()),
           ))
       }),
   )
   ```

3. **Pas de validation inter-champs** - Impossible de valider `confirmPassword == password`

4. **Pas de champs conditionnels** - Impossible de rendre un champ requis selon un autre

5. **Erreurs peu exploitables** - Format basique, pas de groupement, pas de traduction

### 2.3 Améliorations V2 actuelles

La V2 apporte déjà:
- ✅ Fluent API (method chaining)
- ✅ Support i18n avec package dédié
- ✅ Introspection et génération JSON Schema/OpenAPI
- ✅ Templates réutilisables
- ✅ Meilleure structure d'erreurs (`ValidationErrors`)
- ✅ Séparation HTTP dans `adapters/`

### 2.4 Gaps restants

Malgré les améliorations, des fonctionnalités critiques manquent:
- ❌ Cross-field validation
- ❌ Conditional fields
- ❌ Validation groups
- ❌ Struct tags pour auto-génération
- ❌ Fail-fast mode
- ❌ Async validation
- ❌ Schema composition

---

## 3. Objectifs

### 3.1 Objectifs business

| Objectif | Métrique cible |
|----------|----------------|
| Adoption | +50% d'utilisateurs vs Main en 6 mois |
| Satisfaction | NPS > 40 (feedback développeurs) |
| Migration | 80% des projets Main migrés vers V2 en 12 mois |

### 3.2 Objectifs techniques

| Objectif | Description |
|----------|-------------|
| **Expressivité** | Réduire le boilerplate de 40% vs Main |
| **Type-safety** | 100% des APIs avec génériques, 0 `interface{}` exposé |
| **Performance** | Validation < 1ms pour 50 champs |
| **Testabilité** | Couverture > 90%, tous les cas edge documentés |

### 3.3 Principes directeurs

1. **Explicit over implicit** - Pas de magie, comportement prévisible
2. **Progressive disclosure** - Simple par défaut, puissant si besoin
3. **Fail fast, fail clear** - Erreurs détaillées et actionnables
4. **Composition over inheritance** - Schemas réutilisables et composables

---

## 4. Analyse Comparative Main vs V2

### 4.1 Comparaison DSL

```go
// ═══════════════════════════════════════════════════════════════
// MAIN - Version actuelle (verbose, nested callbacks)
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(
    poxxy.Value("email", &email,
        poxxy.WithTransformers(poxxy.TrimSpace(), poxxy.ToLower()),
        poxxy.WithValidators(poxxy.Required(), poxxy.Email()),
        poxxy.WithDescription("User email"),
    ),
    poxxy.Struct("address", &address,
        poxxy.WithSubSchema(func(s *poxxy.Schema, a *Address) {
            poxxy.WithSchema(s, poxxy.Value("city", &a.City,
                poxxy.WithValidators(poxxy.Required()),
            ))
            poxxy.WithSchema(s, poxxy.Value("zip", &a.Zip,
                poxxy.WithValidators(poxxy.Required(), poxxy.Pattern(`^\d{5}$`)),
            ))
        }),
    ),
)
err := schema.Apply(data)

// ═══════════════════════════════════════════════════════════════
// V2 ACTUELLE - Meilleur mais perfectible
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(
    poxxy.Field("email", &email).
        Transform(poxxy.TrimSpace(), poxxy.ToLower()).
        Validate(poxxy.Required[string](), poxxy.Email()).
        Describe("User email"),
    poxxy.Struct("address", &address, func(s *poxxy.Schema) {
        s.Add(poxxy.Field("city", &a.City).Validate(poxxy.Required[string]()))
        s.Add(poxxy.Field("zip", &a.Zip).Validate(poxxy.Required[string](), poxxy.Pattern(`^\d{5}$`)))
    }),
)
err := schema.ParseMap(data)

// ═══════════════════════════════════════════════════════════════
// V2 AMÉLIORÉE - Vision cible (ce PRD)
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(
    poxxy.Field("email", &email).
        Sanitize(poxxy.TrimSpace(), poxxy.ToLower()).
        Validate(poxxy.Required(), poxxy.Email()).
        Describe("User email"),
    poxxy.StructAuto("address", &address), // Auto depuis struct tags
)
err := schema.Parse(data)
```

### 4.2 Tableau comparatif des fonctionnalités

| Fonctionnalité | Main | V2 Actuelle | V2 Améliorée |
|----------------|------|-------------|--------------|
| Fluent API | ❌ | ✅ | ✅ |
| i18n | ❌ | ✅ | ✅ |
| JSON Schema | ❌ | ✅ | ✅ |
| Cross-field validation | ❌ | ❌ | ✅ |
| Conditional fields | ❌ | ❌ | ✅ |
| Validation groups | ❌ | ❌ | ✅ |
| Struct tags | ❌ | ❌ | ✅ |
| Fail-fast mode | ❌ | ❌ | ✅ |
| Slice `.Each()` | ❌ | Partiel | ✅ |
| Discriminated unions | Basique | Basique | ✅ |
| Sanitizers séparés | ❌ | ❌ | ✅ |
| Schema composition | ❌ | ❌ | ✅ |
| HTTP helpers | Basique | Adapter | ✅ Enhanced |
| Async validation | ❌ | ❌ | ✅ |
| Error formatting | Basique | Amélioré | ✅ Extensible |

---

## 5. Spécifications Fonctionnelles

### 5.1 Cross-field Validation

#### 5.1.1 Description

Permettre la validation d'un champ en fonction de la valeur d'un ou plusieurs autres champs du même schema.

#### 5.1.2 Cas d'usage

| Use Case | Exemple |
|----------|---------|
| Confirmation mot de passe | `confirmPassword` doit égaler `password` |
| Plage de dates | `endDate` doit être après `startDate` |
| Dépendance conditionnelle | `discountCode` valide si `total > 100` |
| Somme contrôlée | `items.sum(price)` doit égaler `total` |

#### 5.1.3 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// OPTION A: Référence directe (recommandé pour cas simples)
// ═══════════════════════════════════════════════════════════════

var password, confirmPassword string
var startDate, endDate time.Time

schema := poxxy.NewSchema(
    poxxy.Field("password", &password).
        Validate(poxxy.Required(), poxxy.MinLength(8)),

    // Référence au champ password via pointeur
    poxxy.Field("confirm_password", &confirmPassword).
        Validate(poxxy.Required(), poxxy.EqualTo(&password)),

    poxxy.Field("start_date", &startDate).
        Validate(poxxy.Required()),

    // Comparaison temporelle
    poxxy.Field("end_date", &endDate).
        Validate(poxxy.Required(), poxxy.After(&startDate)),
)

// ═══════════════════════════════════════════════════════════════
// OPTION B: Callback avec accès au schema (cas complexes)
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(
    poxxy.Field("password", &password).Validate(poxxy.Required()),
    poxxy.Field("confirm_password", &confirmPassword).
        ValidateWith(func(value string, s *poxxy.Schema) error {
            if value != password {
                return fmt.Errorf("passwords do not match")
            }
            return nil
        }),
)

// ═══════════════════════════════════════════════════════════════
// OPTION C: Validation au niveau schema (validation globale)
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(
    poxxy.Field("password", &password),
    poxxy.Field("confirm_password", &confirmPassword),
).CrossValidate(func(s *poxxy.Schema) error {
    if password != confirmPassword {
        return poxxy.FieldError("confirm_password", "must match password")
    }
    return nil
})
```

#### 5.1.4 Validators cross-field intégrés

| Validator | Signature | Description |
|-----------|-----------|-------------|
| `EqualTo` | `EqualTo[T](*T)` | Valeur doit égaler la référence |
| `NotEqualTo` | `NotEqualTo[T](*T)` | Valeur doit différer |
| `GreaterThan` | `GreaterThan[T Ordered](*T)` | Valeur > référence |
| `LessThan` | `LessThan[T Ordered](*T)` | Valeur < référence |
| `After` | `After(*time.Time)` | Date après référence |
| `Before` | `Before(*time.Time)` | Date avant référence |
| `Between` | `Between[T Ordered](*T, *T)` | Entre deux références |

#### 5.1.5 Comportement

1. Les validateurs cross-field s'exécutent en **Phase 2** (après l'assignation de tous les champs)
2. L'ordre de déclaration n'importe pas - tous les champs sont assignés avant validation
3. Si le champ référencé est absent/null, le comportement est configurable:
   ```go
   poxxy.EqualTo(&password).SkipIfAbsent() // Skip validation si password absent
   poxxy.EqualTo(&password).FailIfAbsent() // Fail si password absent (défaut)
   ```

#### 5.1.6 Contraintes techniques

- Les références doivent être des pointeurs vers des variables du même scope
- Type-safety via génériques: `EqualTo[T]` ne compile pas si types incompatibles
- Thread-safe: pas de race condition car validation séquentielle

---

### 5.2 Conditional Fields

#### 5.2.1 Description

Permettre de rendre un champ requis, optionnel, ou d'appliquer des validators différents selon la valeur d'autres champs.

#### 5.2.2 Cas d'usage

| Use Case | Condition | Comportement |
|----------|-----------|--------------|
| Paiement carte | `paymentMethod == "card"` | `cardNumber` requis |
| Livraison | `deliveryType == "shipping"` | `address` requis |
| Professionnel | `accountType == "business"` | `companyName`, `vatNumber` requis |
| Remise | `hasDiscount == true` | `discountCode` requis |

#### 5.2.3 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Syntaxe fluide avec .When()
// ═══════════════════════════════════════════════════════════════

var paymentMethod string
var cardNumber, cardCVV string
var iban string

schema := poxxy.NewSchema(
    poxxy.Field("payment_method", &paymentMethod).
        Validate(poxxy.Required(), poxxy.In("card", "bank_transfer", "paypal")),

    // Requis seulement si paiement par carte
    poxxy.Field("card_number", &cardNumber).
        When(poxxy.Equals(&paymentMethod, "card")).
        Validate(poxxy.Required(), poxxy.CreditCard()),

    poxxy.Field("card_cvv", &cardCVV).
        When(poxxy.Equals(&paymentMethod, "card")).
        Validate(poxxy.Required(), poxxy.Length(3, 4)),

    // Requis si virement bancaire
    poxxy.Field("iban", &iban).
        When(poxxy.Equals(&paymentMethod, "bank_transfer")).
        Validate(poxxy.Required(), poxxy.IBAN()),
)

// ═══════════════════════════════════════════════════════════════
// Condition complexe avec callback
// ═══════════════════════════════════════════════════════════════

var total float64
var discountCode string

schema := poxxy.NewSchema(
    poxxy.Field("total", &total).Validate(poxxy.Required(), poxxy.Min(0.0)),

    poxxy.Field("discount_code", &discountCode).
        When(func() bool { return total > 100 }).
        Validate(poxxy.Required(), poxxy.Pattern(`^[A-Z]{4}\d{4}$`)),
)

// ═══════════════════════════════════════════════════════════════
// Helpers raccourcis
// ═══════════════════════════════════════════════════════════════

poxxy.Field("card_number", &cardNumber).
    RequiredWhen(poxxy.Equals(&paymentMethod, "card")). // Raccourci
    Validate(poxxy.CreditCard())

poxxy.Field("notes", &notes).
    OptionalWhen(poxxy.Equals(&priority, "low")). // Ignore validation si low
    Validate(poxxy.MaxLength(1000))
```

#### 5.2.4 Conditions intégrées

| Condition | Signature | Description |
|-----------|-----------|-------------|
| `Equals` | `Equals[T](*T, T)` | Vrai si champ == valeur |
| `NotEquals` | `NotEquals[T](*T, T)` | Vrai si champ != valeur |
| `In` | `In[T](*T, ...T)` | Vrai si champ dans liste |
| `NotIn` | `NotIn[T](*T, ...T)` | Vrai si champ pas dans liste |
| `IsPresent` | `IsPresent(*T)` | Vrai si champ fourni |
| `IsAbsent` | `IsAbsent(*T)` | Vrai si champ absent |
| `IsEmpty` | `IsEmpty(*T)` | Vrai si champ vide |
| `IsNotEmpty` | `IsNotEmpty(*T)` | Vrai si champ non vide |
| `GreaterThan` | `GreaterThan[T](*T, T)` | Vrai si champ > valeur |
| `Custom` | `Custom(func() bool)` | Condition arbitraire |

#### 5.2.5 Combinaison de conditions

```go
// ET logique
poxxy.Field("cvv", &cvv).
    When(poxxy.And(
        poxxy.Equals(&paymentMethod, "card"),
        poxxy.IsNotEmpty(&cardNumber),
    ))

// OU logique
poxxy.Field("contact", &contact).
    When(poxxy.Or(
        poxxy.IsAbsent(&email),
        poxxy.IsAbsent(&phone),
    ))

// NON logique
poxxy.Field("reason", &reason).
    When(poxxy.Not(poxxy.Equals(&status, "approved")))
```

#### 5.2.6 Comportement

1. **Évaluation lazy**: La condition est évaluée **après** l'assignation de tous les champs
2. **Skip complet**: Si condition fausse, le champ est complètement ignoré (pas d'assignation, pas de validation)
3. **Chaînage**: Plusieurs `.When()` sont combinés en AND
4. **Erreurs claires**: Message indique pourquoi le champ est requis

```go
// Erreur exemple
{
    "field": "card_number",
    "error": "required when payment_method is 'card'",
    "condition": "payment_method == card"
}
```

---

### 5.3 Validation Groups

#### 5.3.1 Description

Permettre d'appliquer des règles de validation différentes selon le contexte d'utilisation (création, mise à jour, import, etc.).

#### 5.3.2 Cas d'usage

| Contexte | Exemple |
|----------|---------|
| **Create** | `id` auto-généré (pas requis), `email` requis et unique |
| **Update** | `id` requis, `email` optionnel (garde existant si absent) |
| **Import** | Validation plus souple, `id` peut être fourni |
| **Admin** | Champs supplémentaires autorisés |

#### 5.3.3 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Définition des groupes (constantes typées)
// ═══════════════════════════════════════════════════════════════

const (
    GroupCreate poxxy.Group = "create"
    GroupUpdate poxxy.Group = "update"
    GroupImport poxxy.Group = "import"
)

// ═══════════════════════════════════════════════════════════════
// Schema avec validation par groupe
// ═══════════════════════════════════════════════════════════════

var id int
var email, name string

schema := poxxy.NewSchema(
    // ID: requis seulement en update
    poxxy.Field("id", &id).
        Groups(GroupUpdate).              // Ce champ n'existe qu'en update
        Validate(poxxy.Required()),

    // Email: requis en create, optionnel en update
    poxxy.Field("email", &email).
        ValidateIn(GroupCreate, poxxy.Required(), poxxy.Email()).
        ValidateIn(GroupUpdate, poxxy.Email()). // Optionnel
        ValidateIn(GroupImport, poxxy.Email()), // Optionnel

    // Name: toujours requis (pas de groupe = tous les groupes)
    poxxy.Field("name", &name).
        Validate(poxxy.Required(), poxxy.MinLength(2)),
)

// ═══════════════════════════════════════════════════════════════
// Utilisation
// ═══════════════════════════════════════════════════════════════

// Création: vérifie Required sur email
err := schema.ParseWithGroups(data, GroupCreate)

// Mise à jour: id requis, email optionnel
err := schema.ParseWithGroups(data, GroupUpdate)

// Sans groupe: applique tous les validators (union)
err := schema.Parse(data)

// Plusieurs groupes (validators des deux)
err := schema.ParseWithGroups(data, GroupCreate, GroupAdmin)
```

#### 5.3.4 Comportement avancé

```go
// ═══════════════════════════════════════════════════════════════
// Champ exclusif à un groupe
// ═══════════════════════════════════════════════════════════════

poxxy.Field("created_by", &createdBy).
    OnlyIn(GroupAdmin). // Ignoré complètement si pas admin
    Validate(poxxy.Required())

// ═══════════════════════════════════════════════════════════════
// Exclusion de groupe
// ═══════════════════════════════════════════════════════════════

poxxy.Field("password", &password).
    ExceptIn(GroupUpdate). // Jamais en update
    Validate(poxxy.Required(), poxxy.MinLength(8))

// ═══════════════════════════════════════════════════════════════
// Groupe par défaut
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(...).DefaultGroup(GroupCreate)
err := schema.Parse(data) // Utilise GroupCreate par défaut
```

#### 5.3.5 Introspection

```go
// Liste des champs par groupe (pour doc/OpenAPI)
createFields := schema.FieldsForGroup(GroupCreate)
updateFields := schema.FieldsForGroup(GroupUpdate)

// Génération OpenAPI avec groupes
openapi.GenerateWithGroups(schema, map[poxxy.Group]string{
    GroupCreate: "UserCreate",
    GroupUpdate: "UserUpdate",
})
```

---

### 5.4 Struct Tags Auto-Schema

#### 5.4.1 Description

Permettre de générer automatiquement un schema à partir des struct tags Go, réduisant drastiquement le boilerplate pour les cas courants.

#### 5.4.2 Cas d'usage

- Formulaires simples avec validation standard
- DTOs avec règles de validation courantes
- Prototypage rapide
- Migration depuis d'autres bibliothèques (go-playground/validator)

#### 5.4.3 Syntaxe des tags

```go
type User struct {
    // Format: `poxxy:"validator1,validator2=arg,transform=t1|t2,default=value"`

    ID        int       `poxxy:"-"`                                    // Ignoré
    Name      string    `poxxy:"required,min=2,max=100,transform=trim|capitalize"`
    Email     string    `poxxy:"required,email,transform=sanitize_email"`
    Age       int       `poxxy:"min=18,max=120,default=25"`
    Role      string    `poxxy:"required,in=admin|user|guest,default=user"`
    Website   string    `poxxy:"url,optional"`                         // Optionnel explicite
    Phone     string    `poxxy:"pattern=^\\+?[0-9]{10\\,15}$"`         // Regex
    CreatedAt time.Time `poxxy:"default=now"`                          // Valeur spéciale

    // Nested struct: validation récursive automatique
    Address   Address   `poxxy:"required"`

    // Slice avec validation des éléments
    Tags      []string  `poxxy:"min_items=1,max_items=10,each=min=1|max=50"`
}

type Address struct {
    Street  string `poxxy:"required,min=5"`
    City    string `poxxy:"required"`
    ZipCode string `poxxy:"required,pattern=^\\d{5}$"`
    Country string `poxxy:"required,in=FR|BE|CH|CA,default=FR"`
}
```

#### 5.4.4 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Auto-génération complète
// ═══════════════════════════════════════════════════════════════

var user User
schema := poxxy.NewSchema(
    poxxy.StructAuto("user", &user),
)

// ═══════════════════════════════════════════════════════════════
// Hybride: tags + overrides manuels
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(
    poxxy.StructAuto("user", &user).
        Override("email", func(f *poxxy.FieldBuilder[string]) {
            f.Validate(customEmailValidator) // Ajoute un validator custom
        }).
        Ignore("internal_field"), // Ignore un champ même si taggé
)

// ═══════════════════════════════════════════════════════════════
// Schema from type (sans variable)
// ═══════════════════════════════════════════════════════════════

schema := poxxy.SchemaFromType[User]()
var user User
err := schema.ParseInto(data, &user)

// ═══════════════════════════════════════════════════════════════
// Configuration globale des tags
// ═══════════════════════════════════════════════════════════════

poxxy.SetTagName("validate") // Utilise `validate:"..."` au lieu de `poxxy:"..."`
```

#### 5.4.5 Mapping des tags

| Tag | Validator/Comportement | Exemple |
|-----|------------------------|---------|
| `required` | `Required()` | `required` |
| `optional` | Pas de `Required` | `optional` |
| `min=N` | `Min(N)` ou `MinLength(N)` | `min=5` |
| `max=N` | `Max(N)` ou `MaxLength(N)` | `max=100` |
| `email` | `Email()` | `email` |
| `url` | `URL()` | `url` |
| `uuid` | `UUID()` | `uuid` |
| `in=a\|b\|c` | `In("a","b","c")` | `in=admin\|user` |
| `pattern=X` | `Pattern(X)` | `pattern=^\d+$` |
| `default=V` | `Default(V)` | `default=25` |
| `transform=X` | Transformers chaînés | `transform=trim\|lower` |
| `each=X` | Validation éléments slice | `each=min=1` |
| `min_items=N` | `MinItems(N)` | `min_items=1` |
| `max_items=N` | `MaxItems(N)` | `max_items=10` |
| `-` | Ignorer le champ | `-` |

#### 5.4.6 Transformers disponibles via tags

| Tag | Transformer |
|-----|-------------|
| `trim` | `TrimSpace()` |
| `lower` | `ToLower()` |
| `upper` | `ToUpper()` |
| `title` | `TitleCase()` |
| `capitalize` | `Capitalize()` |
| `sanitize_email` | `SanitizeEmail()` |

#### 5.4.7 Comportement

1. **Opt-in explicite**: Seuls les champs avec tag `poxxy:` sont validés
2. **Nested automatique**: Les structs imbriqués sont validés récursivement
3. **Cache**: Le parsing des tags est caché pour performance
4. **Erreurs claires**: Erreur si tag invalide avec position exacte
5. **Priorité override**: Les overrides manuels ont priorité sur les tags

---

### 5.5 Fail-Fast Mode

#### 5.5.1 Description

Permettre d'arrêter la validation dès la première erreur, utile pour:
- Performance (éviter validations coûteuses inutiles)
- UX (montrer une erreur à la fois)
- Validators dépendants (pas besoin de vérifier format email si champ vide)

#### 5.5.2 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Fail-fast au niveau schema
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(...).FailFast()

// ═══════════════════════════════════════════════════════════════
// Fail-fast au niveau champ (chain de validators)
// ═══════════════════════════════════════════════════════════════

poxxy.Field("email", &email).
    Validate(
        poxxy.Required(),           // Si fail, stop ici
        poxxy.Email(),              // Si fail, stop ici
        poxxy.UniqueInDB(checkFn),  // Coûteux, seulement si précédents OK
    ).
    FailFast() // Active pour ce champ

// ═══════════════════════════════════════════════════════════════
// Pipeline explicite (fail-fast par défaut)
// ═══════════════════════════════════════════════════════════════

poxxy.Field("email", &email).
    Pipeline(
        poxxy.Required(),
        poxxy.Email(),
        poxxy.UniqueInDB(checkFn),
    ) // Stop à la première erreur

// vs .Validate() qui collecte toutes les erreurs
poxxy.Field("email", &email).
    Validate(
        poxxy.Required(),
        poxxy.Email(),
        poxxy.MinLength(5),
    ) // Retourne toutes les erreurs
```

#### 5.5.3 Comportement

| Mode | Comportement |
|------|--------------|
| **Normal** (défaut) | Tous les validators s'exécutent, toutes les erreurs retournées |
| **FailFast schema** | Arrêt au premier champ en erreur |
| **FailFast field** | Arrêt au premier validator en erreur pour ce champ |
| **Pipeline** | Syntaxe explicite pour fail-fast par champ |

---

### 5.6 Simplified Slice API

#### 5.6.1 Description

Simplifier la validation des slices avec une API plus intuitive et moins de boilerplate.

#### 5.6.2 Problème actuel

```go
// V2 actuelle: callback verbeux
poxxy.Slice("tags", &tags, func(s *poxxy.Schema, tag *string) {
    s.Add(poxxy.Field("", tag).Validate(poxxy.MinLength(1), poxxy.MaxLength(50)))
})
```

#### 5.6.3 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Slices de types simples
// ═══════════════════════════════════════════════════════════════

var tags []string

poxxy.Slice("tags", &tags).
    // Validation sur la slice elle-même
    Validate(poxxy.MinItems(1), poxxy.MaxItems(10), poxxy.Unique()).
    // Validation sur chaque élément
    Each(poxxy.MinLength(1), poxxy.MaxLength(50), poxxy.Pattern(`^[a-z]+$`))

// ═══════════════════════════════════════════════════════════════
// Slices de structs
// ═══════════════════════════════════════════════════════════════

var items []OrderItem

poxxy.SliceOf("items", &items).
    Validate(poxxy.MinItems(1)).
    EachWith(func(s *poxxy.Schema, item *OrderItem) {
        s.Add(poxxy.Field("product_id", &item.ProductID).Validate(poxxy.Required()))
        s.Add(poxxy.Field("quantity", &item.Quantity).Validate(poxxy.Required(), poxxy.Min(1)))
        s.Add(poxxy.Field("price", &item.Price).Validate(poxxy.Required(), poxxy.Min(0.0)))
    })

// ═══════════════════════════════════════════════════════════════
// Avec struct tags (auto)
// ═══════════════════════════════════════════════════════════════

type OrderItem struct {
    ProductID int     `poxxy:"required"`
    Quantity  int     `poxxy:"required,min=1"`
    Price     float64 `poxxy:"required,min=0"`
}

poxxy.SliceOf("items", &items).
    Validate(poxxy.MinItems(1)).
    EachAuto() // Utilise les struct tags

// ═══════════════════════════════════════════════════════════════
// Validation avancée sur éléments
// ═══════════════════════════════════════════════════════════════

poxxy.Slice("scores", &scores).
    Each(poxxy.Min(0), poxxy.Max(100)).
    // Validators spéciaux pour slices
    ValidateAll(func(scores []int) error {
        sum := 0
        for _, s := range scores {
            sum += s
        }
        if sum > 500 {
            return fmt.Errorf("total score exceeds maximum")
        }
        return nil
    })
```

#### 5.6.4 Validators de slice intégrés

| Validator | Description |
|-----------|-------------|
| `MinItems(n)` | Minimum n éléments |
| `MaxItems(n)` | Maximum n éléments |
| `ExactItems(n)` | Exactement n éléments |
| `Unique()` | Tous éléments uniques |
| `UniqueBy(fn)` | Uniques selon clé |
| `Sorted()` | Éléments triés |
| `SortedDesc()` | Éléments triés décroissant |
| `Contains(v)` | Contient valeur |
| `NotContains(v)` | Ne contient pas valeur |

---

### 5.7 Discriminated Unions

#### 5.7.1 Description

Améliorer le support des types union (polymorphisme) avec une API claire pour les discriminated unions.

#### 5.7.2 Problème actuel

```go
// V2 actuelle: resolver manuel complexe
poxxy.Union("content", &content, func(data map[string]any) (any, error) {
    // Logique manuelle pour déterminer le type
    if t, ok := data["type"].(string); ok {
        switch t {
        case "text":
            return new(TextContent), nil
        case "image":
            return new(ImageContent), nil
        }
    }
    return nil, fmt.Errorf("unknown content type")
})
```

#### 5.7.3 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Discriminated union avec champ type
// ═══════════════════════════════════════════════════════════════

type Content interface{}
type TextContent struct {
    Text string `poxxy:"required,max=10000"`
}
type ImageContent struct {
    URL    string `poxxy:"required,url"`
    Width  int    `poxxy:"min=1"`
    Height int    `poxxy:"min=1"`
}

var content Content

poxxy.Union("content", &content).
    Discriminator("type"). // Champ qui indique le type
    Case("text", func() any { return new(TextContent) }).
    Case("image", func() any { return new(ImageContent) }).
    Default(func() any { return new(TextContent) }) // Optionnel

// ═══════════════════════════════════════════════════════════════
// Avec schemas explicites (sans tags)
// ═══════════════════════════════════════════════════════════════

poxxy.Union("content", &content).
    Discriminator("type").
    CaseSchema("text", func() any { return new(TextContent) },
        poxxy.NewSchema(
            poxxy.Field("text", &textContent.Text).Validate(poxxy.Required()),
        ),
    ).
    CaseSchema("image", func() any { return new(ImageContent) },
        poxxy.NewSchema(
            poxxy.Field("url", &imageContent.URL).Validate(poxxy.Required(), poxxy.URL()),
        ),
    )

// ═══════════════════════════════════════════════════════════════
// Union par inférence (essaie chaque type)
// ═══════════════════════════════════════════════════════════════

poxxy.Union("content", &content).
    Try(func() any { return new(TextContent) }, textSchema).
    Try(func() any { return new(ImageContent) }, imageSchema).
    Fallback(func() any { return new(GenericContent) })

// ═══════════════════════════════════════════════════════════════
// Union inline (types Go)
// ═══════════════════════════════════════════════════════════════

var value any // string | int | bool

poxxy.UnionOf("value", &value).
    Types(
        poxxy.StringType().Validate(poxxy.MinLength(1)),
        poxxy.IntType().Validate(poxxy.Min(0)),
        poxxy.BoolType(),
    )
```

#### 5.7.4 Comportement

1. **Discriminator**: Le champ discriminateur est extrait en premier
2. **Case matching**: Le case correspondant est sélectionné
3. **Schema application**: Le schema du case est appliqué
4. **Type assignment**: La valeur typée est assignée au pointeur union
5. **Erreur si no match**: Si aucun case ne matche et pas de default

---

### 5.8 Sanitizers vs Transformers

#### 5.8.1 Description

Séparer clairement les opérations de **sécurité** (sanitization) des **transformations métier**, avec des garanties d'exécution différentes.

#### 5.8.2 Différence conceptuelle

| Aspect | Sanitizer | Transformer |
|--------|-----------|-------------|
| **But** | Sécurité, nettoyage | Logique métier |
| **Exécution** | Toujours, avant tout | Après sanitization |
| **Erreur possible** | Non (best effort) | Oui (peut échouer) |
| **Exemples** | Strip HTML, escape SQL | Capitalize, format phone |

#### 5.8.3 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Chaîne explicite: Sanitize → Transform → Validate
// ═══════════════════════════════════════════════════════════════

poxxy.Field("bio", &bio).
    Sanitize(
        poxxy.StripTags(),           // Supprime HTML
        poxxy.TruncateUnsafe(5000),  // Limite taille (sécurité)
    ).
    Transform(
        poxxy.TrimSpace(),           // Nettoyage métier
        poxxy.NormalizeWhitespace(), // Un seul espace entre mots
    ).
    Validate(
        poxxy.MaxLength(1000),       // Règle métier
    )

// ═══════════════════════════════════════════════════════════════
// Sanitizers intégrés (sécurité)
// ═══════════════════════════════════════════════════════════════

poxxy.StripTags()              // Supprime toutes les balises HTML
poxxy.StripTagsExcept("b","i") // Garde certaines balises
poxxy.EscapeHTML()             // Échappe < > & " '
poxxy.EscapeSQL()              // Échappe quotes SQL
poxxy.SanitizeFilename()       // Supprime caractères dangereux
poxxy.TruncateUnsafe(n)        // Coupe à n caractères (sécurité DoS)
poxxy.NormalizeUnicode()       // Normalise unicode (prévient homoglyphes)
poxxy.StripNullBytes()         // Supprime \x00

// ═══════════════════════════════════════════════════════════════
// Configuration globale
// ═══════════════════════════════════════════════════════════════

// Applique des sanitizers par défaut à tous les strings
schema := poxxy.NewSchema(...).
    DefaultSanitizers(poxxy.StripNullBytes(), poxxy.NormalizeUnicode())
```

#### 5.8.4 Ordre d'exécution

```
Input → Sanitize → Transform → Assign → Validate
         (1)         (2)        (3)       (4)

1. Sanitizers: Toujours exécutés, ne peuvent pas échouer
2. Transformers: Peuvent échouer, erreur = arrêt
3. Assign: Assigne la valeur transformée
4. Validate: Vérifie les règles métier
```

---

### 5.9 Schema Composition

#### 5.9.1 Description

Permettre de réutiliser et composer des schemas pour éviter la duplication et faciliter la maintenance.

#### 5.9.2 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Schema fragments réutilisables
// ═══════════════════════════════════════════════════════════════

// Fragment définit des champs sans les lier à des variables
var timestampFields = poxxy.Fragment(
    poxxy.Field("created_at", (*time.Time)(nil)).
        Validate(poxxy.Required()).
        Default(time.Now),
    poxxy.Field("updated_at", (*time.Time)(nil)).
        Default(time.Now),
)

var auditFields = poxxy.Fragment(
    poxxy.Field("created_by", (*string)(nil)).Validate(poxxy.Required()),
    poxxy.Field("updated_by", (*string)(nil)),
)

// ═══════════════════════════════════════════════════════════════
// Composition de schemas
// ═══════════════════════════════════════════════════════════════

type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
    CreatedBy string
    UpdatedBy string
}

var user User

schema := poxxy.NewSchema(
    poxxy.Field("id", &user.ID).Validate(poxxy.Required()),
    poxxy.Field("name", &user.Name).Validate(poxxy.Required()),
    poxxy.Field("email", &user.Email).Validate(poxxy.Required(), poxxy.Email()),
).
    Include(timestampFields.BindTo(&user.CreatedAt, &user.UpdatedAt)).
    Include(auditFields.BindTo(&user.CreatedBy, &user.UpdatedBy))

// ═══════════════════════════════════════════════════════════════
// Héritage de schema
// ═══════════════════════════════════════════════════════════════

baseUserSchema := poxxy.NewSchema(
    poxxy.Field("email", &email).Validate(poxxy.Required(), poxxy.Email()),
    poxxy.Field("name", &name).Validate(poxxy.Required()),
)

createUserSchema := baseUserSchema.Extend(
    poxxy.Field("password", &password).Validate(poxxy.Required(), poxxy.MinLength(8)),
)

updateUserSchema := baseUserSchema.Extend(
    poxxy.Field("id", &id).Validate(poxxy.Required()),
).Without("password") // Retire un champ du parent

// ═══════════════════════════════════════════════════════════════
// Merge de schemas
// ═══════════════════════════════════════════════════════════════

fullSchema := poxxy.Merge(
    basicInfoSchema,
    addressSchema,
    preferencesSchema,
) // Combine tous les champs
```

#### 5.9.3 Résolution des conflits

```go
// Si même nom de champ dans plusieurs schemas
schema := poxxy.Merge(
    schemaA, // a "email" field
    schemaB, // also has "email" field
).OnConflict(poxxy.ConflictLast)     // Garde le dernier (défaut)
 .OnConflict(poxxy.ConflictFirst)    // Garde le premier
 .OnConflict(poxxy.ConflictError)    // Erreur si conflit
 .OnConflict(poxxy.ConflictMerge)    // Merge les validators
```

---

### 5.10 HTTP Ergonomics

#### 5.10.1 Description

Fournir des helpers HTTP pour simplifier l'intégration dans les applications web.

#### 5.10.2 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Helper handler avec validation automatique
// ═══════════════════════════════════════════════════════════════

// Définition du schema factory
func createUserSchema(u *CreateUserInput) *poxxy.Schema {
    return poxxy.NewSchema(
        poxxy.Field("email", &u.Email).Validate(poxxy.Required(), poxxy.Email()),
        poxxy.Field("name", &u.Name).Validate(poxxy.Required()),
        poxxy.Field("age", &u.Age).Validate(poxxy.Min(18)),
    )
}

// Handler wrapper
http.HandleFunc("/users", poxxy.HandleJSON(createUserSchema, func(w http.ResponseWriter, r *http.Request, input CreateUserInput) {
    // input est déjà validé ici
    user := createUser(input)
    json.NewEncoder(w).Encode(user)
}))

// ═══════════════════════════════════════════════════════════════
// Avec options de configuration
// ═══════════════════════════════════════════════════════════════

poxxy.HandleJSON(createUserSchema, handler,
    poxxy.WithMaxBodySize(1<<20),           // 1MB
    poxxy.WithErrorHandler(customErrorHandler),
    poxxy.WithSuccessStatus(http.StatusCreated),
)

// ═══════════════════════════════════════════════════════════════
// Error handler personnalisé
// ═══════════════════════════════════════════════════════════════

customErrorHandler := func(w http.ResponseWriter, errs *poxxy.ValidationErrors) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnprocessableEntity)
    json.NewEncoder(w).Encode(map[string]any{
        "success": false,
        "errors":  errs.ToMap(),
    })
}

// ═══════════════════════════════════════════════════════════════
// Middleware style
// ═══════════════════════════════════════════════════════════════

func ValidateMiddleware[T any](schemaFn func(*T) *poxxy.Schema) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var input T
            schema := schemaFn(&input)
            if err := poxxy.ParseHTTPRequest(schema, r); err != nil {
                poxxy.WriteError(w, err)
                return
            }
            ctx := context.WithValue(r.Context(), "validated", input)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Utilisation avec chi/gorilla/etc
r.With(ValidateMiddleware(createUserSchema)).Post("/users", createUserHandler)

// ═══════════════════════════════════════════════════════════════
// Support multipart/form-data
// ═══════════════════════════════════════════════════════════════

type UploadInput struct {
    Title       string
    Description string
    File        *poxxy.UploadedFile // Type spécial pour fichiers
}

schema := poxxy.NewSchema(
    poxxy.Field("title", &input.Title).Validate(poxxy.Required()),
    poxxy.Field("description", &input.Description),
    poxxy.File("file", &input.File).
        Validate(
            poxxy.MaxFileSize(10<<20),              // 10MB
            poxxy.AllowedMimeTypes("image/jpeg", "image/png"),
            poxxy.AllowedExtensions(".jpg", ".png"),
        ),
)
```

#### 5.10.3 Types HTTP spéciaux

```go
// UploadedFile représente un fichier uploadé
type UploadedFile struct {
    Filename    string
    Size        int64
    ContentType string
    Reader      io.Reader
    Header      *multipart.FileHeader
}

// Methods
file.Save("/path/to/destination")
file.Bytes() ([]byte, error)
file.String() (string, error)
```

---

### 5.11 Async Validation

#### 5.11.1 Description

Permettre des validations asynchrones (requêtes DB, appels API) avec gestion du contexte et timeout.

#### 5.11.2 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Validation async simple
// ═══════════════════════════════════════════════════════════════

poxxy.Field("email", &email).
    Validate(poxxy.Required(), poxxy.Email()).
    ValidateAsync(func(ctx context.Context, email string) error {
        exists, err := db.EmailExists(ctx, email)
        if err != nil {
            return fmt.Errorf("database error: %w", err)
        }
        if exists {
            return fmt.Errorf("email already registered")
        }
        return nil
    })

// ═══════════════════════════════════════════════════════════════
// Parse avec context
// ═══════════════════════════════════════════════════════════════

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := schema.ParseContext(ctx, data)

// ═══════════════════════════════════════════════════════════════
// Configuration async
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(...).
    AsyncTimeout(3 * time.Second).       // Timeout par défaut
    AsyncConcurrency(10)                  // Max validations parallèles

// ═══════════════════════════════════════════════════════════════
// Validators async prédéfinis
// ═══════════════════════════════════════════════════════════════

poxxy.UniqueInDB(func(ctx context.Context, email string) (bool, error) {
    return db.EmailExists(ctx, email)
})

poxxy.ExistsInDB(func(ctx context.Context, userID int) (bool, error) {
    return db.UserExists(ctx, userID)
})

poxxy.RemoteValidation("https://api.example.com/validate")
```

#### 5.11.3 Ordre d'exécution

```
1. Sanitize (sync)
2. Transform (sync)
3. Assign (sync)
4. Validate sync (tous les validators sync)
5. Validate async (en parallèle, avec context)
```

Les validations async ne s'exécutent que si les validations sync passent.

---

### 5.12 Error Formatting

#### 5.12.1 Description

Permettre des formats d'erreur personnalisables et compatibles avec différents standards.

#### 5.12.2 API proposée

```go
// ═══════════════════════════════════════════════════════════════
// Formats prédéfinis
// ═══════════════════════════════════════════════════════════════

// Format simple (map field -> messages)
errs.ToMap()
// {"email": ["is required", "invalid format"], "age": ["must be >= 18"]}

// Format flat (liste d'erreurs)
errs.ToList()
// [{"field": "email", "message": "is required"}, ...]

// Format JSON:API (https://jsonapi.org/format/#errors)
errs.ToJSONAPI()
// {"errors": [{"source": {"pointer": "/data/attributes/email"}, "detail": "is required"}]}

// Format RFC 7807 Problem Details
errs.ToProblemDetails()
// {"type": "validation-error", "title": "Validation Failed", "errors": [...]}

// ═══════════════════════════════════════════════════════════════
// Formatter personnalisé
// ═══════════════════════════════════════════════════════════════

schema := poxxy.NewSchema(...).
    WithErrorFormatter(func(errs *poxxy.ValidationErrors) any {
        return map[string]any{
            "success": false,
            "code":    "VALIDATION_ERROR",
            "errors":  errs.GroupByField(),
        }
    })

// ═══════════════════════════════════════════════════════════════
// HTTP response helpers
// ═══════════════════════════════════════════════════════════════

// Écrit automatiquement la réponse HTTP
poxxy.WriteErrors(w, errs, http.StatusBadRequest)
poxxy.WriteErrorsJSON(w, errs) // JSON avec status 422
poxxy.WriteErrorsJSONAPI(w, errs)
poxxy.WriteErrorsProblemDetails(w, errs)

// ═══════════════════════════════════════════════════════════════
// Accès aux métadonnées d'erreur
// ═══════════════════════════════════════════════════════════════

for _, err := range errs.All() {
    fmt.Println(err.Field)      // "email"
    fmt.Println(err.Rule)       // "email"
    fmt.Println(err.Message)    // "invalid email format"
    fmt.Println(err.Params)     // map[string]any{}
    fmt.Println(err.Value)      // "invalid@"
}
```

#### 5.12.3 Structure ValidationError enrichie

```go
type ValidationError struct {
    Field       string            // Nom du champ ("email", "user.address.city")
    Rule        string            // Règle qui a échoué ("required", "min", "email")
    Message     string            // Message d'erreur localisé
    Params      map[string]any    // Paramètres de la règle ({"min": 5})
    Value       any               // Valeur qui a échoué validation
    Path        []string          // Chemin complet ["user", "address", "city"]
    Index       *int              // Index si dans une slice
    Condition   string            // Condition si champ conditionnel
}
```

---

## 6. Priorités et Roadmap

### 6.1 Matrice de priorisation

| Feature | Impact DX | Complexité | Demande | Priorité |
|---------|-----------|------------|---------|----------|
| Cross-field validation | 🔴 Haute | 🟡 Moyenne | 🔴 Haute | **P0** |
| Conditional fields | 🔴 Haute | 🟡 Moyenne | 🔴 Haute | **P0** |
| Struct tags | 🔴 Haute | 🟡 Moyenne | 🟡 Moyenne | **P1** |
| Simplified slice API | 🟡 Moyenne | 🟢 Basse | 🟡 Moyenne | **P1** |
| Fail-fast mode | 🟡 Moyenne | 🟢 Basse | 🟡 Moyenne | **P1** |
| Schema composition | 🟡 Moyenne | 🟡 Moyenne | 🟡 Moyenne | **P1** |
| HTTP ergonomics | 🟡 Moyenne | 🟢 Basse | 🟡 Moyenne | **P2** |
| Sanitizers separation | 🟡 Moyenne | 🟢 Basse | 🟢 Basse | **P2** |
| Error formatting | 🟡 Moyenne | 🟢 Basse | 🟡 Moyenne | **P2** |
| Async validation | 🟡 Moyenne | 🟡 Moyenne | 🟢 Basse | **P2** |
| Validation groups | 🟢 Basse | 🟡 Moyenne | 🟢 Basse | **P3** |
| Discriminated unions | 🟢 Basse | 🟡 Moyenne | 🟢 Basse | **P3** |

### 6.2 Phases de développement

```
┌─────────────────────────────────────────────────────────────────┐
│                        PHASE 1 - Core                           │
│                      (Fondations critiques)                     │
├─────────────────────────────────────────────────────────────────┤
│ • Cross-field validation                                        │
│ • Conditional fields                                            │
│ • Documentation et migration guide                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PHASE 2 - Productivity                      │
│                    (Réduction boilerplate)                      │
├─────────────────────────────────────────────────────────────────┤
│ • Struct tags auto-schema                                       │
│ • Simplified slice API                                          │
│ • Fail-fast mode                                                │
│ • Schema composition                                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PHASE 3 - Integration                       │
│                    (Écosystème et polish)                       │
├─────────────────────────────────────────────────────────────────┤
│ • HTTP ergonomics                                               │
│ • Sanitizers separation                                         │
│ • Error formatting                                              │
│ • Async validation                                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PHASE 4 - Advanced                         │
│                      (Use cases avancés)                        │
├─────────────────────────────────────────────────────────────────┤
│ • Validation groups                                             │
│ • Discriminated unions améliorés                                │
│ • Performance optimizations                                     │
│ • Additional validators                                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 7. Métriques de Succès

### 7.1 Métriques quantitatives

| Métrique | Baseline (Main) | Target (V2) | Méthode de mesure |
|----------|-----------------|-------------|-------------------|
| Lignes de code par schema | 100 LOC (10 champs) | 60 LOC (-40%) | Benchmark examples |
| Temps d'apprentissage | 2h | 30min | Survey nouveaux users |
| Couverture de tests | 75% | >90% | go test -cover |
| Performance (50 champs) | 0.5ms | <0.5ms | Benchmark |
| Issues "confusion API" | 15/mois | <3/mois | GitHub issues |

### 7.2 Métriques qualitatives

| Aspect | Critère de succès |
|--------|-------------------|
| **Lisibilité** | Un schema peut être compris sans documentation |
| **Découvrabilité** | Autocomplete guide vers les bonnes méthodes |
| **Prévisibilité** | Comportement évident, pas de surprises |
| **Debuggabilité** | Erreurs pointent vers la cause exacte |

### 7.3 Critères d'acceptation par feature

#### Cross-field validation
- [ ] `EqualTo`, `GreaterThan`, `After` fonctionnent avec références
- [ ] Erreurs indiquent le champ référencé
- [ ] Tests couvrent: même type, types différents, absent, null

#### Conditional fields
- [ ] `.When()` accepte conditions et callbacks
- [ ] Combinateurs `And`, `Or`, `Not` fonctionnent
- [ ] Champ conditionnel absent ne génère pas d'erreur

#### Struct tags
- [ ] Tous les validators de base supportés
- [ ] Override manuel fonctionne
- [ ] Cache des tags pour performance
- [ ] Erreurs de parsing tags claires

---

## 8. Risques et Mitigations

### 8.1 Risques techniques

| Risque | Probabilité | Impact | Mitigation |
|--------|-------------|--------|------------|
| Breaking changes trop nombreux | 🟡 Moyenne | 🔴 Haute | Migration guide détaillé, codemods |
| Performance dégradée | 🟢 Basse | 🟡 Moyenne | Benchmarks CI, profiling régulier |
| Complexité accrue du code | 🟡 Moyenne | 🟡 Moyenne | Code review stricte, documentation interne |
| Incompatibilité i18n | 🟢 Basse | 🟡 Moyenne | Tests avec toutes les langues supportées |

### 8.2 Risques produit

| Risque | Probabilité | Impact | Mitigation |
|--------|-------------|--------|------------|
| Faible adoption V2 | 🟡 Moyenne | 🔴 Haute | Communication claire des bénéfices |
| Migration douloureuse | 🟡 Moyenne | 🔴 Haute | Outils de migration automatique |
| Feature creep | 🟡 Moyenne | 🟡 Moyenne | Scope strict par phase |

### 8.3 Plan de rollback

Si une feature cause des problèmes majeurs:
1. Feature flag pour désactiver
2. Patch release sans la feature
3. Post-mortem et redesign

---

## 9. Annexes

### 9.1 Glossaire

| Terme | Définition |
|-------|------------|
| **Schema** | Définition de la structure et des règles de validation |
| **Field** | Un champ dans un schema, avec type, validators, transformers |
| **Validator** | Fonction qui vérifie une contrainte sur une valeur |
| **Transformer** | Fonction qui modifie une valeur (normalisation) |
| **Sanitizer** | Transformer orienté sécurité, toujours appliqué |
| **Cross-field** | Validation impliquant plusieurs champs |
| **Discriminated union** | Type union avec champ indiquant le type concret |

### 9.2 Références

- [JSON Schema Specification](https://json-schema.org/)
- [OpenAPI 3.1 Specification](https://spec.openapis.org/oas/v3.1.0)
- [RFC 7807 - Problem Details for HTTP APIs](https://tools.ietf.org/html/rfc7807)
- [JSON:API Error Format](https://jsonapi.org/format/#errors)
- [go-playground/validator](https://github.com/go-playground/validator) - Inspiration pour struct tags

### 9.3 Comparaison avec alternatives Go

| Feature | Poxxy V2 | go-validator | ozzo-validation |
|---------|----------|--------------|-----------------|
| Generics | ✅ | ❌ | ❌ |
| Type-safe | ✅ | ❌ (reflect) | ❌ (reflect) |
| Cross-field | ✅ | ✅ | Partiel |
| i18n | ✅ | ✅ | ❌ |
| Struct tags | ✅ | ✅ | ❌ |
| JSON Schema | ✅ | ❌ | ❌ |
| HTTP helpers | ✅ | ❌ | ❌ |
| Async | ✅ | ❌ | ❌ |

### 9.4 Exemple complet - Vision finale

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/arkan/poxxy"
)

// Modèles avec struct tags
type CreateOrderRequest struct {
    CustomerEmail string          `poxxy:"required,email,transform=sanitize_email"`
    Items         []OrderItem     `poxxy:"required,min_items=1,max_items=100"`
    DeliveryType  string          `poxxy:"required,in=pickup|shipping"`
    ShippingAddr  *ShippingAddress `poxxy:"required_when=DeliveryType:shipping"`
    PromoCode     string          `poxxy:"pattern=^[A-Z]{4}\\d{4}$,optional"`
    Notes         string          `poxxy:"max=1000,transform=trim,sanitize=strip_tags"`
}

type OrderItem struct {
    ProductID int     `poxxy:"required,exists=products"`
    Quantity  int     `poxxy:"required,min=1,max=100"`
    Price     float64 `poxxy:"required,min=0"`
}

type ShippingAddress struct {
    Street  string `poxxy:"required,min=5,max=200"`
    City    string `poxxy:"required"`
    ZipCode string `poxxy:"required,pattern=^\\d{5}$"`
    Country string `poxxy:"required,in=FR|BE|CH"`
}

func main() {
    http.HandleFunc("/orders", poxxy.HandleJSON(orderSchema, createOrder,
        poxxy.WithErrorFormat(poxxy.JSONAPIFormat),
        poxxy.WithTranslator(poxxy.French()),
    ))
    http.ListenAndServe(":8080", nil)
}

func orderSchema(req *CreateOrderRequest) *poxxy.Schema {
    return poxxy.NewSchema(
        poxxy.StructAuto("", req),
    ).
        CrossValidate(func(s *poxxy.Schema) error {
            // Validation métier complexe
            total := 0.0
            for _, item := range req.Items {
                total += item.Price * float64(item.Quantity)
            }
            if req.PromoCode != "" && total < 50 {
                return poxxy.FieldError("promo_code", "minimum order €50 for promo codes")
            }
            return nil
        }).
        FailFast()
}

func createOrder(w http.ResponseWriter, r *http.Request, req CreateOrderRequest) {
    // req est validé, typé, transformé, sanitizé
    // Logique métier ici...
}
```

---

**Document Version:** 1.0
**Last Updated:** 2026-01-17
**Status:** Ready for Review
