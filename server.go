package main

import (
	"context"
	// "crypto/x509"
	"encoding/json"
	// "errors"
	"fmt"
	// "io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Server struct {
	Router *mux.Router
	Context context.Context
	Client *mongo.Client
	DriversCollectoion *mongo.Collection
	AttendeesCollection *mongo.Collection
	mongoServer string
}
type IDSearch struct {
	ID   primitive.ObjectID `bson:"_id" json:"id"`
	Type string `json:"type"`
}

type Participant struct {
	ID   primitive.ObjectID `bson:"_id" json:"id"`
	Name string             `json:"name"`
	Address string          `json:"address"`
	City string             `json:"city"`
	State string            `json:"state"`
	Zip string              `json:"zip"`
	CellPhone string        `json:"cellphone"`
	HomePhone string        `json:"homephone"`
	Email string            `json:"email"`
	Location Location				`json:"location"`
	Notes string            `json:"notes"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (s *Server) Initialize() error {
	// rootPEM, err := ioutil.ReadFile("./combined")
	// if err != nil {
	// 	return err
	// }
	// roots := x509.NewCertPool()
	// ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	// if !ok {
	// 	return errors.New("failed to parse root certificate")
	// }
	s.Router = mux.NewRouter().StrictSlash(true)
	s.mongoServer = os.Getenv("MONGO_SERVER")
	fmt.Println("MONGO_SERVER: ", s.mongoServer)

	return nil
}

func (s *Server) InitializeRoutes() {
	// s.Router.HandleFunc("/", s.Home).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/{type}", s.getType).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/{type}/{id}", s.getIndividual).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/api/delete/{type}/{id}", s.deleteIndividual).Methods("POST", "OPTIONS")
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	s.Router.Use(cors)
}

func (s *Server) Run(addr string) {
	fmt.Println("Running ... ", addr)
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	s.Router.Use(cors)
	fs := http.FileServer(http.Dir("./static"))
	s.Router.Handle("/", fs)
	log.Fatal(http.ListenAndServe(addr, cors(s.Router)))
	// log.Fatal(http.ListenAndServeTLS(addr, "./combined", "./combined", cors(s.Router)))
}

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
// type spaHandler struct {
// 	staticPath string
// 	indexPath  string
// }

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "./index.html")
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be serves. If not, the
// file located at the index path on the SPA handler will be serves. This
// is suitable behavior for serving an SPA (single page application).
// func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	// get the absolute path to prevent directory traversal
// 	path, err := filepath.Abs(r.URL.Path)
// 	if err != nil {
// 		// if we failed to get the absolute path respond with a 400 bad request
// 		// and stop
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// prepend the path with the path to the static directory
// 	//path = filepath.Join(h.staticPath, path)

// 	// check whether a file exists at the given path
// 	_, err = os.Stat(path)
// 	if os.IsNotExist(err) {
// 		// file does not exist, serve index.html
// 		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
// 		return
// 	} else if err != nil {
// 		// if we got an error (that wasn't that the file doesn't exist) stating the
// 		// file, return a 500 internal server error and stop
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// otherwise, use http.FileServer to serve the static dir
// 	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
// }

/* get one participant
 * @param collection - mongo collection to search
 * @param filter - id of participant
 * @param ctx - context
*/
func (s *Server) getOne(query map[string]string) (primitive.M, error) {
	var result primitive.M
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.mongoServer))
	if err != nil {
		log.Fatal(err)
	}
	if client != nil {
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connected to MongoDB!")
		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				panic(err)
			}
		}()
	}
	var collection *mongo.Collection
	if strings.ToLower(query["type"]) == "drivers" {
		collection = client.Database("Blind").Collection("Drivers")
	} else if strings.ToLower(query["type"]) == "attendees" {
		collection = client.Database("Blind").Collection("Attendees")
	}
	var search = IDSearch{}
	search.ID, _ = primitive.ObjectIDFromHex(query["id"])
 	filter := bson.D{{Key: "_id", Value: search.ID}}
	// 	if strings.ToLower(vars["type"]) == "drivers" {
	// 		part, err := getOne(driversCollection, filter, ctx)
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

/* get all participants
 * @param collection - mongo collection to search
 * @param ctx - context
*/
func (s *Server) getAll(qType string) ([]primitive.M, error) {
	fmt.Println("getAll")
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.mongoServer))
	if err != nil {
		fmt.Println("Error connecting to MongoDB: ", err)
		return nil, err
	}
	if client != nil {
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			return nil, err
		}
		fmt.Println("Connected to MongoDB!")
		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				return
			}
		}()
	}
	var collection *mongo.Collection
	var results []primitive.M
	if strings.ToLower(qType) == "drivers" {
		collection = client.Database("Blind").Collection("Drivers")
	} else if strings.ToLower(qType) == "attendees" {
		collection = client.Database("Blind").Collection("Attendees")
	}
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
  if err := cur.All(s.Context, &results); err != nil {
    return nil, err
  }
	cur.Close(s.Context)
	return results, nil
}


	// get the whole group
	func (s *Server) getType(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		fmt.Println(vars["type"])
		if vars["type"] == "undefined" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		results, err := s.getAll(vars["type"])
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			json.NewEncoder(w).Encode(results)
	}

	func (s *Server) deleteIndividual(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		fmt.Println(vars["type"])
		w.Header().Set("Content-Type", "application/json")
		results, err := s.deleteOne(vars)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			json.NewEncoder(w).Encode(results)
	}

	func (s *Server) getIndividual(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		fmt.Println(vars["type"])
		w.Header().Set("Content-Type", "application/json")
		results, err := s.getOne(vars)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			json.NewEncoder(w).Encode(results)
	}
	// get a person
	// router.Methods("GET").Path("/api/{type}/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	fmt.Println(vars["type"])
	// 	w.Header().Set("Content-Type", "application/json")
	// 	var search = IDSearch{}
	// 	search.ID, _ = primitive.ObjectIDFromHex(vars["id"])
	// 	filter := bson.D{{Key: "_id", Value: search.ID}}
	// 	if strings.ToLower(vars["type"]) == "drivers" {
	// 		part, err := getOne(driversCollection, filter, ctx)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			w.Write([]byte(`{"message": "Error"}`))
	// 			return
	// 		}
	// 		foo, err := json.Marshal(part)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			w.Write([]byte(`{"message": "Error"}`))
	// 			return
	// 		}
	// 		fmt.Println(string(foo))
	// 		json.NewEncoder(w).Encode(part)
	// 	} else if strings.ToLower(vars["type"]) == "attendees" {
	// 		part, err := getOne(attendeeCollection, filter, ctx)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			w.Write([]byte(`{"message": "Error"}`))
	// 			return
	// 		}
	// 		foo, err := json.Marshal(part)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			w.Write([]byte(`{"message": "Error"}`))
	// 			return
	// 		}
	// 		fmt.Println(string(foo))
	// 		json.NewEncoder(w).Encode(part)
	// 	}
	// })

	// Delete a person
	func (s *Server) deleteOne(vars map[string]string) (interface{}, error) {
		ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.mongoServer))
	if err != nil {
		log.Fatal(err)
	}
	if client != nil {
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connected to MongoDB!")
		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				panic(err)
			}
		}()
	}
	var collection *mongo.Collection
	if strings.ToLower(vars["type"]) == "drivers" {
		collection = client.Database("Blind").Collection("Drivers")
	} else if strings.ToLower(vars["type"]) == "attendees" {
		collection = client.Database("Blind").Collection("Attendees")
	}

		var search = IDSearch{}
		search.ID, _ = primitive.ObjectIDFromHex(vars["id"])
		filter := bson.D{{Key: "_id", Value: search.ID}}
		_, err = collection.DeleteOne(ctx, filter)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return nil, nil
	}

	// // update a person
	// router.Methods("PUT").Path("/api/{type}/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	fmt.Println(vars["type"])
	// 	b, err := io.ReadAll(r.Body)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		w.Write([]byte(`{"message": "Error"}`))
	// 		return
	// 	}
	// 	fmt.Printf("%s", string(b))
	// 	w.Header().Set("Content-Type", "application/json")
	// 	var search = IDSearch{}
	// 	search.ID, _ = primitive.ObjectIDFromHex(vars["id"])
	// 	part := Participant{}
	// 	_ = json.Unmarshal(b, &part)
	// 	fmt.Printf("part: %+v\n", part)
	// 	update := bson.D{
	// 			{Key: "$set", Value: bson.D{
	// 				{Key: "Name", Value: part.Name},
	// 				{Key: "Email", Value: part.Email},
	// 				{Key: "HomePhone", Value: part.HomePhone},
	// 				{Key: "CellPhone", Value: part.CellPhone},
	// 				{Key: "Address", Value: part.Address},
	// 				{Key: "City", Value: part.City},
	// 				{Key: "State", Value: part.State},
	// 				{Key: "Zip", Value: part.Zip},
	// 				{Key: "Notes", Value: part.Notes},
	// 			}},
	// 		}
	// 	filter := bson.D{{Key: "_id", Value: search.ID}}
	// 	if strings.ToLower(vars["type"]) == "drivers" {

	// 		_, err = driversCollection.UpdateOne(ctx, filter, update)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			w.Write([]byte(`{"message": "Error"}`))
	// 			return
	// 		}
	// 		fmt.Printf("part: Success\n")
	// 		w.Write([]byte(`{"message": "Success"}`))
	// 	} else if strings.ToLower(vars["type"]) == "attendees" {
	// 		_, err = attendeeCollection.UpdateOne(ctx, filter, update)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			w.Write([]byte(`{"message": "Error"}`))
	// 			return
	// 		}
	// 		w.Write([]byte(`{"message": "Success"}`))
	// 	}
	// })

	// spa := spaHandler{staticPath: ".", indexPath: "index.html"}
	// // spa2 := spaHandler{staticPath: ".", indexPath: "update.html"}
	// router.PathPrefix("/").Handler(spa)
	// router.PathPrefix("/update").Handler(spa2)

	// srv := &http.Server{
	// 	Handler: router,
	// 	Addr:    "0.0.0.0:3000",
	// 	// Good practice: enforce timeouts for servers you create!
	// 	WriteTimeout: 15 * time.Second,
	// 	ReadTimeout:  15 * time.Second,
	// }
	// log.Fatal(srv.ListenAndServe())
	// // log.Fatal(srv.ListenAndServeTLS("/home/davidgs/.node-red/combined", "/home/davidgs/.node-red/combined"))

