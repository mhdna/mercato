package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// branchHub tracks which branches currently have a live WebSocket
// connection, so a mutation to global master data (currencies,
// cashbox_accounts) can push a lightweight "go re-fetch this" notification
// immediately, rather than the branch finding out on its next poll.
//
// Payloads are never sent over the socket — only a delta pointer
// ({"type": "..."}) telling the branch what changed. The branch always
// reacts by calling the same catch-up endpoint it already polls, which
// makes live push and catch-up provably idempotent against each other: a
// redundant or out-of-order notification just triggers a redundant fetch,
// never a correctness problem.
//
// Single in-memory map assumes one kashi instance — fine for the current
// single-Postgres/single-binary deployment, but it means a branch's live
// connection only helps if it happens to be connected to the same
// instance that made the change. Note as a known limitation if kashi is
// ever horizontally scaled.
type branchHub struct {
	mu    sync.RWMutex
	conns map[int64]*wsConn
}

func newBranchHub() *branchHub {
	return &branchHub{conns: make(map[int64]*wsConn)}
}

func (h *branchHub) add(branchID int64, conn *wsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.conns[branchID]; ok {
		// A branch reconnecting (e.g. after a network blip) shouldn't leave
		// the old socket open and leaking — replace it, closing the old one.
		existing.close()
	}
	h.conns[branchID] = conn
}

func (h *branchHub) remove(branchID int64, conn *wsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[branchID] == conn {
		delete(h.conns, branchID)
	}
}

// connectedBranchIDs reports which distinct branches currently have a live
// connection. Since add() replaces rather than accumulates a branch's
// connection (see its comment above), this is really "connected branches",
// not "connected devices" -- a branch with several tills sharing one
// branch API key still appears once here, since branchHub has no way to
// tell those apart.
func (h *branchHub) connectedBranchIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]int64, 0, len(h.conns))
	for branchID := range h.conns {
		ids = append(ids, branchID)
	}
	return ids
}

type branchWSMessage struct {
	Type      string `json:"type"`
	CommandID int64  `json:"command_id,omitempty"`
	// ClientRef carries the originating expense's client_ref for
	// "expense_images_uploaded" -- unlike every other pushed message
	// today, kashi-pos needs to know *which* locally-tracked entity this
	// is about (to flip the matching QR dialog to success), not just "go
	// re-fetch this kind of thing".
	ClientRef string `json:"client_ref,omitempty"`
}

// notify sends a message to exactly one branch, if it currently has a live
// connection. Unlike broadcastAll, this is for branch-scoped events (a
// command enqueued for one specific branch) — silently doing nothing when
// the branch isn't connected is fine, since notify is always paired with a
// poll-based catch-up path (see kashi-pos's runSyncTick fetching pending
// commands every tick) that doesn't depend on this push ever arriving.
func (h *branchHub) notify(branchID int64, message branchWSMessage) {
	h.mu.RLock()
	conn, ok := h.conns[branchID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	if err := conn.writeJSON(message); err != nil {
		log.Printf("branch hub: notify branch %d failed: %v", branchID, err)
	}
}

// broadcastAll notifies every currently-connected branch — appropriate for
// today's only pushed entities (currencies, cashbox_accounts), which apply
// identically to every branch. A branch-scoped notification (once
// branch-scoped pushed entities exist) would instead look up and send to
// one specific branch id.
func (h *branchHub) broadcastAll(message branchWSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for branchID, conn := range h.conns {
		if err := conn.writeJSON(message); err != nil {
			log.Printf("branch hub: notify branch %d failed: %v", branchID, err)
		}
	}
}

var wsUpgrader = websocket.Upgrader{
	// kashi-pos is a Go client dialing directly, not a browser page, so
	// there's no cross-origin risk here the way there would be for a
	// browser-facing endpoint — the real authentication is the branch
	// code/key headers already checked by branchAuthMiddleware before this
	// handler runs at all.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// branchWS is push-only from kashi's side — kashi-pos never sends anything
// over this connection (all writes go through the outbox/REST path in
// branch_sync.go). The read loop exists purely to detect the peer closing
// the connection; nothing it reads is acted on.
func (server *Server) branchWS(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("branch %d: websocket upgrade failed: %v", branchID, err)
		return
	}
	defer conn.Close()

	wrapped := newWSConn(conn)
	server.branchHub.add(branchID, wrapped)
	server.adminHub.broadcastAll(adminWSMessage{Type: "branch_connection_changed", BranchIDs: server.branchHub.connectedBranchIDs()})
	defer func() {
		server.branchHub.remove(branchID, wrapped)
		server.adminHub.broadcastAll(adminWSMessage{Type: "branch_connection_changed", BranchIDs: server.branchHub.connectedBranchIDs()})
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
