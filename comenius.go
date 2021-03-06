package main

import (
    "strconv"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type LearnerDetails struct {
	CertificateList            []Certificate
	MoneyRaisedWeek            int64
	TotalContributionsReceived int64
	ContributionHistory        []int64
}

type ContributorDetails struct {
	CertificateList  []Certificate
	ContributionList []Contribution
	TotalMoneyRaised int64
	TotalImpact      int64
	PeopleImpacted   map[string]int64
}

type Certificate struct {
	CertificateURL string    `json:"certificateURL"`
	CourseImageURL string    `json:"courseImageURL"`
	Name           string    `json:"name"`
	Platform       string    `json:"platform"`
	Price          int64     `json:"price"`
	URL            string    `json:"url"`
	Date           time.Time `json:"date"`
	FullyFunded    bool      `json:"fullyFunded"`
	RaisedAmount   int64     `json:"raisedAmount"`
}

type Learner struct {
	FullName string
	Login    string
}

type Contributor struct {
	FullName string
	Login    string
}

type LoginRequest struct {
	User string `json:"username"`
	Pass string `json:"password"`
	Type string `json:"type"`
}

type CertificateRequest struct {
	User string
}

type DonateRequest struct {
	Amount    string `json:"amount"`
	User      string `json:"user"`
	Recipient string `json:"recipient"`
	CertID    string `json:"certID"`
}

type Contribution struct {
	Amount            int64
	CertificateID     string
	Date              time.Time
	Recipient         string
	TransactionNumber string
}

var opt = option.WithCredentialsFile("./serviceAccountKey.json")
var conf = &firebase.Config{}
var app, _ = firebase.NewApp(context.Background(), nil, opt)
var client, _ = app.Firestore(context.Background())

func learner(w http.ResponseWriter, r *http.Request) {
	user := Learner{FullName: r.URL.Path[len("/learner/"):], Login: r.URL.Path[1:]}
	t, err := template.ParseFiles("static/learner.html")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	t.Execute(w, user)
}

func contributor(w http.ResponseWriter, r *http.Request) {
	user := Contributor{FullName: r.URL.Path[len("/contributor/"):], Login: r.URL.Path[len("/contributor/"):]}
	t, err := template.ParseFiles("static/contributor.html")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	t.Execute(w, user)
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	buffer, err := io.ReadAll(body)

	if err != nil {
		fmt.Println("Request read error:")
		fmt.Println(err)
	}

	var request LoginRequest
	json.Unmarshal(buffer, &request)

	username := request.User

	iter := client.Collection(request.Type).Documents(context.Background())

	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			fmt.Fprintf(w, "Error %v", err)
		}

		if username == doc.Data()["username"].(string) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"authenticate": true}`))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"authenticate": false}`))
}

func loginGet(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/login.html")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	t.Execute(w, nil)
}

func certificate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"submitted": true}`))
}

func donate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request DonateRequest
	decoder.Decode(&request)

	fmt.Println(request)

	money, _ := strconv.ParseInt(request.Amount, 10, 64)

	client.Collection("contribution").Add(context.Background(),
		&Contribution{
			Amount:            money * 100,
			CertificateID:     "certificate/yND8KwflMUbc78tR7Ri5",
			Date:              time.Now(),
			Recipient:         "JohnDoe2713",
			TransactionNumber: "",
		},
	)

    docsnap, _ := client.Collection("contributor").Doc("OeGjk5ea18jllboHwCw8").Get(context.Background())
    data := docsnap.Data()
    list := data["contributionList"].([]interface {})
    a := "contribution/" + ref.ID
    list = append(list, a)
    data["contributionList"] = list
    client.Collection("contributor").Doc("OeGjk5ea18jllboHwCw8").Set(context.Background(), data)
    fmt.Println(list)

    fmt.Println(ref.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"submitted": true}`))
}

func getLearnerDetails(w http.ResponseWriter, r *http.Request) {
	var Certs []Certificate

	username := r.URL.Query().Get("username")

	// Finding user certificates/active listings
	iter := client.Collection("learner").Documents(context.Background())

	var certList []interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "Error %v", err)
		}
		if username == doc.Data()["username"] {
			certList = doc.Data()["certificateList"].([]interface{})
			break
		}
	}

	for _, s := range certList {
		learner_doc := client.Doc(s.(string))
		docsnap, _ := learner_doc.Get(context.Background())
		dataMap := docsnap.Data()
		course_doc := client.Doc(dataMap["courseID"].(string))
		coursedocsnap, _ := course_doc.Get(context.Background())
		courseDataMap := coursedocsnap.Data()

		Cert := Certificate{
			CertificateURL: dataMap["certificateURL"].(string),
			CourseImageURL: courseDataMap["courseImageURL"].(string),
			Name:           courseDataMap["name"].(string),
			Platform:       courseDataMap["platform"].(string),
			Price:          courseDataMap["price"].(int64),
			URL:            courseDataMap["url"].(string),
			Date:           dataMap["date"].(time.Time),
			FullyFunded:    dataMap["fullyFunded"].(bool),
			RaisedAmount:   dataMap["raisedAmount"].(int64),
		}
		Certs = append(Certs, Cert)
	}

	// Calculating total contributions
	contributionIter := client.Collection("contribution").Where("recipient", "==", username).Documents(context.Background())
	var totalContributions int64 = 0
	var weeklyContributions int64 = 0
	contributionHistory := []int64{0, 0, 0, 0, 0, 0, 0}
	for {
		doc, err := contributionIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "Error %v", err)
		}
		totalContributions += doc.Data()["amount"].(int64)
		duration := time.Since(doc.Data()["date"].(time.Time))
		if duration.Hours() <= 168 {
			weeklyContributions += doc.Data()["amount"].(int64)
			contributionHistory[int64(duration.Hours()/24)] += doc.Data()["amount"].(int64)
		}
	}

	learnerDetails := LearnerDetails{
		CertificateList:            Certs,
		MoneyRaisedWeek:            int64(weeklyContributions),
		TotalContributionsReceived: int64(totalContributions),
		ContributionHistory:        contributionHistory,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(learnerDetails)
}

func getContributorDetails(w http.ResponseWriter, r *http.Request) {
	var Contribs []Contribution

	username := r.URL.Query().Get("username")

	// Finding user certificates/active listings
	iter := client.Collection("contributor").Documents(context.Background())

	var contribList []interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "Error %v", err)
		}
		if username == doc.Data()["username"].(string) {
			contribList = doc.Data()["contributionList"].([]interface{})
			break
		}
	}

	var totalMoneyRaised int64 = 0
	peopleImpacted := make(map[string]int64)
	for _, s := range contribList {
		contributor_doc := client.Doc(s.(string))
		docsnap, _ := contributor_doc.Get(context.Background())
		dataMap := docsnap.Data()

		Contrib := Contribution{
			Amount:            dataMap["Amount"].(int64),
			CertificateID:     dataMap["CertificateID"].(string),
			Date:              dataMap["Date"].(time.Time),
			Recipient:         dataMap["Recipient"].(string),
			TransactionNumber: dataMap["TransactionNumber"].(string),
		}
		if _, ok := peopleImpacted[dataMap["Recipient"].(string)]; ok {
			peopleImpacted[dataMap["Recipient"].(string)] = 0
		}
		peopleImpacted[dataMap["Recipient"].(string)] += dataMap["Amount"].(int64)
		Contribs = append(Contribs, Contrib)
		totalMoneyRaised += dataMap["Amount"].(int64)
	}
	var Certs []Certificate

	// Finding user certificates/active listings
	learnerIter := client.Collection("learner").Documents(context.Background())

	var certList []interface{}
	for {
		doc, err := learnerIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "Error %v", err)
		}
		certList = append(certList, doc.Data()["certificateList"].([]interface{})...)
	}

	for _, s := range certList {
		learner_doc := client.Doc(s.(string))
		docsnap, _ := learner_doc.Get(context.Background())
		dataMap := docsnap.Data()
		course_doc := client.Doc(dataMap["courseID"].(string))
		coursedocsnap, _ := course_doc.Get(context.Background())
		courseDataMap := coursedocsnap.Data()

		Cert := Certificate{
			CertificateURL: dataMap["certificateURL"].(string),
			CourseImageURL: courseDataMap["courseImageURL"].(string),
			Name:           courseDataMap["name"].(string),
			Platform:       courseDataMap["platform"].(string),
			Price:          courseDataMap["price"].(int64),
			URL:            courseDataMap["url"].(string),
			Date:           dataMap["date"].(time.Time),
			FullyFunded:    dataMap["fullyFunded"].(bool),
			RaisedAmount:   dataMap["raisedAmount"].(int64),
		}
		Certs = append(Certs, Cert)
	}

	contributorDetails := ContributorDetails{
		CertificateList:  Certs,
		ContributionList: Contribs,
		TotalMoneyRaised: totalMoneyRaised,
		TotalImpact:      int64(len(peopleImpacted)),
		PeopleImpacted:   peopleImpacted,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contributorDetails)
}

func main() {
	port := os.Getenv("PORT")

	port = "8080" // uncomment for local testing

	r := mux.NewRouter()
	r.HandleFunc("/learner_details", getLearnerDetails).Methods(http.MethodGet)
	r.HandleFunc("/contributor_details", getContributorDetails).Methods(http.MethodGet)
	r.HandleFunc("/login", loginPost).Methods(http.MethodPost)
	r.HandleFunc("/login", loginGet).Methods(http.MethodGet)
	r.HandleFunc("/certificate", certificate).Methods(http.MethodPost)
	r.HandleFunc("/donate", donate).Methods(http.MethodPost)
	r.PathPrefix("/learner").HandlerFunc(learner).Methods(http.MethodGet)
	r.PathPrefix("/contributor").HandlerFunc(contributor).Methods(http.MethodGet)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	log.Print("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
