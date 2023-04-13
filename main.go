package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

//go:embed index.html
var html embed.FS

//go:embed index.js
var js embed.FS

type app struct {
	client *mongo.Client
}

func run(args []string) error {
	// flags
	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Int("port", 27017, "port")
	user := flag.String("user", "", "user")
	password := flag.String("password", "", "password")
	auth := flag.String("auth-source", "", "auth source")
	serverPort := flag.String("server-port", "8080", "server port")
	flag.Parse()

	// create uri
	uri := "mongodb://"
	if *user != "" && *password != "" {
		uri += *user + ":" + *password
	}
	uri += "@" + *host + ":" + strconv.Itoa(*port)

	if *auth != "" {
		uri += "/?authSource=" + *auth
	}

	// connect db
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalln(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatalln(err)
		}
	}()
	app := &app{
		client: client,
	}

	// route
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.FS(html))).Methods("GET")
	r.Handle("/index.js", http.FileServer(http.FS(js))).Methods("GET")
	r.HandleFunc("/dbs", app.dbs).Methods("GET")
	r.HandleFunc("/dbs/{db}/collections", app.collections).Methods("GET")
	r.HandleFunc("/dbs/{db}/collections", app.collections).Methods("GET")
	r.HandleFunc("/dbs/{db}/collections/{collection}/documents", app.documents).Methods("GET")

	return http.ListenAndServe(":"+*serverPort, r)
}

func (a *app) dbs(w http.ResponseWriter, r *http.Request) {
	dbs, err := a.client.ListDatabaseNames(r.Context(), bson.D{})
	if err != nil {
		log.Println(err)
		return
	}
	res := map[string]any{"items": dbs}
	body, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (a *app) collections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := a.client.Database(vars["db"])
	collections, err := db.ListCollectionNames(r.Context(), bson.D{})
	if err != nil {
		log.Println(err)
		return
	}
	res := map[string]any{"items": collections}
	body, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (a *app) documents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collection := a.client.Database(vars["db"]).Collection(vars["collection"])
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		log.Println(err)
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		log.Println(err)
		return
	}
	l := int64(limit)
	var skip int64 = (int64(page) - 1) * l
	opt := options.FindOptions{
		Limit: &l,
		Skip:  &skip,
	}
	filter := primitive.E{}
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	if key != "" && value != "" {
		filter.Key = key
		if v, err := strconv.Atoi(value); err != nil {
			filter.Value = value
		} else {
			filter.Value = v
		}
	}

	cursor, err := collection.Find(r.Context(), bson.D{filter}, &opt)
	if err != nil {
		log.Println(err)
		return
	}
	results := make([]bson.M, 0)
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Println(err)
		return
	}
	res := map[string]any{"items": results}
	body, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
