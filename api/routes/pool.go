package routes

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

func poolHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	poolID := params["poolID"]

	var res PoolResult
	err := db.QueryRow("SELECT id, name FROM pools WHERE id IN (SELECT pool_id FROM pool_observers WHERE disabled=false) AND id=$1", poolID).Scan(&res.ID, &res.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), 404)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}
	writeJson(w, res)
}
