package main

// Categories — liste fermée des catégories d'échange autorisées.
// Doit rester synchronisée avec la contrainte CHECK dans
// migrations/001_schema.sql : toute désynchronisation provoquerait des INSERT
// SQL en erreur (23514 check_violation) mal mappés.
var Categories = []string{
	"Informatique",
	"Jardinage",
	"Bricolage",
	"Cuisine",
	"Musique",
	"Langues",
	"Sport",
	"Tutorat",
	"Demenagement",
	"Photographie",
	"Animalier",
	"Couture",
	"Autre",
}

// IsValidCategorie renvoie true si c fait partie de la liste fermée.
func IsValidCategorie(c string) bool {
	for _, v := range Categories {
		if v == c {
			return true
		}
	}
	return false
}
