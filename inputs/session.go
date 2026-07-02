package inputs

import "sync"

type flowKind int

const (
	flowNone flowKind = iota
	flowAddCategory
	flowRenameCategory
	flowAddHabit
	flowRenameHabit
)

type session struct {
	flow       flowKind
	categoryID int64
	habitID    int64
}

var (
	sessionsMu sync.Mutex
	sessions   = map[int64]session{}
)

func setSession(chatID int64, s session) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	sessions[chatID] = s
}

func getSession(chatID int64) (session, bool) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	s, ok := sessions[chatID]
	return s, ok
}

func clearSession(chatID int64) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	delete(sessions, chatID)
}
