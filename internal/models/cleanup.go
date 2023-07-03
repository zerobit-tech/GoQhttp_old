package models

import (
	bolt "go.etcd.io/bbolt"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func DailyDataCleanup_TESTMODE(db *bolt.DB) {
	//go DeleteALLEndpoint(db) //goroutine
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func DailyDataCleanup(db *bolt.DB) {
	//TODO

}
