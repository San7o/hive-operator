package main

import (
	"flag"
	"net/http"
	"io"
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	DefaultPort = "8090"
)

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	log := log.FromContext(r.Context())
	
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err, "error reading request's body")
	}
	log.Info("received request", "body", string(bytes))
}

func main() {

	var portFlag = flag.String("port", DefaultPort, "port to listen to")

	ctrl.SetLogger(zap.New())
	log := log.FromContext(context.Background())
	http.HandleFunc("/ingest", ingestHandler)

	log.Info("Server listening", "port", *portFlag)
	
	if err := http.ListenAndServe(":"+*portFlag, nil); err != nil {
		log.Error(err, "server error")
	}
	
	return
}
