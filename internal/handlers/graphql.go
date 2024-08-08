package handlers

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

func CreateGraphqlSchema() (*graphql.Schema, error) {
	fields := graphql.Fields{
		"ping": &graphql.Field{
			Type: graphql.String,
			Name: "Ping",
			Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
				return "pong", nil
			},
		},
	}

	rootQuery := graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: fields,
	}

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create graphql schema: %w", err)
	}

	return &schema, nil
}
