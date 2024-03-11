package linkchecker

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CheckLinkRequest struct {
	UUID    string
	PageURL string
}

type SchedulerResult struct {
	UUID   string
	URL    string
	Status string
	Result *LinkCheckResult
}

var (
	uidToResultMap = make(map[string]*SchedulerResult)
	mapLock        = &sync.RWMutex{}
)

func GetResultForUUID(uid string) *SchedulerResult {
	mapLock.RLock()
	defer mapLock.RUnlock()
	return uidToResultMap[uid]
}

func SubmitLinkForCheck(link string) string {
	// TODO better for this to be the sha256 of the link
	uid := uuid.New().String()

	uidToResultMap[uid] = &SchedulerResult{
		UUID:   uid,
		URL:    link,
		Status: "PENDING",
	}

	log.Printf("Submitted link %s for check with UUID %s", link, uid)

	go func() {
		result, err := CheckLink(link)
		log.Printf("Link check for %s completed with %d errors", link, len(result.FoundErrors))

		mapLock.Lock()
		if err != nil {
			uidToResultMap[uid].Status = err.Error()
			return
		}

		uidToResultMap[uid].Result = &result
		uidToResultMap[uid].Status = "COMPLETED"
		mapLock.Unlock()

		// keep the results around for a while and then remove from the map
		go func() {
			<-time.After(2 * 24 * time.Hour)
			mapLock.Lock()
			delete(uidToResultMap, uid)
			mapLock.Unlock()
		}()
	}()

	return uid
}
