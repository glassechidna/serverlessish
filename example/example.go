package main

import (
	"net/http"
	"net/http/httputil"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dump, _ := httputil.DumpRequest(r, true)
		w.Header().Add("Some-Header-Set-Once", "once")
		w.Header().Add("Some-Header-Set-Twice", "first")
		w.Header().Add("Some-Header-Set-Twice", "second")
		w.Write(append([]byte("this is what i received on port 8080:\n\n"), dump...))
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
	    panic(err)
	}
}
