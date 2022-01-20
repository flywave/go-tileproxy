package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/mock/{x:[0-9]+}/{y:[0-9]+}.add", MockHandler)

	http.ListenAndServe(":8001", r)
}

func MockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])

	w.Write([]byte(fmt.Sprintf("%d", x+y)))
	w.WriteHeader(http.StatusOK)
}
