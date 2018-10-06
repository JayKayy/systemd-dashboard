package main

import (
	"fmt"
	"github.com/coreos/go-systemd/dbus"
	"html/template"
	"net/http"
	"strings"
)

func drawTable(w http.ResponseWriter, data []map[string]interface{}) {

	t := template.Must(template.ParseFiles("tmpl/services_table.html"))
	err := t.Execute(w, &data)
	if err != nil {
		panic(err)
	}
}

func getAllServices() []string {
	conn, err := dbus.New()
	defer conn.Close()
	units, err := conn.ListUnitFilesByPatterns([]string{}, []string{"*.service"})
	var serviceNames []string = make([]string, len(units), cap(units))

	for idx := range units {
		noSlash := strings.Split(units[idx].Path, "/")
		service := noSlash[len(noSlash)-1]
		removeServiceTag := strings.Split(service, ".")[0]
		serviceNames[idx] = removeServiceTag
	}
	if err != nil {
		fmt.Println(err.Error())
	}
	return serviceNames
}

func getActiveState(service string) string {
	conn, _ := dbus.New()
	defer conn.Close()
	property, err := conn.GetUnitProperty(service+".service", "ActiveState")
	if err != nil {
		panic(err)
	}
	return property.Value.String()
}

func getUnitFileState(service string) string {
	conn, _ := dbus.New()
	defer conn.Close()
	property, err := conn.GetUnitProperty(service+".service", "UnitFileState")
	if err != nil {
		panic(err)
	}
	return property.Value.String()
}

func getService(service string) map[string]interface{} {
	service = strings.TrimPrefix(service, "/") + ".service"
	conn, err := dbus.New()
	defer conn.Close()
	unit, err := conn.GetUnitProperties(service)

	if err != nil {
		fmt.Println(err.Error())
	}
	return unit
}
func query(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Path
	var serviceData []map[string]interface{}
	if query == "/" {
		results := getAllServices()
		//logging
		serviceData = make([]map[string]interface{}, len(results), cap(results))
		for idx := 0; idx < len(results); idx++ {
			serviceData[idx] = getService(results[idx])
		}
	} else {
		serviceData = make([]map[string]interface{}, 1, 1)
		serviceData[0] = getService(query)
		//logging

	}
	drawTable(w, serviceData)
}

func main() {
	http.HandleFunc("/", query)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
	fmt.Println("Service Dashboard")
}
