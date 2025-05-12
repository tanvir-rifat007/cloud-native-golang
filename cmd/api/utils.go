package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)


type envelope map[string]any



func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {

    js, err := json.MarshalIndent(data, "", "\t")
    if err != nil {
        return err
    }

    js = append(js, '\n')

    for key, value := range headers {
        w.Header()[key] = value
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(js)

    return nil
}


func (app *application) readJSON(w http.ResponseWriter, r *http.Request,dst any)error{
	// limit the req body to 1mb:
	maxBytes:= 1_048_576;
	r.Body = http.MaxBytesReader(w,r.Body,int64(maxBytes))
	dec:= json.NewDecoder(r.Body)

	// this will make sure that the json decoder returns an error if the request body contains any additional fields which cannot be mapped to the target destination
	dec.DisallowUnknownFields()

	err:= dec.Decode(dst)


	if err!=nil{
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError


		switch{
		case errors.As(err,&syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)",syntaxError.Offset)

		case errors.Is(err,io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
            if unmarshalTypeError.Field != "" {
                return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
            }
            return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)


    case errors.Is(err, io.EOF):
            return errors.New("body must not be empty")

	   case strings.HasPrefix(err.Error(), "json: unknown field "):
            fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
            return fmt.Errorf("body contains unknown key %s", fieldName)

        // Use the errors.As() function to check whether the error has the type 
        // *http.MaxBytesError. If it does, then it means the request body exceeded our 
        // size limit of 1MB and we return a clear error message.
        case errors.As(err, &maxBytesError):
            return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)



    case errors.As(err, &invalidUnmarshalError):
            panic(err)

    default:
            return err
		}


	}

	// if there is any additional JSON data in the request body, return an error
// like : curl -d '{"title": "Moana"}{"title": "Top Gun"}' localhost:4000/v1/movies

// here for this 2nd one {"title": "Top Gun"} we will get an error
	 err = dec.Decode(&struct{}{})
    if !errors.Is(err, io.EOF) {
        return errors.New("body must only contain a single JSON value")
    }
return nil
}


func (app *application) catchAllClientRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the client application
	http.ServeFile(w,r,"./public/index.html")


}

func (app *application) readIDParam(r *http.Request) (int, error) {
    idString := chi.URLParam(r,"id")
    id, err := strconv.Atoi(idString)
    if err != nil {
        return 0, fmt.Errorf("invalid id parameter: %s", idString)
    }
    return id, nil
}


