package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	sugar "github.com/btubbs/graphql-sugar"
	"github.com/graphql-go/graphql"
	handler "github.com/graphql-go/handler"
)

func main() {

	mutationType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"saveUser": &graphql.Field{
					Type:    userType,
					Args:    sugar.ArgsConfig(user{}),
					Resolve: resolveSaveUser,
				},
			},
		})

	schema, _ := graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		},
	)

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

type user struct {
	ID               string     `json:"id" arg:"id,required" desc:"A short identifier for this user."`
	Name             string     `json:"name" arg:"name,required" desc:"This user's name."`
	JoinedAt         time.Time  `json:"joinedAt"`
	NumberOfChildren int        `json:"numberOfChildren" arg:"numberOfChildren" desc:"The number of children that this user has."`
	FavoriteMovies   moviesList `json:"favoriteMovies" arg:"favoriteMovies" desc:"A JSON-formatted list of this user's favorite movies."`
}

var userType = sugar.OutputType("User", "A user, dummy", user{})

var data = map[string]user{
	"bob": {
		ID:   "bob",
		Name: "Bob Loblaw",
		FavoriteMovies: []string{
			"The Shawshank Redemption",
			"Weekend at Bernie's 2",
		},
		NumberOfChildren: 7,
		JoinedAt:         time.Date(2012, time.February, 3, 9, 19, 38, 4213, time.UTC),
	},
}

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: resolveUser,
			},
		},
	})

func resolveUser(p graphql.ResolveParams) (interface{}, error) {
	argID, ok := p.Args["id"].(string)
	if !ok {
		return nil, errors.New("id is not a string")
	}
	if user, ok := data[argID]; !ok {
		return nil, fmt.Errorf("user %s not found", argID)
	} else {
		return user, nil
	}
}

type moviesList []string

func init() {
	parseMoviesList := func(arg interface{}) (moviesList, error) {
		if ms, ok := arg.([]interface{}); ok {
			out := moviesList{}
			for _, m := range ms {
				out = append(out, m.(string))
			}
			return out, nil
		}
		return nil, errors.New("invalid movie list")
	}
	if err := sugar.RegisterArgParser(parseMoviesList, graphql.NewList(graphql.String)); err != nil {
		panic(err)
	}
}

func resolveSaveUser(p graphql.ResolveParams) (interface{}, error) {
	u := user{}
	if err := sugar.LoadArgs(p, &u); err != nil {
		return nil, err
	}
	u.JoinedAt = time.Now()
	data[u.ID] = u
	return u, nil
}
