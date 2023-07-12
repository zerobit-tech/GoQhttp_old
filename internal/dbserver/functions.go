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

	if s.Type == "" {
		s.Type = "IBM I"
	}
	driversMu.RLock()
	dbX, ok := drivers[s.Type]
	driversMu.RUnlock()
	if !ok {
		return fmt.Errorf("sql: unknown driver %q (forgotten import?)", dbX)
	}
	s.dbDriver = dbX
	dbX.LoadX(s)
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
	dbX.LoadX(s)
	return s.dbDriver
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) APICall(ctx context.Context, callID string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
	return s.GetDbDriver().APICallX(ctx, callID, sp, params)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) Refresh(ctx context.Context, sp *storedProc.StoredProc) error {
	return s.GetDbDriver().RefreshX(ctx, sp)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) PrepareToSave(ctx context.Context, sp *storedProc.StoredProc) error {
	return s.GetDbDriver().PrepareToSaveX(ctx, sp)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) DummyCall(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, error) {
	return s.GetDbDriver().DummyCallX(sp, givenParams)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) ListPromotion(withupdate bool) ([]*storedProc.PromotionRecord, error) {
	return s.GetDbDriver().ListPromotionX(withupdate)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) UpdateStatusForPromotionRecord(p storedProc.PromotionRecord) {
	s.GetDbDriver().UpdateStatusForPromotionRecordX(p)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// func (s *Server) PromotionRecordToStoredProc(p storedProc.PromotionRecord) *storedProc.StoredProc {
// 	return s.GetDbDriver().PromotionRecordToStoredProcX(p)

// }

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) PromotionRecordToStoredProc(p storedProc.PromotionRecord) *storedProc.StoredProc {
	sp := &storedProc.StoredProc{
		EndPointName: p.Endpoint,
		HttpMethod:   p.Httpmethod,
		Name:         p.Storedproc,
		Lib:          p.Storedproclib,
	}
	if p.UseSpecificName == "Y" {
		sp.UseSpecificName = true
	}

	if p.UseWithoutAuth == "Y" {
		sp.AllowWithoutAuth = true
	}
	srcd := &storedProc.ServerRecord{
		ID:   s.ID,
		Name: s.Name,
	}
	sp.DefaultServer = srcd

	return sp
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) Exists(ctx context.Context, sp *storedProc.StoredProc) (bool, error) {
	return s.GetDbDriver().ExistsX(ctx, sp)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) UpdateStatusUserTokenTable(p storedProc.UserTokenSyncRecord) {
	s.GetDbDriver().UpdateStatusUserTokenTableX(p)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) SyncUserTokenRecords(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error) {
	return s.GetDbDriver().SyncUserTokenRecordsX(withupdate)

}
