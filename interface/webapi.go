package delivery

import (
	"github.com/kvwmap-backup/usecases"
	"github.com/kvwmap-backup/models"
	"github.com/kvwmap-backup/configuration"
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"errors"
	"strconv"
)

func URIHandler() {
	r := mux.NewRouter()
	r.HandleFunc("/", list_backup_configs)
        r.HandleFunc("/backup-config", list_backup_configs)
        r.HandleFunc("/backup-config/{File}", show_backup_config)
	//tar
        r.HandleFunc("/backup-config/{File}/tar/new", tar_new)
	r.HandleFunc("/backup-config/{File}/tar/save", tar_save)
	r.HandleFunc("/backup-config/{File}/tar/edit/{Id}", tar_edit)
	r.HandleFunc("/backup-config/{File}/tar/delete/{Id}", tar_delete)
	//mysql
	r.HandleFunc("/backup-config/{File}/mysql/new", mysql_new)
	r.HandleFunc("/backup-config/{File}/mysql/save", mysql_save)
	r.HandleFunc("/backup-config/{File}/mysql/edit/{Id}", mysql_edit)
	r.HandleFunc("/backup-config/{File}/mysql/delete/{Id}", mysql_delete)

	//statische Inhalte, CSS+JS
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./interface/templates/static"))))

	r.NotFoundHandler = http.HandlerFunc(httpNotFound)

	listenport := config.GetConfigValFor(config.KeyHTTPPort)
	log.Printf("starting server on port %s", listenport)
        err := http.ListenAndServe(":" + listenport, r)
	if err != nil {
		log.Fatal(err)
	}
}

// /backup-config
func list_backup_configs(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("./interface/templates/configs.html")
	if err != nil {
		httpError(w, r, err)
	} else {
		templ.Execute(w, api_usecases.List_backup_configs())
	}
}

// /backup-config/{name}
func show_backup_config(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templ, err := template.ParseFiles("./interface/templates/config.html")
	if err != nil {
		httpError(w, r, err)
	}

	var c structs.HTMLTemplateData
	c.Vars = vars
	c.Backup, err = api_usecases.Load_backup_config(vars["File"])
	if err != nil {
		httpError(w, r, err)
	} else {
		templ.Execute(w, c)
	}
}

// /backup-config/{name}/tar/new
func tar_new(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var htmldata structs.HTMLTemplateData
	htmldata.Vars = vars
	templ, err := template.ParseFiles("./interface/templates/taritem.html")
	if err != nil {
		httpError(w, r, err)
	}
	templ.Execute(w, htmldata)
}

func item_save(w http.ResponseWriter, r *http.Request, t string){
	vars := mux.Vars(r)

	var request = structs.Request{}
	switch t {
		case "tar":
			item := structs.TarItem{}
			item.Source = r.PostFormValue("source")
			item.TargetName = r.PostFormValue("target")
			item.Exclude = r.PostFormValue("exclude")
			request = api_usecases.SaveItem(item , vars["File"], r.PostFormValue("Id"))
		case "mysql":
			item := structs.MysqlDumpItem{}
			item.ContainerId = r.PostFormValue("containerid")
			item.DbName = r.PostFormValue("dbname")
			item.DockerNetwork = r.PostFormValue("dockernetwork")
			item.TargetName = r.PostFormValue("targetname")
			request = api_usecases.SaveItem(item , vars["File"], r.PostFormValue("Id"))
	}

	if request.Success {
		httpSuccess(w,r, vars["File"])
	} else {
		if len(request.Errors) < 1 {
			httpErrorStr(w, r, "unbekannter Fehler")
		} else {
			httpErrorStr(w, r, request.Errors[0])
		}
	}
}

// /backup-config/{File}/tar/save
func tar_save(w http.ResponseWriter, r *http.Request) {
	item_save(w, r, "tar")
}

func item_edit(w http.ResponseWriter, r *http.Request, t string){
	vars := mux.Vars(r)

	var htmlData structs.HTMLTemplateData
	htmlData.Vars = vars
	BackupData, err := api_usecases.Load_backup_config(vars["File"])
	htmlData.Backup = BackupData
	if err != nil {
		httpErrorStr(w, r, "Laden der Backup-Config fehlgeschlagen! " + err.Error() )
	}

	id_int, err := strconv.Atoi(vars["Id"])
	if err != nil {
		httpError(w, r, err)
	}

	htmlData.Vars = vars
	var templateFile string
	switch t {
		case "tar":
			templateFile = "./interface/templates/taritem.html"
			if len(htmlData.Backup.Tar) >= id_int {
				htmlData.Vars["source"]  = htmlData.Backup.Tar[id_int].Source
				htmlData.Vars["target"]  = htmlData.Backup.Tar[id_int].TargetName
				htmlData.Vars["exclude"] = htmlData.Backup.Tar[id_int].Exclude
			} else {
				httpErrorStr(w, r, "Ausgewähltes Element nicht vorhaden!" )
			}
		case "mysql":
			templateFile = "./interface/templates/mysqlitem.html"
			if len(htmlData.Backup.MysqlDump) >= id_int {
				htmlData.Vars["ContainerId"] = htmlData.Backup.MysqlDump[id_int].ContainerId
				htmlData.Vars["DockerNetwork"] = htmlData.Backup.MysqlDump[id_int].DockerNetwork
				htmlData.Vars["DbName"] = htmlData.Backup.MysqlDump[id_int].DbName
				htmlData.Vars["TargetName"] = htmlData.Backup.MysqlDump[id_int].TargetName
			} else {
				httpErrorStr(w, r, "Ausgewähltes Element nicht vorhanden" )
			}
	}

        templ, err := template.ParseFiles(templateFile)
        if err != nil {
                httpError(w, r, err)
        }
	templ.Execute(w, htmlData)
}

func item_delete(w http.ResponseWriter, r *http.Request, item interface{}){
	vars := mux.Vars(r)
	request := api_usecases.DeleteItem(vars["File"], vars["Id"], item)
	if request.Success {
		httpSuccess(w, r, vars["File"])
	} else {
		httpErrorStr(w, r, request.Errors[0])
	}
}

// /backup-config/{File}/tar/edit/{Id}
func tar_edit(w http.ResponseWriter, r *http.Request) {
	item_edit(w, r, "tar")
}

// /backup-config/{File}/tar/delete/{Id}
func tar_delete(w http.ResponseWriter, r *http.Request) {
	item := structs.TarItem{}
	item_delete(w, r, item)
}

func httpSuccess(w http.ResponseWriter, r *http.Request, forwardto string) {
	log.Printf("httpSuccess, forwardto=%s", forwardto)
	vars := mux.Vars(r)
	templ, err := template.ParseFiles("./interface/templates/status/success.html")
	if err != nil {
		httpError(w, r, err)
	}
	vars["forwardto"] = forwardto
	templ.Execute(w, vars)
}

func httpNotFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("httpNotFound requested URI %s", r.RequestURI)
	http.ServeFile(w, r, "./interface/templates/status/notfound.html")
}

func httpErrorPanic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./interface/templates/status/panic.html")
}

func httpErrorStr(w http.ResponseWriter, r *http.Request, e string) {
	log.Printf("httpErrorStr error:%s", e)
	err := errors.New(e)
	httpError(w,r,err)
}

func httpError(w http.ResponseWriter, r *http.Request, e error) {
	log.Printf("httpError error:%s", e.Error())
        templ, err := template.ParseFiles("./interface/templates/status/error.html")
	var error_map map[string]string
	error_map = make(map[string]string)
	error_map["Error"] = e.Error()
        if err != nil {
		log.Println("Fehler beim parsen: "+err.Error())
		httpErrorPanic(w, r)
        } else {
	        templ.Execute(w, error_map)
	}
}

// /backup-config/{name}/mysql/new
func mysql_new(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var htmldata structs.HTMLTemplateData
	htmldata.Vars = vars
	templ, err := template.ParseFiles("./interface/templates/mysqlitem.html")
	if err != nil {
		httpError(w, r, err)
	}
	templ.Execute(w, htmldata)
}

// /backup-config/{File}/mysql/save
func mysql_save(w http.ResponseWriter, r *http.Request) {
	item_save(w, r, "mysql")
}

// /backup-config/{File}/mysql/edit
func mysql_edit(w http.ResponseWriter, r *http.Request) {
	item_edit(w, r, "mysql")
}

// /backup-config/{File}/mysql/delete/{Id}
func mysql_delete(w http.ResponseWriter, r *http.Request) {
	item := structs.MysqlDumpItem{}
	item_delete(w, r, item)
}
