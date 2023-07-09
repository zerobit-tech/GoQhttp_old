package dbserver

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) Load() error {
	t := s.Type

	if t == "" {
		t = "IBM I"
	}
	driversMu.RLock()
	dbX, ok := drivers[t]
	driversMu.RUnlock()
	if !ok {
		return fmt.Errorf("sql: unknown driver %q (forgotten import?)", dbX)
	}
	s.dbDriver = dbX
	dbX.Load(s)
	return nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetDbDriver() DbDriver {

	if s.dbDriver != nil {
		return s.dbDriver
	}
	t := s.Type

	if t == "" {
		t = "IBM I"
	}
	driversMu.RLock()
	dbX, ok := drivers[t]
	driversMu.RUnlock()
	if !ok {
		log.Fatalf("sql: unknown driver %q (forgotten import?)", dbX)
	}
	s.dbDriver = dbX
	dbX.Load(s)
	return s.dbDriver
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) APICall(ctx context.Context, callID string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
	return s.GetDbDriver().APICall(ctx, callID, sp, params)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) Refresh(ctx context.Context, sp *storedProc.StoredProc) error {
	return s.GetDbDriver().Refresh(ctx, sp)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) PreapreToSave(ctx context.Context, sp *storedProc.StoredProc) error {
	return s.GetDbDriver().PreapreToSave(ctx, sp)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) DummyCall(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, error) {
	return s.GetDbDriver().DummyCall(sp, givenParams)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) ListPromotion(withupdate bool) ([]*storedProc.PromotionRecord, error) {
	return s.GetDbDriver().ListPromotion(withupdate)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) UpdateStatusForPromotionRecord(p storedProc.PromotionRecord) {
	s.GetDbDriver().UpdateStatusForPromotionRecord(p)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) PromotionRecordToStoredProc(p storedProc.PromotionRecord) *storedProc.StoredProc {
	return s.GetDbDriver().PromotionRecordToStoredProc(p)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) Exists(ctx context.Context, sp *storedProc.StoredProc) (bool, error) {
	return s.GetDbDriver().Exists(ctx, sp)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) UpdateStatusUserTokenTable(p storedProc.UserTokenSyncRecord) {
	s.GetDbDriver().UpdateStatusUserTokenTable(p)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) SyncUserTokenRecords(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error) {
	return s.GetDbDriver().SyncUserTokenRecords(withupdate)

}
