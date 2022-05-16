package main

import (
	"fmt"
	"github.com/antonlindstrom/pgstore"
	"html/template"
	"log"
	"net/http"
	"time"
)

var sessionID string = ""
var exchange string = ""
var vhost string = ""

type BackendData struct {
	Host      string
	User      string
	Pass      string
	SessionID string
	Exchange  string
	Vhost     string
}

/*
Make sure you login into the PG SQL instance and create the DB

psql -P expanded=auto -h <HOST_IP> -U <user> <db>

create database sessions_db;

metadata-> \l
                                                List of databases
     Name      |       Owner       | Encoding |  Collate   |   Ctype    |            Access privileges
---------------+-------------------+----------+------------+------------+-----------------------------------------
 sessions_db   | pcc               | UTF8     | en_US.UTF8 | en_US.UTF8 |

metadata=> \c sessions_db

SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
You are now connected to database "sessions_db" as user "pcc".
sessions_db=> \d
                List of relations
 Schema |         Name         |   Type   | Owner
--------+----------------------+----------+-------
 public | http_sessions        | table    | pcc
 public | http_sessions_id_seq | sequence | pcc
(2 rows)

sessions_db=> select * from http_sessions;

-[ RECORD 1 ]---------------------------------------------------------------------------
id          | 1
key         | \x4153563258563551324b43495043525237433258355336335737345a4b4458465... ...
data        | \x4d5459314d6a49774f446b784d48784564693143516b464651 ... ...
created_on  | 2022-05-10 18:55:10.635414+00
modified_on | 2022-05-10 18:55:10.635414+00
expires_on  | 2022-06-09 18:55:10.635414+00
*/

func Home(w http.ResponseWriter, r *http.Request) {
	// Fetch new store.
	message := ""
	store, err := pgstore.NewPGStore(
		"postgres://"+PGUSER+":"+PGPASS+"@"+PGHOST+":5432/sessions_db?sslmode=disable",
		[]byte("................................"))
	if err != nil {
		message = fmt.Sprintf("error : could not create a session store using pgstore.NewPGStore : %v", err.Error())
		log.Printf(message)
		return
	}
	defer store.Close()

	// Run a background goroutine to clean up expired sessions from the database.
	defer store.StopCleanup(store.Cleanup(time.Minute * 5))

	// Get a session.
	session, err := store.Get(r, "http-session")
	if err != nil {
		message = fmt.Sprintf("error : could not get a session object from store using store.Get : %v", err.Error())
		log.Printf(message)
	}

	//session, _ := cookie.Get(r, "http-session")

	var SessionID interface{} = session.Values["SessionID"]

	if SessionID == nil {
		log.Printf("@> Home : SessionID is <nil> !")
		sessionID = "sess_" + getRandomString() + "_" + getRandomString()
		log.Printf("@> Home : sessionID : %v", sessionID)
		session.Values["SessionID"] = sessionID
		_ = session.Save(r, w)
	} else {
		log.Printf("@> found (SessionID) : (%v)", sessionID)
		sessionID = session.Values["SessionID"].(string)
	}

	// Delete session.
	//session.Options.MaxAge = -1
	//if err = session.Save(r, w); err != nil {
	//	log.Fatalf("Error saving session: %v", err)
	//}

	err = renderPage(w, r, "stomp_websocket.html")
	if err != nil {
		log.Printf("error : could not render page using renderPage : %v", err.Error())
	}
}

func getBackendConfiguration() BackendData {

	//randomQueueName := generateRandomQueueName()

	details := BackendData{
		Host:      MQHOST,
		User:      MQUSER,
		Pass:      MQPASS,
		SessionID: sessionID,
		Exchange:  exchange,
		Vhost:     vhost,
	}

	log.Printf("getBackendConfiguration() : SessionID : %v", sessionID)
	log.Printf("getBackendConfiguration() : exchange : %v", exchange)
	log.Printf("getBackendConfiguration() : vhost : %v", vhost)

	return details
}

func renderPage(w http.ResponseWriter, r *http.Request, webpage string) error {
	tmpl := template.Must(template.ParseFiles(webpage))
	backendConfig := getBackendConfiguration()
	err := tmpl.Execute(w, backendConfig)
	if err != nil {
		log.Printf("renderPage() : tmpl.Execute(w, nil) : error : %v", err.Error())
		return err
	}
	return nil
}
