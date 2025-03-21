package models

type MovieMedia struct {
    Cover      *string  `json:"cover"`
    Screenshots []string `json:"screenshots"`
}