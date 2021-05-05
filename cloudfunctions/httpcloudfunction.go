package cloudfunctions

import (
	"fmt"
	"log"

	"encoding/json"

	"net/http"

	"github.com/prometheus/alertmanager/template"
)

func PrometheusWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	log.Println(data)

	fmt.Fprint(w, "successfully processed")
}
