package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"encoding/json"
	"net/http"
	"embed"
	"os"
	"sort"
	"strings"
	"text/template"
)
var (
	//go:embed templates
	res embed.FS
	pages = map[string]string{
		"/": "templates/index.html",
	}
)

//service to oas version
var cachedAvailableServices = make([]map[string]string, 0)
var versions = make([]string, 0)
var apidocsExtension = ""
var apiSchemaUrlEnv = ""
var apiSchemaUrlEnvExists = false

func main() {
	refreshCron, exists := os.LookupEnv("REFRESH_CRON")
	if !exists {
		refreshCron = "@every 1m"
	}

	servicesEnv := os.Getenv("SERVICES")
	if servicesEnv == "" {
		log.Println("Environment variable \"SERVICES\" is empty")
		os.Exit(2)
	}
	services := make([]string, 0)
	services = mapValues(strings.Split(servicesEnv[1:len(servicesEnv)-1], ","), func(s string) string {
		return s[1 : len(s)-1]
	})
	sort.Strings(services)

	//set versions
	versionsEnv, versionsEnvExists := os.LookupEnv("VERSIONS")
	apidocsExtensionEnv, apidocsExtensionEnvExists := os.LookupEnv("APIDOCS_EXTENSION")
	apiSchemaUrlEnv, apiSchemaUrlEnvExists = os.LookupEnv("API_SCHEMA_URL")

	if versionsEnvExists {
		versions = mapValues(strings.Split(versionsEnv[1:len(versionsEnv)-1], ","), func(s string) string {
			return s[1 : len(s)-1]
		})
	} else {
		versions = []string{"v2", "v3"}
	}

	if (apidocsExtensionEnvExists && len(apidocsExtensionEnv) != 0) {
		log.Println("Trying swagger extension: " + apidocsExtensionEnv)
		apidocsExtension = "." + apidocsExtensionEnv
	}

	log.Println("Server started on 3000 port!")
	log.Println("Services:", services)
	if (apiSchemaUrlEnvExists) {
		log.Println("Schema URL:", apiSchemaUrlEnv)
	}
	log.Println("Discovering versions:", versions, " with extension", apidocsExtensionEnv)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page, ok := pages[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		tpl, err := template.ParseFS(res, page)
		if err != nil {
			log.Printf("page %s not found in pages cache...", r.RequestURI)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		content, err := json.Marshal(cachedAvailableServices)
		data := string(content)
		if err := tpl.Execute(w, data); err != nil {
			return
		}
	})


	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		refreshCache(services)
		w.WriteHeader(200)
	})

	refreshCache(services)

	c := cron.New()
	c.AddFunc(refreshCron, func() {
		log.Println("Cron init")
		refreshCache(services)
		log.Println("Cron has been finished")
	})
	c.Start()

	_ = http.ListenAndServe(":3000", nil)
}

func checkService(service string) {
	log.Println(apiSchemaUrlEnv)
	if (apiSchemaUrlEnv != "") {
		url := "http://" + service + "/" + apiSchemaUrlEnv
		log.Println("Trying url: " + url)
		resp, err := http.Get(url)

		if resp != nil {
			log.Println("for schema " + apiSchemaUrlEnv + " status code is " + resp.Status)
			resp.Body.Close()
		}

		log.Println("for " + service + " schemaUrl is '" + apiSchemaUrlEnv + "'")
		if err == nil && strings.Contains(resp.Status, "200") {
			cachedAvailableServices = append(cachedAvailableServices, map[string]string{
				"name": service,
				"url":  "http://" + service + "/" + apiSchemaUrlEnv,
			})
		}
	} else {
		passedVersion := ""
		for _, ver := range versions {

			url := "http://" + service + "/" + ver + "/api-docs" + apidocsExtension
			log.Println("Trying url: " + url)
			resp, err := http.Get(url)

			if err == nil && strings.Contains(resp.Status, "200") {
				passedVersion = ver
			}
			if resp != nil {
				log.Println("for version " + ver + " status code is " + resp.Status)
				resp.Body.Close()
			}
		}

		log.Println("for " + service + " version is '" + passedVersion + "'")
		if passedVersion != "" {
			cachedAvailableServices = append(cachedAvailableServices, map[string]string{
				"name": service,
				"url":  "/" + service + "/" + passedVersion + "/api-docs" + apidocsExtension,
			})
		}
	}
}

func mapValues(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func refreshCache(services []string) {
	log.Println("Refresh start")
	cachedAvailableServices = cachedAvailableServices[:0]
	for _, service := range services {
		checkService(service)
	}
	log.Println("Refresh finish")
}
