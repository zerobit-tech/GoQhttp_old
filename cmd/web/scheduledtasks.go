package main

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/internal/ibmiServer"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/lic"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
)

// --------------------------------
//
// --------------------------------
func (app *application) RefreshStoredProces() {
	log.Println("Starting scheduled RefreshStoredProces")
	for _, sp := range app.storedProcs.List(false) {
		log.Println("Checking sp:", sp.Name)
		serverRcd := sp.DefaultServer
		if serverRcd != nil && serverRcd.ID != "" {
			server, err := app.servers.Get(serverRcd.ID)
			if err == nil {
				log.Println("Refreshing endpoint: ", sp.EndPointName, " ", sp.Name, " ", sp.Lib)
				err := server.Refresh(context.Background(), sp)
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
	for _, sp := range app.storedProcs.List(false) {
		serverRcd := sp.DefaultServer
		if serverRcd != nil && serverRcd.ID != "" {
			server, err := app.servers.Get(serverRcd.ID)
			if err == nil {
				exits, err := server.Exists(context.Background(), sp)
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
func (app *application) ProcessPromotion(s *ibmiServer.Server) {

	defer concurrent.Recoverer("ProcessPromotion")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	promotionRecords, err := s.ListPromotion(true)

	//fmt.Println(">>>>>>>>>>>>> promotionRecords>>>>>>>>", promotionRecords)
	if err == nil {
		for _, pr := range promotionRecords {
			app.ProcessPromotionRecord(s, pr)
		}
	}

	s.LastAutoPromoteDate = time.Now().Format(godbc.TimestampFormat)
	//s.Password = s.GetPassword() // make sure it dont update the password
	app.servers.Update(s, false)

}

// --------------------------------
//
//	process single promotion record
//
// --------------------------------
func (app *application) ProcessPromotionRecord(s *ibmiServer.Server, pr *storedProc.PromotionRecord) {
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
			newSP := s.PromotionRecordToStoredProc(*pr)
			newSP.ID = newSP.Slug() // id is by name_httpmethod --> auto replace old if alreay exits

			var err error = nil
			if app.RpgEndpointModel.DuplicateByName(newSP.EndPointName, newSP.HttpMethod, newSP.Namespace) {

				err = errors.New("Duplicate name. Conflict with program endpoint")

			}
			if err == nil {
				err = s.PrepareToSave(context.Background(), newSP)
			}
			if err == nil {
				newSP.AddAllowedServer(s.ID, s.Name)

				// handle param alias
				for _, p := range newSP.Parameters {
					for _, pALias := range pr.ParamAliasRcds {
						if strings.EqualFold(p.Name, pALias.Name) {
							p.Alias = strings.TrimSpace(strings.ToUpper(pALias.Alias))

						}
					}

					for _, pPlacement := range pr.ParamPlacementRcds {
						if strings.EqualFold(p.Name, pPlacement.Name) {
							p.Placement = strings.TrimSpace(strings.ToUpper(pPlacement.Placement))

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

	s.UpdateStatusForPromotionRecord(*pr)
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
		if !app.ShouldPingServer(s) {
			log.Println("skilling server ping based on last use:", s.Name)
			continue
		}

		s.PingQuery = "values('1')"
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
func (app *application) SyncUserToken(s *ibmiServer.Server) error {
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
func (app *application) ProcessSyncUserToken(s *ibmiServer.Server, tk *storedProc.UserTokenSyncRecord) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()

	//app.ProcessPromotionRecord(s, tk)
	user, err := app.users.GetByEmail(tk.Username)

	if err == nil {

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
	} else {
		tk.Status = "E"
		tk.StatusMessage = err.Error()
	}
	s.UpdateStatusUserTokenTable(*tk)
}
