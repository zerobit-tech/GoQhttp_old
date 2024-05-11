package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zerobit-tech/GoQhttp/env"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) onLoad() {
	go app.createRPGDrivers()

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) createRPGDrivers() {

	type requiredEP struct {
		lib    string
		spName string
	}

	for _, server := range app.servers.List() {

		rpgDriverLib := env.RpgDriverLib(server.Name)
		nameSpace := env.RpgDriverNameSpace(server.Name)

		requiredEndpoints := []requiredEP{
			{lib: rpgDriverLib, spName: "iPLUG4K"},
			{lib: rpgDriverLib, spName: "iPLUG32K"},
			{lib: rpgDriverLib, spName: "iPLUG65K"},
			{lib: rpgDriverLib, spName: "iPLUG512K"},
			{lib: rpgDriverLib, spName: "iPLUG1M"},
			{lib: rpgDriverLib, spName: "iPLUG5M"},
			{lib: rpgDriverLib, spName: "iPLUG10M"},
			{lib: rpgDriverLib, spName: "iPLUG15M"},
		}

		for _, driver := range requiredEndpoints {

			id := fmt.Sprintf("%s_%s_post", nameSpace, driver.spName)
			_, err := app.storedProcs.Get(id)
			if err != nil {
				// create new endpoint
				sp := &storedProc.StoredProc{
					EndPointName:     driver.spName,
					HttpMethod:       "POST",
					Name:             driver.spName,
					Lib:              driver.lib,
					AllowWithoutAuth: false,
					Namespace:        nameSpace,
					IsSpecial:        true,
				}

				srcd := &storedProc.ServerRecord{ID: server.ID, Name: server.Name}
				sp.DefaultServer = srcd
				sp.AddAllowedServer(server.ID, server.Name)

				err = server.PrepareToSave(context.Background(), sp)
				if err != nil {
					log.Println("Error createing Program drivers ", err)
				}
				_, err := app.storedProcs.Save(sp)
				if err != nil {
					log.Println("Error creating Program drivers ", err)
				}
			}

		}
	}
	app.invalidateEndPointCache()

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) deleteRPGDrivers() {
	for _, e := range app.storedProcs.List(true) {
		if e.IsSpecial {
			app.storedProcs.Delete(e.ID)
		}
	}
}
