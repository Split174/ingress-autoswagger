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

func main() {
	refreshCron, exists := os.LookupEnv("REFRESH_CRON")
	if !exists {
		refreshCron = "@every 1m"
	}

	servicesEnv := os.Getenv("SERVICES")
	openApiUrlsEnv, openApiUrlsEnvExists := os.LookupEnv("OPENAPI_URLS")
	if (servicesEnv == "" && openApiUrlsEnv == "") {
		log.Println("Environment variable \"SERVICES\" and OPENAPI_URLS is empty")
		os.Exit(2)
	}

	services := make([]string, 0)
	if (servicesEnv != "") {
		services = mapValues(strings.Split(servicesEnv[1:len(servicesEnv)-1], ","), func(s string) string {
			return s[1 : len(s)-1]
		})
		sort.Strings(services)
	}

	//set versions
	versionsEnv, versionsEnvExists := os.LookupEnv("VERSIONS")
	apidocsExtensionEnv, apidocsExtensionEnvExists := os.LookupEnv("APIDOCS_EXTENSION")

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

	var openApiUrls = make([]string, 0)
	if openApiUrlsEnvExists {
		openApiUrls = mapValues(strings.Split(openApiUrlsEnv[1:len(openApiUrlsEnv)-1], ","), func(s string) string {
			return s[1 : len(s)-1]
		})
		log.Println("OpenApi Paths URI:", openApiUrls)
	}

	log.Println("Server started on 3000 port!")
	log.Println("Services:", services)
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
		if openApiUrlsEnvExists {
			refreshCacheOpenApiUrls(openApiUrls)
		} else {
			refreshCache(services)
		}
		w.WriteHeader(200)
	})

	if openApiUrlsEnvExists {
		refreshCacheOpenApiUrls(openApiUrls)
	} else {
		refreshCache(services)
	}

	c := cron.New()
	c.AddFunc(refreshCron, func() {
		log.Println("Cron init")
		if openApiUrlsEnvExists {
			refreshCacheOpenApiUrls(openApiUrls)
		} else {
			refreshCache(services)
		}
		log.Println("Cron has been finished")
	})
	c.Start()

	_ = http.ListenAndServe(":3000", nil)
}

func checkService(service string) {
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

func mapValues(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func refreshCache(services []string) {
	log.Println("Refresh start")
	log.Println("WARNING: The SERVICES, VERSION and APIDOCS_EXTENSION environment variables are deprecated, ", 
			"it is recommended to use the OPENAPI_URLS variable instead")
	cachedAvailableServices = cachedAvailableServices[:0]
	for _, service := range services {
		checkService(service)
	}
	log.Println("Refresh finish")
}

func refreshCacheOpenApiUrls(openApiUrls []string) {
	log.Println("Refresh start")
	cachedAvailableServices = cachedAvailableServices[:0]
	for _, openApiUrl := range openApiUrls {
		log.Println("Trying url: " + openApiUrl)
		resp, err := http.Get(openApiUrl)

		if resp != nil {
			log.Println("schema for url " + openApiUrl + " status code is " + resp.Status)
			resp.Body.Close()
		}

		var domain = strings.Replace(openApiUrl, "https://", "", -1)
		domain = strings.Replace(domain, "http://", "", -1)
		
		var chacheUrl = ""
		if strings.Contains(domain, ".") {
			chacheUrl = openApiUrl
		} else {
			chacheUrl = "/" + domain
		}

		if err == nil && strings.Contains(resp.Status, "200") {
			cachedAvailableServices = append(cachedAvailableServices, map[string]string{
				"name": domain,
				"url": chacheUrl,
			})
		}
	}
	log.Println("Refresh finish")
}
