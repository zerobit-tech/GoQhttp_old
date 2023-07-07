package main

import (
	"context"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/models"
	"github.com/onlysumitg/GoQhttp/lic"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
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
				err := sp.Refresh(context.Background(), server)
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
				exits, err := sp.Exists(context.Background(), server)
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

	_, err := lic.VerifyLicFiles()
	if err != nil {
		log.Println("Process Promotions Skipped.....: ", err.Error())
	}

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

	defer concurrent.Recoverer("ProcessPromotion")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

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
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	if pr.Status == "P" {

		switch pr.Action {
		case "D": // Delete end point
			app.storedProcs.DeleteByName(pr.Endpoint, pr.Httpmethod)
		case "I", "R": // insert /update endpoint
			newSP := pr.ToStoredProc(s)
			newSP.ID = newSP.Slug()

			err := newSP.PreapreToSave(context.Background(), *s)

			if err == nil {
				newSP.AddAllowedServer(s)

				// handle param alias
				for _, p := range newSP.Parameters {
					for _, pALias := range pr.ParamAliasRcds {
						if strings.EqualFold(p.Name, pALias.Name) {
							p.Alias = strings.TrimSpace(strings.ToUpper(pALias.Alias))

						}
					}
				}

				app.storedProcs.Save(newSP)
				app.invalidateEndPointCache()

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
			log.Println("ping connection 1", r)
		}
	}()
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	for _, s := range app.servers.List() {
		s.PingQuery = "select * from qsys2.systables"
		log.Println("Pinging server:", s.Name)
		s.GetConnection()
		// if err == nil {
		// 	db.Close()
		// }

	}
	//time.Sleep(30 * time.Second)

}

// --------------------------------
//
//	for single server
//
// --------------------------------
func (app *application) SyncUserToken(s *models.Server) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()

	tokenRecords, err := s.SyncUserTokenRecords(true)
	//fmt.Println(">>>>>>>>>>>>> promotionRecords>>>>>>>>", promotionRecords)
	if err == nil {
		for _, tk := range tokenRecords {
			app.ProcessSyncUserToken(s, tk)
		}
	}
	return err
}

// --------------------------------
//
//	for single server
//
// --------------------------------
func (app *application) ProcessSyncUserToken(s *models.Server, tk *models.UserTokenSyncRecord) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()

	//app.ProcessPromotionRecord(s, tk)
	user, err := app.users.GetByUserName(tk.Username)

	if tk.Status == "P" {
		if err == nil {
			if user.ServerId != s.ID {
				tk.Status = "E"
				tk.StatusMessage = "User has a different default server"

			} else {
				user.Token = tk.Token

				app.users.Save(user, false)
				tk.Status = "C"
				tk.StatusMessage = "Completed"
			}
		} else {
			tk.Status = "E"
			tk.StatusMessage = err.Error()
		}
	}
	tk.UpdateStatusUserTokenTable(s)
}
