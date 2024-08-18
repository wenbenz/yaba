package handlers

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

var ping = &graphql.Field {
	Type: graphql.String,
	Name: "Ping",
	Description: "Responds 'pong' to ping",
	Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
		return "pong", nil
	},
}

func CreateGraphqlSchema() (*graphql.Schema, error) {
	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "RootQuery",
			Description: "Root query for GraphQL requests whose fields are query paths.",
			Fields: graphql.Fields{
				"ping": ping,
				// add new paths here
			},
		}),
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create graphql schema: %w", err)
	}

	return &schema, nil
}
