package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Street struct {
	Postalcode string `json:"Postalcode"`
	Name       string `json:"Name"`
	District   string `json:"District"`
	City       string `json:"City"`
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Whizr addresses micro service")
}

func handler(w http.ResponseWriter, r *http.Request) {

	postalcode := mux.Vars(r)["postalcode"]

	var street = find(postalcode)

	if street.Name == "" {
		street = getLocation(postalcode)
		store(street)
	}

	fmt.Fprint(w, street)
}

func find(postalcode string) Street {

	streetsCollection := conectionDatabase("streets")

	var streetDB Street

	streetsCollection.FindOne(context.TODO(), bson.D{{"postalcode", postalcode}}).Decode(&streetDB)

	return streetDB
}

func store(street Street) Street {

	streetsCollection := conectionDatabase("streets")

	streetsCollection.InsertOne(context.TODO(), street)

	return street
}

func getLocation(postalcode string) Street {
	response, err := http.Get("http://location.melhorenvio.com/" + postalcode)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	searchedStreet := map[string]string{}

	json.Unmarshal([]byte(responseData), &searchedStreet)

	street := Street{
		Postalcode: postalcode,
		Name:       searchedStreet["logradouro"],
		District:   searchedStreet["bairro"],
		City:       searchedStreet["cidade"]}

	return street
}

func conectionDatabase(collection string) *mongo.Collection {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	return client.Database("whizr").Collection(collection)

}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", home)
	router.HandleFunc("/{postalcode}", handler)
	log.Fatal(http.ListenAndServe(":8080", router))

}
