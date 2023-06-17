package main

import (
	"log"
	"time"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/models"
)

// --------------------------------
//
// --------------------------------
func (app *application) RefreshStoredProces() {
	log.Println("Starting scheduled RefreshStoredProces")
	for _, sp := range app.storedProcs.List() {
		log.Println("Checking sp:", sp.Name)
		serverRcd := sp.DefaultServer
		if serverRcd != nil && serverRcd.ID != "" {
			server, err := app.servers.Get(serverRcd.ID)
			if err == nil {
				log.Println("Refreshing endpoint: ", sp.EndPointName, " ", sp.Name, " ", sp.Lib)
				err := sp.Refresh(*server)
				if err == nil {
					app.storedProcs.Save(sp)
				}
			}
		}
	}
	log.Println("Finished scheduled RefreshStoredProces")

}

// --------------------------------
//
// --------------------------------

func (app *application) RemoveDeletedStoredProcs() {
	for _, sp := range app.storedProcs.List() {
		serverRcd := sp.DefaultServer
		if serverRcd != nil && serverRcd.ID != "" {
			server, err := app.servers.Get(serverRcd.ID)
			if err == nil {
				exits, err := sp.Exists(*server)
				if err == nil && !exits {
					log.Println("Deleting endpoint: ", sp.EndPointName, " ", sp.Name, " ", sp.Lib)
					app.storedProcs.Delete(sp.ID)
				}
			}
		}
	}
}

// --------------------------------
//
// --------------------------------
func (app *application) ProcessPromotions() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()
	log.Println("Starting scheduled Promotion process")
	for _, s := range app.servers.List() {
		app.ProcessPromotion(s)
	}

	log.Println("Finished scheduled Promotion finished")

}

// --------------------------------
//
//	for single server
//
// --------------------------------
func (app *application) ProcessPromotion(s *models.Server) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()

	promotionRecords, err := s.ListPromotion(true)

	//fmt.Println(">>>>>>>>>>>>> promotionRecords>>>>>>>>", promotionRecords)
	if err == nil {
		for _, pr := range promotionRecords {
			app.ProcessPromotionRecord(s, pr)
		}
	}

	s.LastAutoPromoteDate = time.Now().Format(go_ibm_db.TimestampFormat)
	s.Password = s.GetPassword() // make sure it dont update the password
	app.servers.Update(s, false)

}

// --------------------------------
//
//	process single promotion record
//
// --------------------------------
func (app *application) ProcessPromotionRecord(s *models.Server, pr *models.PromotionRecord) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()

	if pr.Status == "P" {

		switch pr.Action {
		case "D": // Delete end point
			app.storedProcs.DeleteByName(pr.Endpoint, pr.Httpmethod)
		case "I", "R": // insert /update endpoint
			newSP := pr.ToStoredProc(*s)
			newSP.ID = newSP.Slug()

			err := newSP.PreapreToSave(*s)

			if err == nil {
				newSP.AddAllowedServer(s)
				app.storedProcs.Save(newSP)
				pr.Status = "C"
				pr.StatusMessage = "Completed"

			} else {
				pr.Status = "E"
				pr.StatusMessage = err.Error()
			}
		default:
			pr.Status = "E"
			pr.StatusMessage = "Unknown Action"
		}
	}

	pr.UpdateStatus(s)
}

// --------------------------------
//
//	for all servers
//
// --------------------------------
func (app *application) PingServers() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()

	for {
		for _, s := range app.servers.List() {
			s.PingQuery = "values(1)"
			log.Println("Pinging server:", s.Name)
			s.GetConnection()

		}
		time.Sleep(10 * time.Second)
	}

}
