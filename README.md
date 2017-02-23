#Production Readu Golang HTTP.Handler for graphl-go

**Notes:**

This GraphQL Handler is compatible only with `https://github.com/playlyfe/go-graphql` implementation. 
Usage

**Usage:**
``` ruby
	
	package main
	
	import (
	    "net/http"
	    "github.com/krypton97/HandleGraphQL"
	    "github.com/playlyfe/go-graphql"
	)

	func main() {
		schema := `
	    interface Pet {
		name: String
	    }
	    type Dog implements Pet {
		name: String
		woofs: Boolean
	    }
	    type Cat implements Pet {
		name: String
		meows: Boolean
	    }
	    type QueryRoot {
		pets: [Pet]
	    }
    `
	resolvers := map[string]interface{}{}
	resolvers["QueryRoot/pets"] = func(params *graphql.ResolveParams) (interface{}, error) {
		return []map[string]interface{}{
			{
				"__typename": "Dog",
				"name":       "Odie",
				"woofs":      true,
			},
			{
				"__typename": "Cat",
				"name":       "Garfield",
				"meows":      false,
			},
		}, nil
	}

	executor, err := graphql.NewExecutor(schema, "QueryRoot", "", resolvers)
	executor.ResolveType = func(value interface{}) string {
		if object, ok := value.(map[string]interface{}); ok {
			return object["__typename"].(string)
		}
		return ""
	}

	if err != nil {
		panic(err)
	}

	api := handler.New(&handler.Config{
		Executor: executor,
		Context:  "",
		Pretty:   true,
	})

	http.Handle("/graphql", api)
	http.ListenAndServe(":3000", nil)
}

```
**Details**

The handler will accept requests with the parameters:

***`query`***:    A string GraphQL document to be executed.

***`variables`***: The runtime values to use for any GraphQL query variables as a JSON object.

**`*operationName*`**: If the provided query contains multiple named operations, this specifies which operation should be executed. If not provided, an 400 error will be returned if the query contains multiple named operations.

GraphQL will first look for each parameter in the URL's query-string:

`/graphql?query=query+getUser($id:ID){user(id:$id){name}}&variables={"id":"4"}`

If not found in the query-string, it will look in the POST request body. The handler will interpret it depending on the provided Content-Type header.

**`application/json`**: the POST body will be parsed as a JSON object of parameters.

**`application/x-www-form-urlencoded:`** this POST body will be parsed as a url-encoded string of key-value pairs.

**`application/graphql`**: The POST body will be parsed as GraphQL query string, which provides the query parameter.


-----------------------------------------------------------------

**Credits**

This handler is a custom version of the *`https://github.com/graphql-go/handler`* handler that integrates with `https://github.com/playlyfe/go-graphql`.  For contributions, make a pull request.
